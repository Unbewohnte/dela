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

package db

import "database/sql"

// User structure
type User struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	TimeCreatedUnix uint64 `json:"timeCreatedUnix"`
}

func scanUser(rows *sql.Rows) (*User, error) {
	rows.Next()
	var user User
	err := rows.Scan(&user.Username, &user.Password, &user.TimeCreatedUnix)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Searches for user with username and returns it
func (db *DB) GetUser(username string) (*User, error) {
	rows, err := db.Query("SELECT * FROM users WHERE username=?", username)
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
		"INSERT INTO users(username, password, time_created_unix) VALUES(?, ?, ?)",
		newUser.Username,
		newUser.Password,
		newUser.TimeCreatedUnix,
	)

	return err
}

// Deletes user with given username
func (db *DB) DeleteUser(username string) error {
	_, err := db.Exec(
		"DELETE FROM users WHERE username=?",
		username,
	)

	return err
}

// Deletes a user and all his TODOs (with groups) as well
func (db *DB) DeleteUserClean(username string) error {
	err := db.DeleteAllUserTodoGroups(username)
	if err != nil {
		return err
	}

	err = db.DeleteAllUserTodos(username)
	if err != nil {
		return err
	}

	err = db.DeleteUser(username)
	if err != nil {
		return err
	}

	return nil
}
