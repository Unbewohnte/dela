package server

import (
	"Unbewohnte/dela/db"
	"Unbewohnte/dela/logger"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func (s *Server) UserEndpoint(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		// Delete an existing user
		defer req.Body.Close()

		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to delete a user: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var newUser db.User
		err = json.Unmarshal(body, &newUser)
		if err != nil {
			logger.Warning("[Server] Received invalid user JSON for deletion: %s", err)
			http.Error(w, "Invalid user JSON", http.StatusBadRequest)
			return
		}

		username := GetUsernameFromAuth(req)

		// Check if auth data is valid
		if !IsRequestAuthValid(req, s.db) {
			logger.Warning("[Server] %s failed to authenticate as %s", req.RemoteAddr, username)
			http.Error(w, "Invalid user auth data", http.StatusBadRequest)
			return
		}

		// It is, indeed, a user
		// Delete with all TODOs
		err = s.db.DeleteUserClean(username)
		if err != nil {
			logger.Error("[Server] Failed to delete %s: %s", username, err)
			http.Error(w, "Failed to delete user or TODO contents", http.StatusInternalServerError)
			return
		}

		// Success!
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		// Create a new user
		defer req.Body.Close()
		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to create a new user: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var newUser db.User
		err = json.Unmarshal(body, &newUser)
		if err != nil {
			logger.Warning("[Server] Received invalid user JSON for creation: %s", err)
			http.Error(w, "Invalid user JSON", http.StatusBadRequest)
			return
		}

		// Check for validity
		valid, reason := IsUserValid(newUser)
		if !valid {
			logger.Info("[Server] Rejected creating %s for reason: %s", newUser.Username, reason)
			http.Error(w, "Invalid user data: "+reason, http.StatusBadRequest)
			return
		}

		// Add user to the database
		newUser.TimeCreatedUnix = uint64(time.Now().Unix())
		err = s.db.CreateUser(newUser)
		if err != nil {
			http.Error(w, "User already exists", http.StatusInternalServerError)
			return
		}

		// Create an initial TODO group
		err = s.db.CreateTodoGroup(
			db.TodoGroup{
				Name:            "Todos",
				TimeCreatedUnix: uint64(time.Now().Unix()),
				OwnerUsername:   newUser.Username,
			},
		)
		if err != nil {
			// Oops, that's VERY bad. Delete newly created user
			s.db.DeleteUser(newUser.Username)
			logger.Error("[SERVER] Failed to create an initial TODO group for a newly created \"%s\": %s. Deleted.", newUser.Username, err)
			http.Error(w, "Failed to create initial TODO group", http.StatusInternalServerError)
			return
		}

		// Success!
		w.WriteHeader(http.StatusOK)
		logger.Info("[Server] Created a new user \"%s\"", newUser.Username)
	case http.MethodGet:
		// Check if user information is valid
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) TodoEndpoint(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		// Delete an existing TODO
		defer req.Body.Close()

		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to possibly delete a TODO: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var todo db.Todo
		err = json.Unmarshal(body, &todo)
		if err != nil {
			logger.Warning("[Server] Received invalid TODO JSON for deletion: %s", err)
			http.Error(w, "Invalid TODO JSON", http.StatusBadRequest)
			return
		}

		// Check if given user actually owns this TODO
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		if !DoesUserOwnTodo(GetUsernameFromAuth(req), todo.ID, s.db) {
			http.Error(w, "You don't own this TODO", http.StatusForbidden)
			return
		}

		// Now delete
		err = s.db.DeleteTodo(todo.ID)
		if err != nil {
			logger.Error("[Server] Failed to delete %s's TODO: %s", GetUsernameFromAuth(req), err)
			http.Error(w, "Failed to delete TODO", http.StatusInternalServerError)
			return
		}

		// Success!
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		// Create a new TODO
		defer req.Body.Close()
		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to create a new TODO: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var newTodo db.Todo
		err = json.Unmarshal(body, &newTodo)
		if err != nil {
			logger.Warning("[Server] Received invalid TODO JSON for creation: %s", err)
			http.Error(w, "Invalid TODO JSON", http.StatusBadRequest)
			return
		}

		// Check for authentication problems
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		// Add TODO to the database
		newTodo.OwnerUsername = GetUsernameFromAuth(req)
		newTodo.TimeCreatedUnix = uint64(time.Now().Unix())
		err = s.db.CreateTodo(newTodo)
		if err != nil {
			http.Error(w, "Failed to create TODO", http.StatusInternalServerError)
			return
		}

		// Success!
		w.WriteHeader(http.StatusOK)
		logger.Info("[Server] Created a new TODO for %s", newTodo.OwnerUsername)
	case http.MethodGet:
		// Retrieve TODO information
		// Check authentication information
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		// Get TODO
		todos, err := s.db.GetAllUserTodos(GetUsernameFromAuth(req))
		if err != nil {
			http.Error(w, "Failed to get TODOs", http.StatusInternalServerError)
			return
		}

		// Marshal to JSON
		todosBytes, err := json.Marshal(&todos)
		if err != nil {
			http.Error(w, "Failed to marhsal TODOs JSON", http.StatusInternalServerError)
			return
		}

		// Send out
		w.Header().Add("Content-Type", "application/json")
		w.Write(todosBytes)

	case http.MethodPatch:
		// Change TODO due date and text

		// Check authentication information
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to possibly update a TODO: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var todo db.Todo
		err = json.Unmarshal(body, &todo)
		if err != nil {
			logger.Warning("[Server] Received invalid TODO JSON in order to update: %s", err)
			http.Error(w, "Invalid TODO JSON", http.StatusBadRequest)
			return
		}

		// TODO
		err = s.db.UpdateTodo(todo.ID, todo)
		if err != nil {
			logger.Warning("[Server] Failed to update TODO: %s", err)
			http.Error(w, "Failed to update", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) TodoGroupEndpoint(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		// Delete an existing group
		defer req.Body.Close()

		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to possibly delete a TODO group: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var group db.TodoGroup
		err = json.Unmarshal(body, &group)
		if err != nil {
			logger.Warning("[Server] Received invalid TODO group JSON for deletion: %s", err)
			http.Error(w, "Invalid TODO group JSON", http.StatusBadRequest)
			return
		}

		// Check if given user actually owns this group
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		if !DoesUserOwnTodoGroup(GetUsernameFromAuth(req), group.ID, s.db) {
			http.Error(w, "You don't own this group", http.StatusForbidden)
			return
		}

		// Now delete
		err = s.db.DeleteTodoGroup(group.ID)
		if err != nil {
			logger.Error("[Server] Failed to delete %s's TODO group: %s", GetUsernameFromAuth(req), err)
			http.Error(w, "Failed to delete TODO group", http.StatusInternalServerError)
			return
		}

		// Success!
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		// Create a new TODO group
		defer req.Body.Close()
		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to create a new TODO group: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var newGroup db.TodoGroup
		err = json.Unmarshal(body, &newGroup)
		if err != nil {
			logger.Warning("[Server] Received invalid TODO group JSON for creation: %s", err)
			http.Error(w, "Invalid TODO group JSON", http.StatusBadRequest)
			return
		}

		// Check for authentication problems
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		// Add group to the database
		newGroup.OwnerUsername = GetUsernameFromAuth(req)
		newGroup.TimeCreatedUnix = uint64(time.Now().Unix())
		err = s.db.CreateTodoGroup(newGroup)
		if err != nil {
			http.Error(w, "Failed to create TODO group", http.StatusInternalServerError)
			return
		}

		// Success!
		w.WriteHeader(http.StatusOK)
		logger.Info("[Server] Created a new TODO group for %s", newGroup.OwnerUsername)
	case http.MethodGet:
		// Retrieve all todo groups

		// Check authentication information
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		// Get groups
		groups, err := s.db.GetAllUserTodoGroups(GetUsernameFromAuth(req))
		if err != nil {
			http.Error(w, "Failed to get TODO groups", http.StatusInternalServerError)
			return
		}

		// Marshal to JSON
		groupBytes, err := json.Marshal(&groups)
		if err != nil {
			http.Error(w, "Failed to marhsal TODO groups JSON", http.StatusInternalServerError)
			return
		}

		// Send out
		w.Header().Add("Content-Type", "application/json")
		w.Write(groupBytes)

	case http.MethodPatch:
		// Check authentication information
		if !IsRequestAuthValid(req, s.db) {
			http.Error(w, "Invalid user auth data", http.StatusForbidden)
			return
		}

		// Read body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Warning("[Server] Failed to read request body to possibly update a TODO group: %s", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON
		var group db.TodoGroup
		err = json.Unmarshal(body, &group)
		if err != nil {
			logger.Warning("[Server] Received invalid TODO group JSON in order to update: %s", err)
			http.Error(w, "Invalid group JSON", http.StatusBadRequest)
			return
		}

		// TODO
		err = s.db.UpdateTodoGroup(group.ID, group)
		if err != nil {
			logger.Warning("[Server] Failed to update TODO group: %s", err)
			http.Error(w, "Failed to update", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
