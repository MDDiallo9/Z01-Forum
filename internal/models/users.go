package models

import (
	"database/sql"
	/* "errors"
	"time" */
	"github.com/google/uuid"
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
		return "", err
	}

	// Returned ID is int64 type , we convert it before returning
	return UUID, nil
}
