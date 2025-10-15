package models

import (
	"database/sql"
	"errors"

	/* "errors"
	"time" */
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type UsersModel struct {
	DB *sql.DB
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (m *UsersModel) Register(username, email, password, avatar string, role int) (string, error) {
	UUID := uuid.New().String()
	hashedPw, err := HashPassword(password)
	if err != nil {
		return "", err
	}
	session_id := uuid.New().String()
	statement := `INSERT INTO users (id,username,email,password,avatar,role,session_id,session_created_at) 
	VALUES(?,?,?,?,?,?,?,datetime())`

	_, err = m.DB.Exec(statement, UUID, username, email, hashedPw, avatar, role, session_id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return "", ErrDuplicateRecord
		}
		return "", err
	}

	// Returned ID is int64 type , we convert it before returning
	return UUID, nil
}

func (m *UsersModel) Authenticate(emailOrUsername, password string) (string, error) {
	var id string
	var hashedPassword string
	statement := `SELECT id, password FROM users WHERE email = ? OR username = ?`
	row := m.DB.QueryRow(statement, emailOrUsername, emailOrUsername)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}
	return id, nil
}
