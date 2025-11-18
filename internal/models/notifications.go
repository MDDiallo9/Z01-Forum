package models

import (
	"database/sql"
	"time"
)

type Notification struct {
	ID        int
	UserID    string
	ActorID   string
	ActorName string
	PostID    sql.NullInt64
	CommentID sql.NullInt64
	Type      string // 'like', 'dislike', 'comment'
	Read      bool
	CreatedAt time.Time
}

type NotificationsModel struct {
	DB *sql.DB
}

func (m *NotificationsModel) Create(userID, actorID, notifType string, postID, commentID *int) error {
	// Don't create notification if user is acting on their own content
	if userID == actorID {
		return nil
	}

	query := `
		INSERT INTO notifications (user_id, actor_id, type, post_id, comment_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := m.DB.Exec(query, userID, actorID, notifType, postID, commentID, time.Now())
	return err
}

func (m *NotificationsModel) ListByUser(userID string) ([]*Notification, error) {
	query := `
		SELECT n.id, n.user_id, n.actor_id, u.username, n.post_id, n.comment_id, n.type, n.read, n.created_at
		FROM notifications n
		JOIN users u ON n.actor_id = u.id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
	`
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []*Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.ActorID, &n.ActorName, &n.PostID, &n.CommentID, &n.Type, &n.Read, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifs = append(notifs, &n)
	}
	return notifs, nil
}

func (m *NotificationsModel) MarkAsRead(notifID int) error {
	query := `UPDATE notifications SET read = TRUE WHERE id = ?`
	_, err := m.DB.Exec(query, notifID)
	return err
}

func (m *NotificationsModel) MarkAllAsRead(userID string) error {
	query := `UPDATE notifications SET read = TRUE WHERE user_id = ?`
	_, err := m.DB.Exec(query, userID)
	return err
}
