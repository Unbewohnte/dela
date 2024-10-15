/*
  	dela - web TODO list
    Copyright (C) 2023  Kasyanov Nikolay Alexeyevich (Unbewohnte)

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
	"net/http"
)

const (
	MinimalLoginLength    uint = 3
	MinimalUsernameLength uint = 3
	MinimalPasswordLength uint = 5
)

// Check if user is valid. Returns false and a reason-string if not
func IsUserValid(user db.User) (bool, string) {
	if uint(len(user.Login)) < MinimalLoginLength {
		return false, "Login is too small"
	}

	if uint(len(user.Password)) < MinimalPasswordLength {
		return false, "Password is too small"
	}

	for _, char := range user.Password {
		if char < 0x21 || char > 0x7E {
			// Not printable ASCII char!
			return false, "Password has a non printable ASCII character"
		}
	}

	return true, ""
}

// Checks if such a user exists and compares passwords. Returns true if such user exists and passwords do match
func IsUserAuthorized(db *db.DB, user db.User) bool {
	userDB, err := db.GetUser(user.Login)
	if err != nil {
		return false
	}

	if userDB.Password != user.Password {
		return false
	}

	return true
}

/*
Gets auth information from a request and
checks if such a user exists and compares passwords.
Returns true if such user exists and passwords do match
*/
func IsUserAuthorizedReq(req *http.Request, dbase *db.DB) bool {
	login, password, ok := req.BasicAuth()
	if !ok {
		return false
	}

	return IsUserAuthorized(dbase, db.User{
		Login:    login,
		Password: password,
	})
}

func GetLoginFromAuth(req *http.Request) string {
	login, _, _ := req.BasicAuth()
	return login
}
