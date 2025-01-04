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
	OwnerEmail         string `json:"ownerEmail"`
	IsDone             bool   `json:"isDone"`
	CompletionTimeUnix uint64 `json:"completionTimeUnix"`
	Image              []byte `json:"image"`
	TimeCreated        string
	CompletionTime     string
	Due                string
}

func unixToTimeStr(unixTimeSec uint64) string {
	timeUnix := time.Unix(int64(unixTimeSec), 0)
	if timeUnix.Year() == 1970 {
		return "None"
	} else {
		return timeUnix.Format(time.DateOnly)
	}
}

func scanTodo(rows *sql.Rows) (*Todo, error) {
	var newTodo Todo
	err := rows.Scan(
		&newTodo.ID,
		&newTodo.GroupID,
		&newTodo.Text,
		&newTodo.TimeCreatedUnix,
		&newTodo.DueUnix,
		&newTodo.OwnerEmail,
		&newTodo.IsDone,
		&newTodo.CompletionTimeUnix,
		&newTodo.Image,
	)
	if err != nil {
		return nil, err
	}

	// Convert to Basic time
	newTodo.TimeCreated = unixToTimeStr(newTodo.TimeCreatedUnix)
	newTodo.Due = unixToTimeStr(newTodo.DueUnix)
	newTodo.CompletionTime = unixToTimeStr(newTodo.CompletionTimeUnix)

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
		"INSERT INTO todos(group_id, text, time_created_unix, due_unix, owner_email, is_done, completion_time_unix, image) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		todo.GroupID,
		todo.Text,
		todo.TimeCreatedUnix,
		todo.DueUnix,
		todo.OwnerEmail,
		todo.IsDone,
		todo.CompletionTimeUnix,
		todo.Image,
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

// Updates TODO's due date, text, done state, completion time and group id with image
func (db *DB) UpdateTodo(todoID uint64, updatedTodo Todo) error {
	_, err := db.Exec(
		"UPDATE todos SET group_id=?, due_unix=?, text=?, is_done=?, completion_time_unix=?, image=?  WHERE id=?",
		updatedTodo.GroupID,
		updatedTodo.DueUnix,
		updatedTodo.Text,
		updatedTodo.IsDone,
		updatedTodo.CompletionTimeUnix,
		updatedTodo.Image,
		todoID,
	)

	return err
}

// Searches and retrieves TODO groups created by the user
func (db *DB) GetAllUserTodoGroups(email string) ([]*TodoGroup, error) {
	var todoGroups []*TodoGroup

	rows, err := db.Query(
		"SELECT * FROM todo_groups WHERE owner_email=?",
		email,
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
func (db *DB) GetAllUserTodos(email string) ([]*Todo, error) {
	var todos []*Todo

	rows, err := db.Query(
		"SELECT * FROM todos WHERE owner_email=?",
		email,
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
func (db *DB) DeleteAllUserTodos(email string) error {
	_, err := db.Exec(
		"DELETE FROM todos WHERE owner_email=?",
		email,
	)

	return err
}

// Deletes all information regarding TODO groups of specified user
func (db *DB) DeleteAllUserTodoGroups(email string) error {
	_, err := db.Exec(
		"DELETE FROM todo_groups WHERE owner_email=?",
		email,
	)

	return err
}

func (db *DB) DoesUserOwnTodo(todoId uint64, email string) bool {
	todo, err := db.GetTodo(todoId)
	if err != nil {
		return false
	}

	if todo.OwnerEmail != email {
		return false
	}

	return true
}
