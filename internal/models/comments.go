package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"authorId"`
	Username  string    `json:"username"`
	PostID    int       `json:"postId"`
	CreatedAt time.Time `json:"createdAt"`
}

type CommentsModel struct {
	DB *sql.DB
}

func (m *CommentsModel) Create(content, authorID string, postID int) (int, error) {
	statement := `INSERT INTO comments (content, author_id, post_id, created_at) VALUES (?, ?, ?, datetime())`
	result, err := m.DB.Exec(statement, content, authorID, postID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *CommentsModel) GetByPostID(postID int) ([]*Comment, error) {
	statement := `SELECT c.id, c.content, c.author_id, u.username, c.post_id, c.created_at 
	FROM comments c
	JOIN users u ON c.author_id = u.id
	WHERE c.post_id = ?
	ORDER BY c.created_at ASC`

	rows, err := m.DB.Query(statement, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		err := rows.Scan(&c.ID, &c.Content, &c.AuthorID, &c.Username, &c.PostID, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}
