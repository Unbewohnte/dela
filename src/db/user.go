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

import "database/sql"

// User structure
type User struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	TimeCreatedUnix uint64 `json:"timeCreatedUnix"`
	ConfirmedEmail  bool   `json:"confirmedEmail"`
}

func scanUser(rows *sql.Rows) (*User, error) {
	rows.Next()
	var user User
	err := rows.Scan(&user.Email, &user.Password, &user.TimeCreatedUnix, &user.ConfirmedEmail)
	if err != nil {
		return nil, err
	}

	return &user, nil
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
		"INSERT INTO users(email, password, time_created_unix, confirmed_email) VALUES(?, ?, ?, ?)",
		newUser.Email,
		newUser.Password,
		newUser.TimeCreatedUnix,
		newUser.ConfirmedEmail,
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

// Updades user's email address, password, email confirmation with given email address
func (db *DB) UserUpdate(newUser User) error {
	_, err := db.Exec(
		"UPDATE users SET email=?, password=?, confirmed_email=? WHERE email=?",
		newUser.Email,
		newUser.Password,
		newUser.ConfirmedEmail,
		newUser.Email,
	)

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
