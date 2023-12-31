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

// Todo group structure
type TodoGroup struct {
	ID              uint64 `json:"id"`
	Name            string `json:"name"`
	TimeCreatedUnix uint64 `json:"timeCreatedUnix"`
	OwnerUsername   string `json:"ownerUsername"`
}

// Todo structure
type Todo struct {
	ID                 uint64 `json:"id"`
	GroupID            uint64 `json:"groupId"`
	Text               string `json:"text"`
	TimeCreatedUnix    uint64 `json:"timeCreatedUnix"`
	DueUnix            uint64 `json:"dueUnix"`
	OwnerUsername      string `json:"ownerUsername"`
	IsDone             bool   `json:"isDone"`
	CompletionTimeUnix uint64 `json:"completionTimeUnix"`
}

// Creates a new TODO group in the database
func (db *DB) CreateTodoGroup(group TodoGroup) error {
	_, err := db.Exec(
		"INSERT INTO todo_groups(name, time_created_unix, owner_username) VALUES(?, ?, ?)",
		group.Name,
		group.TimeCreatedUnix,
		group.OwnerUsername,
	)

	return err
}

func scanTodoGroup(rows *sql.Rows) (*TodoGroup, error) {
	var newTodoGroup TodoGroup
	err := rows.Scan(
		&newTodoGroup.ID,
		&newTodoGroup.Name,
		&newTodoGroup.TimeCreatedUnix,
		&newTodoGroup.OwnerUsername,
	)
	if err != nil {
		return nil, err
	}

	return &newTodoGroup, nil
}

// Retrieves a TODO group with provided ID from the database
func (db *DB) GetTodoGroup(id uint64) (*TodoGroup, error) {
	rows, err := db.Query(
		"SELECT * FROM todo_groups WHERE id=?",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()
	todoGroup, err := scanTodoGroup(rows)
	if err != nil {
		return nil, err
	}

	return todoGroup, nil
}

// Retrieves information on ALL TODO groups
func (db *DB) GetTodoGroups() ([]*TodoGroup, error) {
	var groups []*TodoGroup

	rows, err := db.Query("SELECT * FROM todo_groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		todoGroup, err := scanTodoGroup(rows)
		if err != nil {
			return groups, err
		}
		groups = append(groups, todoGroup)
	}

	return groups, nil
}

// Deletes information about a TODO group of given ID from the database
func (db *DB) DeleteTodoGroup(id uint64) error {
	_, err := db.Exec(
		"DELETE FROM todo_groups WHERE id=?",
		id,
	)

	return err
}

// Updates TODO group's name
func (db *DB) UpdateTodoGroup(groupID uint64, updatedGroup TodoGroup) error {
	_, err := db.Exec(
		"UPDATE todo_groups SET name=?  WHERE id=?",
		updatedGroup.Name,
		groupID,
	)

	return err
}

func scanTodo(rows *sql.Rows) (*Todo, error) {
	var newTodo Todo
	err := rows.Scan(
		&newTodo.ID,
		&newTodo.GroupID,
		&newTodo.Text,
		&newTodo.TimeCreatedUnix,
		&newTodo.DueUnix,
		&newTodo.OwnerUsername,
		&newTodo.IsDone,
		&newTodo.CompletionTimeUnix,
	)
	if err != nil {
		return nil, err
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
		"INSERT INTO todos(group_id, text, time_created_unix, due_unix, owner_username, is_done, completion_time_unix) VALUES(?, ?, ?, ?, ?, ?, ?)",
		todo.GroupID,
		todo.Text,
		todo.TimeCreatedUnix,
		todo.DueUnix,
		todo.OwnerUsername,
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
func (db *DB) GetAllUserTodoGroups(username string) ([]*TodoGroup, error) {
	var todoGroups []*TodoGroup

	rows, err := db.Query(
		"SELECT * FROM todo_groups WHERE owner_username=?",
		username,
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
func (db *DB) GetAllUserTodos(username string) ([]*Todo, error) {
	var todos []*Todo

	rows, err := db.Query(
		"SELECT * FROM todos WHERE owner_username=?",
		username,
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
func (db *DB) DeleteAllUserTodos(username string) error {
	_, err := db.Exec(
		"DELETE FROM todos WHERE owner_username=?",
		username,
	)

	return err
}

// Deletes all information regarding TODO groups of specified user
func (db *DB) DeleteAllUserTodoGroups(username string) error {
	_, err := db.Exec(
		"DELETE FROM todo_groups WHERE owner_username=?",
		username,
	)

	return err
}
