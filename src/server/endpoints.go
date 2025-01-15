/*
  	dela - web TODO list
    Copyright (C) 2023, 2024  Kasyanov Nikolay Alexeyevich (Unbewohnte)

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package server

import (
	"Unbewohnte/dela/db"
	"Unbewohnte/dela/email"
	"Unbewohnte/dela/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"
)

func (s *Server) EndpointUserCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve user data
	defer req.Body.Close()

	contents, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("[Server][EndpointUserCreate] Failed to read request body: %s", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var user db.User
	err = json.Unmarshal(contents, &user)
	if err != nil {
		logger.Error("[Server][EndpointUserCreate] Failed to unmarshal user data: %s", err)
		http.Error(w, "User JSON unmarshal error", http.StatusInternalServerError)
		return
	}

	// Sanitize
	valid, reason := IsUserValid(user)
	if !valid {
		http.Error(w, reason, http.StatusInternalServerError)
		return
	}
	user.TimeCreatedUnix = uint64(time.Now().Unix())

	// Insert into DB
	err = s.db.CreateUser(user)
	if err != nil {
		logger.Error("[Server][EndpointUserCreate] Failed to insert new user \"%s\" data: %s", user.Email, err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	logger.Info("[Server][EndpointUserCreate] Created a new user with email \"%s\"", user.Email)

	// Create a non-removable default category
	err = s.db.CreateTodoGroup(db.NewTodoGroup(
		"Notes",
		uint64(time.Now().Unix()),
		user.Email,
		false,
	))
	if err != nil {
		http.Error(w, "Failed to create default group", http.StatusInternalServerError)
		logger.Error("[Server][EndpointUserCreate] Failed to create a default group for %s: %s", user.Email, err)
		return
	}

	// Check if email verification is required
	if !s.config.Verification.VerifyEmails {
		// Do not verify email

		// Send cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth",
			Value:    fmt.Sprintf("%s:%s", user.Email, user.Password),
			SameSite: http.SameSiteStrictMode,
			HttpOnly: false,
			Path:     "/",
			Secure:   false,
		})

		// Done
		w.Write([]byte("{\"confirm_email\":false}"))

		logger.Info("[Server][EndpointUserCreate] Successfully sent email notification to %s", user.Email)
		return
	}

	// Send email verification message
	verification, err := GenerateVerificationCode(s.db, user.Email)
	if err != nil {
		logger.Error("[Server][EndpointUserCreate] Failed to generate verification code for %s: %s", user.Email, err)
		http.Error(w, "Failed to generate confirmation code", http.StatusInternalServerError)
		return
	}

	// Send verification email
	err = s.emailer.SendEmail(
		email.NewEmail(
			s.config.Verification.Emailer.User,
			"Dela: Email verification",
			fmt.Sprintf("Your email verification code: <b>%s</b>\nPlease, verify your email in %f hours.\nThis email was specified during Dela account creation. Ignore this message if it wasn't you", verification.Code, float32(verification.LifeSeconds)/3600),
			[]string{user.Email},
		),
	)
	if err != nil {
		logger.Error("[Server][EndpointUserCreate] Failed to send verification email to %s: %s", user.Email, err)
		http.Error(w, "Failed to send email verification message", http.StatusInternalServerError)
		return
	}

	// Autodelete user account if email was not verified in time
	time.AfterFunc(time.Second*time.Duration(verification.LifeSeconds), func() {
		err = s.db.DeleteUnverifiedUserClean(user.Email)
		if err != nil {
			logger.Error("[Server][EndpointUserCreate] Failed to autodelete unverified user %s: %s", user.Email, err)
		}
	})

	w.Write([]byte("{\"confirm_email\":true}"))
}

func (s *Server) EndpointUserVerify(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve data
	defer req.Body.Close()

	contents, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("[Server][EndpointUserVerify] Failed to read request body: %s", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	type verificationAnswer struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	var answer verificationAnswer
	err = json.Unmarshal(contents, &answer)
	if err != nil {
		logger.Error("[Server][EndpointUserVerify] Failed to unmarshal verification answer: %s", err)
		http.Error(w, "Verification answer JSON unmarshal error", http.StatusInternalServerError)
		return
	}

	// Retrieve user
	user, err := s.db.GetUser(answer.Email)
	if err != nil {
		logger.Error("[Server][EndpointUserVerify] Failed to retrieve information on \"%s\": %s", answer.Email, err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Compare codes
	dbCode, err := s.db.GetVerificationByEmail(user.Email)
	if err != nil {
		logger.Error("[Server][EndpointUserVerify] Could not get verification code from DB for %s: %s", user.Email, err)
		http.Error(w, "Could not retrieve verification information for this email", http.StatusInternalServerError)
		return
	}

	if answer.Code != dbCode.Code {
		// Codes do not match!
		logger.Error("[Server][EndpointUserVerify] %s sent wrong verification code", user.Email)
		http.Error(w, "Wrong verification code!", http.StatusForbidden)
		return
	}

	// All's good!
	err = s.db.UserSetEmailConfirmed(user.Email)
	if err != nil {
		http.Error(w, "Failed to save confirmation information", http.StatusInternalServerError)
		logger.Error("[Server][EndpointUserVerify] Failed to set confirmed_email to true for %s: %s", user.Email, err)
		return
	}

	logger.Info("[Server][EndpointUserVerify] %s was successfully verified!", user.Email)

	// Send cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    fmt.Sprintf("%s:%s", user.Email, user.Password),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: false,
		Path:     "/",
		Secure:   false,
	})
	w.WriteHeader(http.StatusOK)
}

func (s *Server) EndpointUserLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve user data
	defer req.Body.Close()

	contents, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("[Server][EndpointUserLogin] Failed to read request body: %s", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var user db.User
	err = json.Unmarshal(contents, &user)
	if err != nil {
		logger.Error("[Server][EndpointUserLogin] Failed to unmarshal user data: %s", err)
		http.Error(w, "User JSON unmarshal error", http.StatusInternalServerError)
		return
	}

	// Check auth data
	if !IsUserAuthorized(s.db, user) {
		http.Error(w, "Failed auth", http.StatusForbidden)
		return
	}

	// Send cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    fmt.Sprintf("%s:%s", user.Email, user.Password),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: false,
		Path:     "/",
		Secure:   false,
	})
	w.WriteHeader(http.StatusOK)
}

func (s *Server) EndpointUserUpdate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve user data
	defer req.Body.Close()

	// Authentication check
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Authentication error", http.StatusForbidden)
		return
	}

	contents, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("[Server][EndpointUserUpdate] Failed to read request body: %s", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var user db.User
	err = json.Unmarshal(contents, &user)
	if err != nil {
		logger.Error("[Server][EndpointUserUpdate] Failed to unmarshal user data: %s", err)
		http.Error(w, "User JSON unmarshal error", http.StatusInternalServerError)
		return
	}

	// Check whether the user in request is the user specified in JSON
	email := GetLoginFromReq(req)
	if email != user.Email {
		// Gotcha!
		logger.Warning("[Server][EndpointUserUpdate] %s tried to update user information of %s!", email, user.Email)
		http.Error(w, "Logins do not match", http.StatusForbidden)
		return
	}

	// Update
	err = s.db.UserUpdate(user)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		logger.Error("[Server][EndpointUserUpdate] Failed to update \"%s\": %s", user.Email, err)
		return
	}

	logger.Info("[Server][EndpointUserUpdate] Updated a user with email \"%s\"", user.Email)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) EndpointUserDelete(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()

	// Authentication check
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Authentication error", http.StatusForbidden)
		return
	}

	// Delete
	email := GetLoginFromReq(req)
	err := s.db.DeleteUser(email)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		logger.Error("[Server][EndpointUserDelete] Failed to delete \"%s\": %s", email, err)
		return
	}

	logger.Info("[Server][EndpointUserDelete] Deleted a user with email \"%s\"", email)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) EndpointUserGet(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()

	// Authentication check
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Authentication error", http.StatusForbidden)
		return
	}

	// Get information from the database
	email := GetLoginFromReq(req)
	userDB, err := s.db.GetUser(email)
	if err != nil {
		logger.Error("[Server][EndpointUserGet] Failed to retrieve information on \"%s\": %s", email, err)
		http.Error(w, "Failed to fetch information", http.StatusInternalServerError)
		return
	}

	userDBBytes, err := json.Marshal(&userDB)
	if err != nil {
		logger.Error("[Server][EndpointUserGet] Failed to marshal information on \"%s\": %s", email, err)
		http.Error(w, "Failed to marshal information", http.StatusInternalServerError)
		return
	}

	// Send
	w.Write(userDBBytes)
}

func (s *Server) EndpointTodoUpdate(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication information
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Obtain TODO ID
	todoIDStr := path.Base(req.URL.Path)
	todoID, err := strconv.ParseUint(todoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid TODO ID", http.StatusBadRequest)
		return
	}

	// Check if the user owns this TODO
	if !s.db.DoesUserOwnTodo(todoID, GetLoginFromReq(req)) {
		http.Error(w, "You don't own this TODO", http.StatusForbidden)
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
	var updatedTodo db.Todo
	err = json.Unmarshal(body, &updatedTodo)
	if err != nil {
		logger.Warning("[Server] Received invalid TODO JSON in order to update: %s", err)
		http.Error(w, "Invalid TODO JSON", http.StatusBadRequest)
		return
	}

	// Update. (Creation date, owner username and an ID do not change)
	err = s.db.UpdateTodo(todoID, updatedTodo)
	if err != nil {
		logger.Warning("[Server] Failed to update TODO: %s", err)
		http.Error(w, "Failed to update", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	logger.Info("[Server] Updated TODO with ID %d", todoID)
}

func (s *Server) EndpointTodoMarkDone(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication information
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Obtain TODO ID
	todoIDStr := path.Base(req.URL.Path)
	todoID, err := strconv.ParseUint(todoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid TODO ID", http.StatusBadRequest)
		return
	}

	// Check if the user owns this TODO
	if !s.db.DoesUserOwnTodo(todoID, GetLoginFromReq(req)) {
		http.Error(w, "You don't own this TODO", http.StatusForbidden)
		return
	}

	todo, err := s.db.GetTodo(todoID)
	if err != nil {
		http.Error(w, "Can't access this TODO", http.StatusInternalServerError)
		return
	}

	// Update
	todo.IsDone = true
	todo.CompletionTimeUnix = uint64(time.Now().Unix())
	err = s.db.UpdateTodo(todoID, *todo)
	if err != nil {
		logger.Warning("[Server] Failed to update TODO: %s", err)
		http.Error(w, "Failed to update", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	logger.Info("[Server] Marked TODO as done %d", todoID)
}

func (s *Server) EndpointTodoDelete(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delete an existing TODO

	// Check if this user actually exists and the password is valid
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Obtain TODO ID
	todoIDStr := path.Base(req.URL.Path)
	todoID, err := strconv.ParseUint(todoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid TODO ID", http.StatusBadRequest)
		return
	}

	// Check if the user owns this TODO
	if !s.db.DoesUserOwnTodo(todoID, GetLoginFromReq(req)) {
		http.Error(w, "You don't own this TODO", http.StatusForbidden)
		return
	}

	// Now delete
	err = s.db.DeleteTodo(todoID)
	if err != nil {
		logger.Error("[Server] Failed to delete %s's TODO: %s", GetLoginFromReq(req), err)
		http.Error(w, "Failed to delete TODO", http.StatusInternalServerError)
		return
	}

	// Success!
	logger.Info("[Server] Deleted TODO with ID %d", todoID)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) EndpointTodoCreate(w http.ResponseWriter, req *http.Request) {
	// Create a new TODO
	defer req.Body.Close()
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Add TODO to the database
	if newTodo.GroupID == 0 {
		http.Error(w, "No group ID was provided", http.StatusBadRequest)
		return
	}

	if !s.db.DoesUserOwnGroup(newTodo.GroupID, GetLoginFromReq(req)) {
		http.Error(w, "You do not own this group", http.StatusForbidden)
		return
	}

	newTodo.OwnerEmail = GetLoginFromReq(req)
	newTodo.TimeCreatedUnix = uint64(time.Now().Unix())
	err = s.db.CreateTodo(newTodo)
	if err != nil {
		http.Error(w, "Failed to create TODO", http.StatusInternalServerError)
		logger.Error("[Server] Failed to put a new todo (%+v) into the db: %s", newTodo, err)
		return
	}

	// Success!
	w.WriteHeader(http.StatusOK)
	logger.Info("[Server] Created a new TODO for %s", newTodo.OwnerEmail)
}

func (s *Server) EndpointUserTodosGet(w http.ResponseWriter, req *http.Request) {
	// Retrieve TODO information

	defer req.Body.Close()

	// Authentication check
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Authentication error", http.StatusForbidden)
		return
	}

	// Check authentication information
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Get all user TODOs
	todos, err := s.db.GetAllUserTodos(GetLoginFromReq(req))
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
}

func (s *Server) EndpointTodoGroupDelete(w http.ResponseWriter, req *http.Request) {
	// Delete an existing group
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if given user actually owns this group
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Get group ID
	groupId, err := strconv.ParseUint(path.Base(req.URL.Path), 10, 64)
	if err != nil {
		http.Error(w, "Bad Category ID", http.StatusBadRequest)
		return
	}

	if !s.db.DoesUserOwnGroup(groupId, GetLoginFromReq(req)) {
		http.Error(w, "You don't own this group", http.StatusForbidden)
		return
	}

	groupDB, err := s.db.GetTodoGroup(groupId)
	if err != nil {
		logger.Error("[Server][EndpointGroupDelete] Failed to fetch TODO group with Id %d: %s", groupId, err)
		http.Error(w, "Failed to retrieve TODO group", http.StatusInternalServerError)
		return
	}

	if !groupDB.Removable {
		// Not removable
		http.Error(w, "Not removable", http.StatusBadRequest)
		return
	}

	// Delete all ToDos associated with this group and then delete the group itself
	err = s.db.DeleteTodoGroupClean(groupId)
	if err != nil {
		logger.Error("[Server][EndpointGroupDelete] Failed to delete %s's TODO group: %s", GetLoginFromReq(req), err)
		http.Error(w, "Failed to delete TODO group", http.StatusInternalServerError)
		return
	}

	// Success!
	logger.Info("[Server][EndpointGroupDelete] Cleanly deleted group ID: %d for %s", groupId, GetLoginFromReq(req))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) EndpointTodoGroupCreate(w http.ResponseWriter, req *http.Request) {
	// Create a new TODO group
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Add group to the database
	newGroup.OwnerEmail = GetLoginFromReq(req)
	newGroup.TimeCreatedUnix = uint64(time.Now().Unix())
	newGroup.Removable = true
	err = s.db.CreateTodoGroup(newGroup)
	if err != nil {
		http.Error(w, "Failed to create TODO group", http.StatusInternalServerError)
		return
	}

	// Success!
	w.WriteHeader(http.StatusOK)
	logger.Info("[Server] Created a new TODO group for %s", newGroup.OwnerEmail)
}

func (s *Server) EndpointTodoGroupGet(w http.ResponseWriter, req *http.Request) {
	// Retrieve all todo groups
	defer req.Body.Close()

	// Check authentication information
	if !IsUserAuthorizedReq(req, s.db) {
		http.Error(w, "Invalid user auth data", http.StatusForbidden)
		return
	}

	// Get groups
	groups, err := s.db.GetAllUserTodoGroups(GetLoginFromReq(req))
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
}

func (s *Server) EndpointTodoGroupUpdate(w http.ResponseWriter, req *http.Request) {
	// Check authentication information
	if !IsUserAuthorizedReq(req, s.db) {
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
}
