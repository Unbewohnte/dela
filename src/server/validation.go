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
	"fmt"
)

const (
	MinimalUsernameLength       uint   = 3
	ForbiddenUsernameCharacters string = "|<>\"'`\\/\u200b"
	MinimalPasswordLength       uint   = 5
)

// Check if user is valid. Returns false and a reason-string if not
func IsUserValid(user db.User) (bool, string) {
	if uint(len(user.Username)) < MinimalUsernameLength {
		return false, "Username is too small"
	}

	for _, char := range user.Username {
		for _, forbiddenChar := range ForbiddenUsernameCharacters {
			if char == forbiddenChar {
				return false, fmt.Sprintf("Username contains a forbidden character \"%c\"", char)
			}
		}
	}

	if uint(len(user.Password)) < MinimalPasswordLength {
		return false, "Password is too small"
	}

	return true, ""
}
