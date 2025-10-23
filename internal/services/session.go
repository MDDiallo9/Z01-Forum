package services

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type SessionManager struct {
	DB         *sql.DB
	CookieName string
	LifeTime   time.Duration
	HardMax    time.Duration
	// ErrorLog   error
}

func (sm *SessionManager) CreateSession(w http.ResponseWriter, r *http.Request, UserID string) error {
	// Generate random session ID
	sessionID, err := GenerateRandomString(32)
	if err != nil {
		return err
	}

	// Calculate expiry times
	createdAt := time.Now().UTC()
	expiresAt := createdAt.Add(sm.LifeTime)

	// Extract client info for session binding
	ip := ExtractIPFromRequest(r)
	userAgent := ExtractUserAgent(r)

	// Insert session record nto sessions table in DB.sql
	statement := `INSERT INTO sessions (id, user_id, created_at, expires_at, ip_address, user_agent)
	VALUES(?, ?, ?, ?, ?, ?)`

	_, err = sm.DB.Exec(statement, sessionID, UserID, createdAt, expiresAt, ip, userAgent)
	if err != nil {
		return err
	}

	// Create cookies
	cookie := &http.Cookie{
		Name:     sm.CookieName,
		Value:    sessionID,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	return nil
}

func (sm *SessionManager) GetUserFromRequest(r *http.Request) (string, error) {
	// Define error varibles
	var (
		ErrSessionNotFound = errors.New("session not found")
		ErrSessionExpired  = errors.New("session expired")
		ErrSessionInvalid  = errors.New("session invalid")
	)
	// Retrieve session cookie
	cookie, err := r.Cookie(sm.CookieName)
	if err == http.ErrNoCookie {
		return "", nil // means: no session cookie, so treat as guest user
	}
	if err != nil {
		// sql.ErNoRows is a standard way to signal "not found" instead of just "err"
		return "", fmt.Errorf("failed to read cookie: %w", err)
	}
	sessionID := cookie.Value

	// Define variables to hold data from the DB
	var UserID string
	var createdAt time.Time
	var expiresAt time.Time
	var ip string
	var userAgent string

	// Fetch session record from the DB
	statement := `SELECT user_id, created_at, expires_at, ip_address, user_agent FROM sessions WHERE id = ?`

	err = sm.DB.QueryRow(statement, sessionID).Scan(&UserID, &createdAt, &expiresAt, &ip, &userAgent)
	if err != nil {
		// If DB query returns no rows, we treat it as an invalid session
		// Or a database error
		if err == sql.ErrNoRows {
			return "", ErrSessionNotFound
		}
		return "", err
	}

	// Validate expiry
	currentTime := time.Now().UTC()
	if currentTime.After(expiresAt) {
		// Session has truly expired
		statement := `DELETE FROM sessions WHERE id = ?`
		_, _ = sm.DB.Exec(statement, sessionID)
		return "", ErrSessionExpired
	}

	// Validate IP and User Agent
	currentIP := ExtractIPFromRequest(r)
	currentUA := ExtractUserAgent(r)

	if currentIP != ip || currentUA != userAgent {
		// TO DO: delete ession from DB
		statement := `DELETE FROM sessions WHERE id = ?`
		_, _ = sm.DB.Exec(statement, sessionID)
		return "", ErrSessionInvalid
	}

	// Rolling expiry extension
	newExpiresAt := currentTime.Add(sm.LifeTime)
	if createdAt.Add(sm.HardMax).After(newExpiresAt) {
		statement := `UPDATE sessions SET expires_at = ? WHERE id = ?`
		_, err = sm.DB.Exec(statement, newExpiresAt, sessionID)
	}

	return UserID, nil
}

func (sm *SessionManager) DestroySession(w http.ResponseWriter, r *http.Request) error {
	// Retrieve cookie from request
	cookie, err := r.Cookie(sm.CookieName)
	if err == http.ErrNoCookie {
		return err
	}
	if err != nil {
		return fmt.Errorf("failed to read cookie %w", err)
	}
	sessionID := cookie.Value

	// Delete session from DB
	statement := `DELETE FROM sessions WHERE id = ?`
	res, err := sm.DB.Exec(statement, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	} else {
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return err
		}
	}

	// Expire cookie on client
	expiredCookie := &http.Cookie{
		Name:     sm.CookieName,
		Value:    "",
		Expires:  time.Now(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, expiredCookie)

	return err
}
