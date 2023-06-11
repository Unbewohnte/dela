package server

import (
	"Unbewohnte/dela/db"
	"Unbewohnte/dela/encryption"
	"net/http"
	"strconv"
	"strings"
)

const (
	RequestHeaderSecurityKey string = "Security-Key"
	// RequestHeaderAuthSeparator string = "\u200b" // username\u200bpassword
	RequestHeaderAuthSeparator string = "<-->" // username<-->password
	RequestHeaderAuthKey       string = "Auth"
	RequestHeaderTodoIDKey     string = "Todo-Key"
	RequestHeaderEncodedB64    string = "EncryptedBase64" // tells whether auth data is encoded in base64
)

// Checks if the request header contains a valid full access key string or not
func DoesRequestHasFullAccess(req *http.Request, accessKey string) bool {
	var headerAccessKey string
	if req.Header.Get(RequestHeaderEncodedB64) == "true" {
		headerAccessKey = encryption.DecodeString(req.Header.Get(RequestHeaderSecurityKey))
	} else {
		headerAccessKey = req.Header.Get(RequestHeaderSecurityKey)
	}

	if headerAccessKey == "" || headerAccessKey != accessKey {
		return false
	}

	return true
}

// Gets auth data from the request and rips the login string from it. Returns ""
// if there is no auth data at all
func GetUsernameFromAuth(req *http.Request) string {
	var authInfoStr string
	if req.Header.Get(RequestHeaderEncodedB64) == "true" {
		authInfoStr = encryption.DecodeString(req.Header.Get(RequestHeaderAuthKey))
	} else {
		authInfoStr = req.Header.Get(RequestHeaderAuthKey)
	}

	authInfoSplit := strings.Split(authInfoStr, RequestHeaderAuthSeparator)
	if len(authInfoSplit) != 2 {
		// no separator or funny username|password
		return ""
	}
	username := authInfoSplit[0]

	return username
}

// Verifies if the request contains a valid user auth information (username-password pair)
func IsRequestAuthValid(req *http.Request, db *db.DB) bool {
	var authInfoStr string
	if req.Header.Get(RequestHeaderEncodedB64) == "true" {
		authInfoStr = encryption.DecodeString(req.Header.Get(RequestHeaderAuthKey))
	} else {
		authInfoStr = req.Header.Get(RequestHeaderAuthKey)
	}
	authInfoSplit := strings.Split(authInfoStr, RequestHeaderAuthSeparator)
	if len(authInfoSplit) != 2 {
		// no separator or funny id|password
		return false
	}

	username, password := authInfoSplit[0], authInfoSplit[1]
	user, err := db.GetUser(username)
	if err != nil {
		// does not exist
		return false
	}

	if password != user.Password {
		// password does not match
		return false
	}

	return true
}

// Checks if given user owns a todo
func DoesUserOwnTodo(username string, todoID uint64, db *db.DB) bool {
	todo, err := db.GetTodo(todoID)
	if err != nil {
		return false
	}

	if todo.OwnerUsername != username {
		return false
	}

	return true
}

// Checks if given user owns a todo group
func DoesUserOwnTodoGroup(username string, todoGroupID uint64, db *db.DB) bool {
	group, err := db.GetTodoGroup(todoGroupID)
	if err != nil {
		return false
	}

	if group.OwnerUsername != username {
		return false
	}

	return true
}

// Retrieves todo ID from request headers
func GetTodoIDFromReq(req *http.Request) (uint64, error) {
	todoIDStr := encryption.DecodeString(req.Header.Get(RequestHeaderTodoIDKey))
	todoID, err := strconv.ParseUint(todoIDStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return todoID, nil
}
