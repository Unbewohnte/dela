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

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApi(t *testing.T) {
	db, err := Create(filepath.Join(os.TempDir(), "dela_test.db"))
	if err != nil {
		t.Fatalf("failed to create database: %s", err)
	}

	// User
	user := User{
		Login:           "user1",
		Password:        "ruohguoeruoger",
		TimeCreatedUnix: 12421467,
	}

	err = db.CreateUser(user)
	if err != nil {
		t.Fatalf("failed to create user: %s", err)
	}

	dbUser, err := db.GetUser(user.Login)
	if err != nil {
		t.Fatalf("failed to retrieve created user: %s", err)
	}

	if dbUser.Password != user.Password {
		t.Fatalf("user passwords don't match")
	}

	// Todos
	group := TodoGroup{
		Name:            "group1",
		TimeCreatedUnix: 13524534,
		OwnerLogin:      user.Login,
	}

	err = db.CreateTodoGroup(group)
	if err != nil {
		t.Fatalf("failed to create todo group: %s", err)
	}

	err = db.UpdateTodoGroup(1, TodoGroup{Name: "updated_name"})
	if err != nil {
		t.Fatalf("failed to update todo group: %s", err)
	}

	dbGroup, err := db.GetTodoGroup(1)
	if err != nil {
		t.Fatalf("failed to get created TODO group: %s", err)
	}

	if dbGroup.Name == group.Name {
		t.Fatalf("name match changed value for a TODO group")
	}

	todo := Todo{
		GroupID:         dbGroup.ID,
		Text:            "Do the dishes",
		TimeCreatedUnix: dbGroup.TimeCreatedUnix,
		DueUnix:         0,
		OwnerLogin:      user.Login,
	}
	err = db.CreateTodo(todo)
	if err != nil {
		t.Fatalf("couldn't create a new TODO: %s", err)
	}

	// Now deletion
	err = db.DeleteUserClean(user.Login)
	if err != nil {
		t.Fatalf("couldn't cleanly delete user with all TODOs: %s", err)
	}
}
