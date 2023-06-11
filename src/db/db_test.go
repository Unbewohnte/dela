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
		Username:        "user1",
		Password:        "ruohguoeruoger",
		TimeCreatedUnix: 12421467,
	}

	err = db.CreateUser(user)
	if err != nil {
		t.Fatalf("failed to create user: %s", err)
	}

	dbUser, err := db.GetUser(user.Username)
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
		OwnerUsername:   user.Username,
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
		OwnerUsername:   user.Username,
	}
	err = db.CreateTodo(todo)
	if err != nil {
		t.Fatalf("couldn't create a new TODO: %s", err)
	}

	// Now deletion
	err = db.DeleteUserClean(user.Username)
	if err != nil {
		t.Fatalf("couldn't cleanly delete user with all TODOs: %s", err)
	}
}
