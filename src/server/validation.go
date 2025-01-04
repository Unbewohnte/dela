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
	"fmt"
	"net/http"
	"strings"
)

const (
	MinimalEmailLength    uint = 3
	MinimalPasswordLength uint = 5
	MaxEmailLength        uint = 60
	MaxPasswordLength     uint = 250
	MaxTodoLength         uint = 150
)

// Check if user is valid. Returns false and a reason-string if not
func IsUserValid(user db.User) (bool, string) {
	if uint(len(user.Email)) < MinimalEmailLength {
		return false, "Email is too small"
	}
	if uint(len(user.Email)) > MaxEmailLength {
		return false, fmt.Sprintf("Email is too big; Email should be up to %d characters", MaxEmailLength)
	}

	if uint(len(user.Password)) < MinimalPasswordLength {
		return false, "Password is too small"
	}
	if uint(len(user.Password)) > MaxPasswordLength {
		return false, fmt.Sprintf("Password is too big; Password should be up to %d characters", MaxPasswordLength)
	}

	return true, ""
}

// Checks if such a user exists and compares passwords. Returns true if such user exists and passwords do match
func IsUserAuthorized(db *db.DB, user db.User) bool {
	userDB, err := db.GetUser(user.Email)
	if err != nil {
		return false
	}

	if userDB.Password != user.Password {
		return false
	}

	return true
}

// Returns email and password from a cookie. If an error is encountered, returns empty strings
func AuthFromCookie(cookie *http.Cookie) (string, string) {
	if cookie == nil {
		return "", ""
	}

	parts := strings.Split(cookie.Value, ":")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

/*
Gets auth information from a request and
checks if such a user exists and compares passwords.
Returns true if such user exists and passwords do match
*/
func IsUserAuthorizedReq(req *http.Request, dbase *db.DB) bool {
	var email, password string
	var ok bool
	email, password, ok = req.BasicAuth()
	if !ok || email == "" || password == "" {
		cookie, err := req.Cookie("auth")
		if err != nil {
			return false
		}

		email, password = AuthFromCookie(cookie)
	}

	return IsUserAuthorized(dbase, db.User{
		Email:    email,
		Password: password,
	})
}

// Returns email value from basic auth or from cookie if the former does not exist
func GetLoginFromReq(req *http.Request) string {
	email, _, ok := req.BasicAuth()
	if !ok || email == "" {
		cookie, err := req.Cookie("auth")
		if err != nil {
			return ""
		}

		email, _ = AuthFromCookie(cookie)
	}

	return email
}
