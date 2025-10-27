package models

import (
	"database/sql"
	"errors"
	"strings"

	/* "errors"
	"time" */
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleNormal = iota
	RoleModerator
	RoleAdmin
)

// A data object (entity) Holds data. Represents a single user in the system.
type User struct {
	ID       string
	Username string
	Email    string
	Password string
	Avatar   string
	Role     int
}

// A service object that manages behaviour (connects, and interacts with the database)
type UsersModel struct {
	DB *sql.DB
}

// Get fetches a specific user from the database by the user ID
func (m *UsersModel) Get(id string) (*User, error) {
	user := &User{}
	statement := `SELECT id, username, email, password, avatar, role FROM users WHERE id = ?`

	// Use QueryRowContext for better cancellation and timeout handling
	err := m.DB.QueryRow(statement, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Avatar, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecords
		}
		return nil, err
	}
	return user, nil
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
	statement := `INSERT INTO users (id,username,email,password,avatar,role) 
	VALUES(?,?,?,?,?,?)`

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
	emailOrUsername = strings.TrimSpace(emailOrUsername)

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

	// Compare provided password if it matches the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	return id, nil
}
