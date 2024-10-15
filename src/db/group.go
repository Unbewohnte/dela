package db

import "database/sql"

// Todo group structure
type TodoGroup struct {
	ID              uint64 `json:"id"`
	Name            string `json:"name"`
	TimeCreatedUnix uint64 `json:"timeCreatedUnix"`
	OwnerLogin      string `json:"ownerLogin"`
}

// Creates a new TODO group in the database
func (db *DB) CreateTodoGroup(group TodoGroup) error {
	_, err := db.Exec(
		"INSERT INTO todo_groups(name, time_created_unix, owner_username) VALUES(?, ?, ?)",
		group.Name,
		group.TimeCreatedUnix,
		group.OwnerLogin,
	)

	return err
}

func scanTodoGroup(rows *sql.Rows) (*TodoGroup, error) {
	var newTodoGroup TodoGroup
	err := rows.Scan(
		&newTodoGroup.ID,
		&newTodoGroup.Name,
		&newTodoGroup.TimeCreatedUnix,
		&newTodoGroup.OwnerLogin,
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

func (db *DB) DoesUserOwnGroup(groupId uint64, login string) bool {
	group, err := db.GetTodoGroup(groupId)
	if err != nil {
		return false
	}

	if group.OwnerLogin != login {
		return false
	}

	return true
}
