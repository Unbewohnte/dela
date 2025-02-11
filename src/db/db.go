/*
  	dela - web TODO list
    Copyright (C) 2023, 2024, 2025  Kasyanov Nikolay Alexeyevich (Unbewohnte)

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
	"os"

	_ "modernc.org/sqlite"
)

// Database wrapper
type DB struct {
	*sql.DB
}

func setUpTables(db *DB) error {
	// Users
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users(
		email TEXT PRIMARY KEY UNIQUE,
		password TEXT NOT NULL,
		time_created_unix INTEGER,
		confirmed_email INTEGER,
		notify_on_todos INTEGER)`,
	)
	if err != nil {
		return err
	}

	// User Email Verification
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS verifications(
		id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		email TEXT NOT NULL,
		code TEXT NOT NULL,
		issued_unix INTEGER,
		life_seconds INTEGER)`,
	)
	if err != nil {
		return err
	}

	// Todo groups
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todo_groups(
		id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		name TEXT,
		time_created_unix INTEGER,
		owner_email TEXT NOT NULL,
		removable INTEGER,
		FOREIGN KEY(owner_email) REFERENCES users(email))`,
	)
	if err != nil {
		return err
	}

	// Todos
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos(
		id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		group_id INTEGER NOT NULL,
		text TEXT NOT NULL,
		time_created_unix INTEGER,
		due_unix INTEGER,
		owner_email TEXT NOT NULL,
		is_done INTEGER,
		completion_time_unix INTEGER,
		image BLOB,
		file BLOB,
		FOREIGN KEY(group_id) REFERENCES todo_groups(id),
		FOREIGN KEY(owner_email) REFERENCES users(email))`,
	)
	if err != nil {
		return err
	}

	return nil
}

// Open database
func FromFile(path string) (*DB, error) {
	driver, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	dbase := &DB{driver}

	err = setUpTables(dbase)
	if err != nil {
		return nil, err
	}

	return dbase, nil
}

// Create database file
func Create(path string) (*DB, error) {
	dbFile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	dbFile.Close()

	driver, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	dbase := &DB{driver}

	err = setUpTables(dbase)
	if err != nil {
		return nil, err
	}

	return dbase, nil
}
