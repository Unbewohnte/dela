package db

import (
	"database/sql"
)

type Verification struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	Code        string `json:"code"`
	IssuedUnix  uint64 `json:"issued_unix"`
	LifeSeconds uint64 `json:"life_seconds"`
}

func NewVerification(email string, code string, issuedUnix uint64, lifeSeconds uint64) *Verification {
	return &Verification{
		Email:       email,
		Code:        code,
		IssuedUnix:  issuedUnix,
		LifeSeconds: lifeSeconds,
	}
}

func scanVerification(rows *sql.Rows) (*Verification, error) {
	var newVerification Verification
	err := rows.Scan(
		&newVerification.ID,
		&newVerification.Email,
		&newVerification.Code,
		&newVerification.IssuedUnix,
		&newVerification.LifeSeconds,
	)
	if err != nil {
		return nil, err
	}

	return &newVerification, nil
}

// Retrieves a verification with given Id from the database
func (db *DB) GetVerification(id uint64) (*Verification, error) {
	rows, err := db.Query(
		"SELECT * FROM verifications WHERE id=?",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()
	verification, err := scanVerification(rows)
	if err != nil {
		return nil, err
	}

	return verification, nil
}

// Returns the last email verification by email
func (db *DB) GetVerificationByEmail(email string) (*Verification, error) {
	rows, err := db.Query(
		"SELECT * FROM verifications WHERE (email=?) ORDER BY life_seconds DESC",
		email,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()
	verification, err := scanVerification(rows)
	if err != nil {
		return nil, err
	}

	return verification, nil
}

// Retrieves information on ALL TODOs
func (db *DB) GetVerifications() ([]*Verification, error) {
	var verifications []*Verification

	rows, err := db.Query("SELECT * FROM verifications")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		verification, err := scanVerification(rows)
		if err != nil {
			return verifications, err
		}
		verifications = append(verifications, verification)
	}

	return verifications, nil
}

// Creates a new verification in the database
func (db *DB) CreateVerification(verification Verification) error {
	_, err := db.Exec(
		"INSERT INTO verifications(email, code, issued_unix, life_seconds) VALUES(?, ?, ?, ?)",
		verification.Email,
		verification.Code,
		verification.IssuedUnix,
		verification.LifeSeconds,
	)

	return err
}

// Deletes information about a verification of certain ID from the database
func (db *DB) DeleteVerification(id uint64) error {
	_, err := db.Exec(
		"DELETE FROM verifications WHERE id=?",
		id,
	)

	return err
}

// Updates verification
func (db *DB) UpdateVerification(verificationID uint64, updatedTodo Verification) error {
	_, err := db.Exec(
		"UPDATE verifications SET code=?, life_seconds=?  WHERE id=?",
		updatedTodo.Code,
		updatedTodo.LifeSeconds,
		verificationID,
	)

	return err
}
