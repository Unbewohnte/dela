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
	Login           string `json:"login"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	TimeCreatedUnix uint64 `json:"timeCreatedUnix"`
}

func scanUser(rows *sql.Rows) (*User, error) {
	rows.Next()
	var user User
	err := rows.Scan(&user.Login, &user.Email, &user.Password, &user.TimeCreatedUnix)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Searches for user with login and returns it
func (db *DB) GetUser(login string) (*User, error) {
	rows, err := db.Query("SELECT * FROM users WHERE login=?", login)
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
		"INSERT INTO users(login, email, password, time_created_unix) VALUES(?, ?, ?, ?)",
		newUser.Login,
		newUser.Email,
		newUser.Password,
		newUser.TimeCreatedUnix,
	)

	return err
}

// Deletes user with given login
func (db *DB) DeleteUser(login string) error {
	_, err := db.Exec(
		"DELETE FROM users WHERE login=?",
		login,
	)

	return err
}

func (db *DB) UserUpdate(newUser User) error {
	_, err := db.Exec(
		"UPDATE users SET email=? password=? WHERE login=?",
		newUser.Email,
		newUser.Password,
		newUser.Login,
	)

	return err
}

// Deletes a user and all his TODOs (with groups) as well
func (db *DB) DeleteUserClean(login string) error {
	err := db.DeleteAllUserTodoGroups(login)
	if err != nil {
		return err
	}

	err = db.DeleteAllUserTodos(login)
	if err != nil {
		return err
	}

	err = db.DeleteUser(login)
	if err != nil {
		return err
	}

	return nil
}
