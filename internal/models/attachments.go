package models

import "database/sql"

// Attachment represents a single media file linked to a post or comment.
type Attachment struct {
	ID         int
	FilePath   string
	PostID     sql.NullInt64
	CommentID  sql.NullInt64
}

// AttachmentsModel provides database access for attachment records.
type AttachmentsModel struct {
	DB *sql.DB
}

// CreateForPost inserts a new attachment record linked to a post.
func (m *AttachmentsModel) CreateForPost(filePath string, postID int64) error {
	stmt := `INSERT INTO attachments (file_path, post_id) VALUES (?, ?)`
	_, err := m.DB.Exec(stmt, filePath, postID)
	return err
}

// CreateForComment inserts a new attachment record linked to a comment.
func (m *AttachmentsModel) CreateForComment(filePath string, commentID int64) error {
	stmt := `INSERT INTO attachments (file_path, comment_id) VALUES (?, ?)`
	_, err := m.DB.Exec(stmt, filePath, commentID)
	return err
}
