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
			username TEXT PRIMARY KEY UNIQUE,
			password TEXT NOT NULL,
			time_created_unix INTEGER)`,
	)
	if err != nil {
		return err
	}

	// Todo groups
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todo_groups(
		id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		name TEXT,
		time_created_unix INTEGER,
		owner_username TEXT NOT NULL,
		FOREIGN KEY(owner_username) REFERENCES users(username))`,
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
		owner_username TEXT NOT NULL,
		FOREIGN KEY(group_id) REFERENCES todo_groups(id),
		FOREIGN KEY(owner_username) REFERENCES users(username))`,
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
