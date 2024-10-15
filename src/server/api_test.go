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

package server

import (
	"Unbewohnte/dela/conf"
	"Unbewohnte/dela/db"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestApi(t *testing.T) {
	// Create a new server
	config := conf.Default()
	config.BaseContentDir = "../../"
	config.ProdDBName = filepath.Join(os.TempDir(), "dela_test_db.db")
	server, err := New(config)
	if err != nil {
		t.Fatalf("failed to create a new server: %s", err)
	}
	defer os.Remove(config.ProdDBName)

	go func() {
		time.Sleep(time.Second * 5)
		server.Stop()
	}()

	go func() {
		server.Start()
	}()

	// Create a new user
	newUser := db.User{
		Login:           "user1",
		Password:        "ruohguoeruoger",
		TimeCreatedUnix: 12421467,
	}
	newUserJsonBytes, err := json.Marshal(&newUser)
	if err != nil {
		t.Fatalf("could not marshal new user JSON: %s", err)
	}

	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/api/user", config.Port), "application/json", bytes.NewBuffer(newUserJsonBytes))
	if err != nil {
		t.Fatalf("failed to post a new user data: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body for user creation: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got non-OK status code for user creation: %s", string(body))
	}
	resp.Body.Close()

	// Create a new TODO
	// newGroup := db.TodoGroup{
	// 	Name:            "group1",
	// 	TimeCreatedUnix: 13524534,
	// 	OwnerUsername:   newUser.Login,
	// }
	// newGroupBytes, err := json.Marshal(&newGroup)
	// if err != nil {
	// 	t.Fatalf("could not marshal new user JSON: %s", err)
	// }

	// req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/api/groups", config.Port), bytes.NewBuffer(newGroupBytes))
	// if err != nil {
	// 	t.Fatalf("failed to create a new POST request to create a new TODO group: %s", err)
	// }
	// req.Header.Add(RequestHeaderAuthKey, fmt.Sprintf("%s%s%s", newUser.Login, RequestHeaderAuthSeparator, newUser.Password))
	// req.Header.Add(RequestHeaderEncodedB64, "false")

	// resp, err = http.DefaultClient.Do(req)
	// if err != nil {
	// 	t.Fatalf("failed to post a new TODO group: %s", err)
	// }

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body for TODO group creation: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got non-OK status code for TODO group creation: %s", string(body))
	}
	resp.Body.Close()

	// TODO creation
	var newTodo db.Todo = db.Todo{
		GroupID:         0,
		Text:            "Do the dishes",
		TimeCreatedUnix: uint64(time.Now().UnixMicro()),
		DueUnix:         uint64(time.Now().Add(time.Hour * 5).UnixMicro()),
		OwnerLogin:      newUser.Login,
	}

	newTodoBytes, err := json.Marshal(&newTodo)
	if err != nil {
		t.Fatalf("could not marshal new Todo: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/api/todo", config.Port), bytes.NewBuffer(newTodoBytes))
	if err != nil {
		t.Fatalf("failed to create a new POST request to create a new TODO: %s", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to post a new Todo: %s", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body for Todo creation: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got non-OK status code for Todo creation: %s", string(body))
	}
	resp.Body.Close()
}
