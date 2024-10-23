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
	"time"
)

// Todo structure
type Todo struct {
	ID                 uint64 `json:"id"`
	GroupID            uint64 `json:"groupId"`
	Text               string `json:"text"`
	TimeCreatedUnix    uint64 `json:"timeCreatedUnix"`
	DueUnix            uint64 `json:"dueUnix"`
	OwnerLogin         string `json:"ownerLogin"`
	IsDone             bool   `json:"isDone"`
	CompletionTimeUnix uint64 `json:"completionTimeUnix"`
	TimeCreated        string
	CompletionTime     string
	Due                string
}

func scanTodo(rows *sql.Rows) (*Todo, error) {
	var newTodo Todo
	err := rows.Scan(
		&newTodo.ID,
		&newTodo.GroupID,
		&newTodo.Text,
		&newTodo.TimeCreatedUnix,
		&newTodo.DueUnix,
		&newTodo.OwnerLogin,
		&newTodo.IsDone,
		&newTodo.CompletionTimeUnix,
	)
	if err != nil {
		return nil, err
	}

	// Convert to Basic time
	timeCreated := time.Unix(int64(newTodo.TimeCreatedUnix), 0)
	if timeCreated.Year() == 1970 {
		newTodo.TimeCreated = "None"
	} else {
		newTodo.TimeCreated = timeCreated.Format(time.DateOnly)
	}

	due := time.Unix(int64(newTodo.DueUnix), 0)
	if due.Year() == 1970 {
		newTodo.Due = "None"
	} else {
		newTodo.Due = due.Format(time.DateOnly)
	}

	completionTime := time.Unix(int64(newTodo.CompletionTimeUnix), 0)
	if completionTime.Year() == 1970 {
		newTodo.CompletionTime = "None"
	} else {
		newTodo.CompletionTime = completionTime.Format(time.DateOnly)
	}

	return &newTodo, nil
}

// Retrieves a TODO with given Id from the database
func (db *DB) GetTodo(id uint64) (*Todo, error) {
	rows, err := db.Query(
		"SELECT * FROM todos WHERE id=?",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()
	todo, err := scanTodo(rows)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

// Retrieves information on ALL TODOs
func (db *DB) GetTodos() ([]*Todo, error) {
	var todos []*Todo

	rows, err := db.Query("SELECT * FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		todo, err := scanTodo(rows)
		if err != nil {
			return todos, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// Creates a new TODO in the database
func (db *DB) CreateTodo(todo Todo) error {
	_, err := db.Exec(
		"INSERT INTO todos(group_id, text, time_created_unix, due_unix, owner_login, is_done, completion_time_unix) VALUES(?, ?, ?, ?, ?, ?, ?)",
		todo.GroupID,
		todo.Text,
		todo.TimeCreatedUnix,
		todo.DueUnix,
		todo.OwnerLogin,
		todo.IsDone,
		todo.CompletionTimeUnix,
	)

	return err
}

// Deletes information about a TODO of certain ID from the database
func (db *DB) DeleteTodo(id uint64) error {
	_, err := db.Exec(
		"DELETE FROM todos WHERE id=?",
		id,
	)

	return err
}

// Updates TODO's due date, text, done state, completion time and group id
func (db *DB) UpdateTodo(todoID uint64, updatedTodo Todo) error {
	_, err := db.Exec(
		"UPDATE todos SET group_id=?, due_unix=?, text=?, is_done=?, completion_time_unix=?  WHERE id=?",
		updatedTodo.GroupID,
		updatedTodo.DueUnix,
		updatedTodo.Text,
		updatedTodo.IsDone,
		updatedTodo.CompletionTimeUnix,
		todoID,
	)

	return err
}

// Searches and retrieves TODO groups created by the user
func (db *DB) GetAllUserTodoGroups(login string) ([]*TodoGroup, error) {
	var todoGroups []*TodoGroup

	rows, err := db.Query(
		"SELECT * FROM todo_groups WHERE owner_login=?",
		login,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		group, err := scanTodoGroup(rows)
		if err != nil {
			continue
		}
		todoGroups = append(todoGroups, group)
	}

	return todoGroups, nil
}

// Searches and retrieves TODOs created by the user
func (db *DB) GetAllUserTodos(login string) ([]*Todo, error) {
	var todos []*Todo

	rows, err := db.Query(
		"SELECT * FROM todos WHERE owner_login=?",
		login,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		todo, err := scanTodo(rows)
		if err != nil {
			continue
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

// Deletes all information regarding TODOs of specified user
func (db *DB) DeleteAllUserTodos(login string) error {
	_, err := db.Exec(
		"DELETE FROM todos WHERE owner_login=?",
		login,
	)

	return err
}

// Deletes all information regarding TODO groups of specified user
func (db *DB) DeleteAllUserTodoGroups(login string) error {
	_, err := db.Exec(
		"DELETE FROM todo_groups WHERE owner_login=?",
		login,
	)

	return err
}

func (db *DB) DoesUserOwnTodo(todoId uint64, login string) bool {
	todo, err := db.GetTodo(todoId)
	if err != nil {
		return false
	}

	if todo.OwnerLogin != login {
		return false
	}

	return true
}
