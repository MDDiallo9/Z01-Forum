package models

import (
	"database/sql"
	"errors"
	"time"
)

// A data object. Holds the data for a single request
type Report struct {
	ID        int
	PostID    int
	UserID    string
	Reason    string
	Status    string
	CreatedAt time.Time
	Username  string
}

// A service object that manages database acess methods for moderator requests
type ReportsModel struct {
	DB *sql.DB
}

// Create inserts a new moderator reqport into the database for a given user
func (m *ReportsModel) Create(postID int, userID, reason string) error {
	// Create a new report.
	// A user can have multiple reports open at a time
	statement := `INSERT INTO reports (post_id, user_id, reason, status) VALUES (?, ?, ?, 'pending')`
	_, err := m.DB.Exec(statement, postID, userID, reason)
	if err != nil {
		return err
	}
	return err
}

// List retrieves all reports, optionally filtered by status.
// Joining the users table helps us attach username to each list
func (m *ReportsModel) List(status string) ([]*Report, error) {
	statement := `SELECT r.id, r.post_id, r.user_id, r.reason, r.status, r.created_at, u.username
	FROM reports r
	JOIN users u ON r.user_id = u.id`

	// Sort by status
	var args []interface{}
	if status != "" { // If query doesn't have a status filter on
		statement = statement + " WHERE r.status = ?"
		args = append(args, status)
	}
	statement = statement + " ORDER BY r.created_at DESC"

	rows, err := m.DB.Query(statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		report := &Report{}
		err := rows.Scan(&report.ID, &report.PostID, &report.UserID, &report.Reason, &report.Status, &report.CreatedAt, &report.Username)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

// UpdateStatus changes the status of a specific report
func (m *ReportsModel) UpdateStatus(reportID int, newStatus string) error {
	statement := `UPDATE reports SET status = ? WHERE id = ?`
	result, err := m.DB.Exec(statement, newStatus, reportID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNoRecords
	}

	return nil
}

// Get retrieves a single moderator report by ID
func (m *ReportsModel) Get(reportID int) (*Report, error) {
	report := &Report{}
	statement := `SELECT id, post_id, user_id, reason, status, created_at FROM reports WHERE id = ?`
	err := m.DB.QueryRow(statement, reportID).Scan(&report.ID, &report.PostID, &report.UserID, &report.Reason, &report.Status, &report.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecords
		}
		return nil, err
	}
	return report, nil
}
