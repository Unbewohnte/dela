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
