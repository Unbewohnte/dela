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

package db

import (
	"database/sql"
)

// User structure
type User struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	TimeCreatedUnix uint64 `json:"timeCreatedUnix"`
	TimeCreated     string `json:"timeCreated"`
	ConfirmedEmail  bool   `json:"confirmedEmail"`
	NotifyOnTodos   bool   `json:"notifyOnTodos"`
}

func scanUserRaw(rows *sql.Rows) (*User, error) {
	var user User
	err := rows.Scan(&user.Email, &user.Password, &user.TimeCreatedUnix, &user.ConfirmedEmail, &user.NotifyOnTodos)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func scanUser(rows *sql.Rows) (*User, error) {
	rows.Next()
	user, err := scanUserRaw(rows)
	if err != nil {
		return nil, err
	}

	// Convert to Basic time string
	user.TimeCreated = unixToTimeStr(user.TimeCreatedUnix)

	return user, nil
}

// Searches for user with email and returns it
func (db *DB) GetUser(email string) (*User, error) {
	rows, err := db.Query("SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := scanUser(rows)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Creates a new user in the database
func (db *DB) CreateUser(newUser User) error {
	_, err := db.Exec(
		"INSERT INTO users(email, password, time_created_unix, confirmed_email, notify_on_todos) VALUES(?, ?, ?, ?, ?)",
		newUser.Email,
		newUser.Password,
		newUser.TimeCreatedUnix,
		newUser.ConfirmedEmail,
		newUser.NotifyOnTodos,
	)

	return err
}

// Deletes user with given email address
func (db *DB) DeleteUser(email string) error {
	_, err := db.Exec(
		"DELETE FROM users WHERE email=?",
		email,
	)

	return err
}

// Updades user's email address, password, email confirmation and todo notification status with given email address
func (db *DB) UserUpdate(newUser User) error {
	_, err := db.Exec(
		"UPDATE users SET email=?, password=?, confirmed_email=?, notify_on_todos=? WHERE email=?",
		newUser.Email,
		newUser.Password,
		newUser.ConfirmedEmail,
		newUser.NotifyOnTodos,
		newUser.Email,
	)

	return err
}

func (db *DB) UserSetNotifyOnTodos(email string, value bool) error {
	_, err := db.Exec("UPDATE users SET notify_on_todos=? WHERE email=?", value, email)
	return err
}

// Deletes a user and all his TODOs (with groups) as well
func (db *DB) DeleteUserClean(email string) error {
	err := db.DeleteAllUserTodoGroups(email)
	if err != nil {
		return err
	}

	err = db.DeleteAllUserTodos(email)
	if err != nil {
		return err
	}

	err = db.DeleteUser(email)
	if err != nil {
		return err
	}

	return nil
}

// Sets confirmed_email to true for given user
func (db *DB) UserSetEmailConfirmed(email string) error {
	_, err := db.Exec(
		"UPDATE users SET confirmed_email=? WHERE email=?",
		true,
		email,
	)

	return err
}

// Cleanly deletes user if email is not confirmed
func (db *DB) DeleteUnverifiedUserClean(email string) error {
	user, err := db.GetUser(email)
	if err != nil {
		return err
	}

	if !user.ConfirmedEmail {
		// Email is not verified, delete information on this user
		err = db.DeleteUserClean(email)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetAllUsersWithNotificationsOn() ([]*User, error) {
	rows, err := db.Query("SELECT * FROM users WHERE notify_on_todos=?", true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user, err := scanUserRaw(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}
