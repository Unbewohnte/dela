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
	verification, err := GenerateVerificationCode(s.db, user.Email, 5, uint64(time.Hour.Seconds()))
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
			fmt.Sprintf("<p>Your email verification code is: <b>%s</b></p><p>Please, verify your email in %.1f hours. Your account will be deleted after some time without verified status.</p><p>This email was specified during Dela account creation. Ignore this message if it wasn't you.</p>", verification.Code, float32(verification.LifeSeconds)/3600),
			[]string{user.Email},
		),
	)
	if err != nil {
		logger.Error("[Server][EndpointUserCreate] Failed to send verification email to %s: %s", user.Email, err)
		http.Error(w, "Failed to send email verification message", http.StatusInternalServerError)
		return
	}
	logger.Info("[Server][EndpointUserCreate] Successfully sent confirmation email to %s", user.Email)

	// Autodelete user account after some more time if email was not verified in time
	time.AfterFunc((time.Second*time.Duration(verification.LifeSeconds))*5, func() {
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
		// Most likely already deleted this user's account
		http.Error(w, "Account no longer exists, try registering again", http.StatusInternalServerError)
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

	// Check for lifetime
	if time.Now().Unix() > int64(dbCode.IssuedUnix+dbCode.LifeSeconds) {
		// Expired!
		http.Error(w, "This code is expired!", http.StatusForbidden)
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

func (s *Server) EndpointUserNotify(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type notifyRequest struct {
		Notify bool `json:"notify"`
	}

	// Retrieve data
	defer req.Body.Close()

	contents, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Error("[Server][EndpointUserNotify] Failed to read request body: %s", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var notifyResult notifyRequest
	err = json.Unmarshal(contents, &notifyResult)
	if err != nil {
		logger.Error("[Server][EndpointUserVerify] Failed to unmarshal notification value change: %s", err)
		http.Error(w, "Bad JSON", http.StatusInternalServerError)
		return
	}

	userEmail := GetEmailFromReq(req)
	err = s.db.UserSetNotifyOnTodos(userEmail, notifyResult.Notify)
	if err != nil {
		logger.Error("[Server][EndpointUserNotify] Failed to UserSetNotifyOnTodos for %s: %s", userEmail, err)
		http.Error(w, "Failed to change user settings", http.StatusInternalServerError)
		return
	}

	if notifyResult.Notify {
		logger.Info("[Server][EndpointUserNotify] Notifying %s for due TODOs", userEmail)
	} else {
		logger.Info("[Server][EndpointUserNotify] Stopped notifying %s for due TODOs", userEmail)
	}

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
	email := GetEmailFromReq(req)
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
	email := GetEmailFromReq(req)
	err := s.db.DeleteUserClean(email)
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
	email := GetEmailFromReq(req)
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

func (s *Server) EndpointTodoFile(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != http.MethodPost && req.Method != http.MethodGet {
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
	if !s.db.DoesUserOwnTodo(todoID, GetEmailFromReq(req)) {
		http.Error(w, "You don't own this TODO", http.StatusForbidden)
		return
	}

	todo, err := s.db.GetTodo(todoID)
	if err != nil {
		http.Error(w, "Failed to retrieve this TODO", http.StatusInternalServerError)
		logger.Error("[Server][EndpointTodoFile] Failed to get TODO with ID %d: %s", todoID, err)
		return
	}

	switch req.Method {
	case http.MethodGet:
		// Retrieve file and send it
		_, err := w.Write(todo.File)
		if err != nil {
			http.Error(w, "Failed to send file", http.StatusInternalServerError)
			logger.Error("[Server][EndpointTodoFile] Failed to send TODO's file with ID %d: %s", todoID, err)
			return
		}

	case http.MethodPost:
		// Retrieve file and update database
		// Parse form
		err := req.ParseMultipartForm(int64(MaxTodoFileSizeBytes))
		if err != nil {
			logger.Error("[Server][EndpointTodoFile] Failed to parse multipart form: %s", err)
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		formFile, fileHeader, err := req.FormFile("file")
		if err != nil {
			logger.Error("[Server][EndpointTodoFile] Failed to retrieve file file from form: %s", err)
			http.Error(w, "Failed to retrieve file", http.StatusInternalServerError)
			return
		}
		defer formFile.Close()

		// Check if thumbnail is good to go
		if fileHeader.Size > int64(MaxTodoFileSizeBytes) {
			logger.Error("[Server][EndpointTodoFile] File file is too big (%d)", fileHeader.Size)
			http.Error(w, "Attachment File is too big", http.StatusBadRequest)
			return
		}

		// Save attachment to database
		fileData, err := io.ReadAll(formFile)
		if err != nil {
			logger.Error("[Server][EndpointTodoFile] Failed to read file from form: %s", err)
			http.Error(w, "Failed to read Attachment File", http.StatusInternalServerError)
			return
		}

		err = s.db.UpdateTodoFile(todoID, fileData)
		if err != nil {
			logger.Error("[Server][EndpointTodoFile] Failed to save attachment file: %s", err)
			http.Error(w, "Failed to save Attachment File", http.StatusInternalServerError)
			return
		}

		logger.Info("[Server][EndpointTodoFile] Successfully saved \"%s\" (%vMB) for %s (todoID: %d)",
			fileHeader.Filename,
			float32(fileHeader.Size)/1024.0/1024.0,
			GetEmailFromReq(req),
			todoID,
		)
	}
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
	if !s.db.DoesUserOwnTodo(todoID, GetEmailFromReq(req)) {
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

	// Validate
	if uint(len([]rune(updatedTodo.Text))) > MaxTodoTextLength {
		http.Error(
			w,
			fmt.Sprintf("Text is too big! Text must be less than %d characters long!", MaxTodoTextLength),
			http.StatusBadRequest,
		)
		return
	}
	updatedTodo.File = nil
	updatedTodo.ID = todoID

	// Update
	err = s.db.UpdateTodoSoft(todoID, updatedTodo)
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
	if !s.db.DoesUserOwnTodo(todoID, GetEmailFromReq(req)) {
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
	if !s.db.DoesUserOwnTodo(todoID, GetEmailFromReq(req)) {
		http.Error(w, "You don't own this TODO", http.StatusForbidden)
		return
	}

	// Now delete
	err = s.db.DeleteTodo(todoID)
	if err != nil {
		logger.Error("[Server] Failed to delete %s's TODO: %s", GetEmailFromReq(req), err)
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

	// Check if text is too long or not
	if uint(len([]rune(newTodo.Text))) > MaxTodoTextLength {
		http.Error(
			w,
			fmt.Sprintf("Text is too big! Text must be less than %d characters long!", MaxTodoTextLength),
			http.StatusBadRequest,
		)
		return
	}

	// Add TODO to the database
	if newTodo.GroupID == 0 {
		http.Error(w, "No group ID was provided", http.StatusBadRequest)
		return
	}

	if !s.db.DoesUserOwnGroup(newTodo.GroupID, GetEmailFromReq(req)) {
		http.Error(w, "You do not own this group", http.StatusForbidden)
		return
	}

	newTodo.OwnerEmail = GetEmailFromReq(req)
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
	todos, err := s.db.GetAllUserTodos(GetEmailFromReq(req))
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

	if !s.db.DoesUserOwnGroup(groupId, GetEmailFromReq(req)) {
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
		logger.Error("[Server][EndpointGroupDelete] Failed to delete %s's TODO group: %s", GetEmailFromReq(req), err)
		http.Error(w, "Failed to delete TODO group", http.StatusInternalServerError)
		return
	}

	// Success!
	logger.Info("[Server][EndpointGroupDelete] Cleanly deleted group ID: %d for %s", groupId, GetEmailFromReq(req))
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
	newGroup.OwnerEmail = GetEmailFromReq(req)
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
	groups, err := s.db.GetAllUserTodoGroups(GetEmailFromReq(req))
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
