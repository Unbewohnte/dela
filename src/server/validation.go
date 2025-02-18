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
	"Unbewohnte/dela/i18n"
	"Unbewohnte/dela/misc"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	MinimalEmailLength    uint = 3
	MinimalPasswordLength uint = 5
	MaxEmailLength        uint = 50
	MaxPasswordLength     uint = 50
	MaxTodoTextLength     uint = 250
	MaxTodoFileSizeBytes  uint = 3145728 // 3MB
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

// Checks if such user exists, passwords match and email is confirmed. Returns true if such user exists, passwords do match and email was verified
func IsUserAuthorized(db *db.DB, user db.User) bool {
	userDB, err := db.GetUser(user.Email)
	if err != nil {
		return false
	}

	if !userDB.ConfirmedEmail {
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
checks if such user exists and passwords match.
Returns true if such user exists, passwords do match and email is confirmed
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
func GetEmailFromReq(req *http.Request) string {
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

/*
Generates a new verification code for given email with 8-digit numeric code,
current issue time and provided life time.
Inserts newly created email verification into database.
*/
func GenerateVerificationCode(dbase *db.DB, email string, length uint, lifeTimeSeconds uint64) (*db.Verification, error) {
	verification := db.NewVerification(
		email, misc.GenerateNumericCode(length), uint64(time.Now().Unix()), lifeTimeSeconds,
	)

	err := dbase.CreateVerification(*verification)
	if err != nil {
		return nil, err
	}

	return verification, nil
}

func LocaleFromReq(req *http.Request) string {
	cookie, err := req.Cookie("locale")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func LanguageFromReq(req *http.Request) i18n.Language {
	switch strings.ToUpper(LocaleFromReq(req)) {
	case "ENG":
		return i18n.ENG

	case "RU":
		return i18n.RU

	default:
		return i18n.ENG
	}
}
