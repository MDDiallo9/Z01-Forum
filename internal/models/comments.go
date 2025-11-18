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

func (m *CommentsModel) ListByUser(userID string) ([]*Comment, error) {
	statement := `SELECT c.id, c.content, c.author_id, u.username, c.post_id, c.created_at 
	FROM comments c
	JOIN users u ON c.author_id = u.id
	WHERE c.author_id = ?
	ORDER BY c.created_at DESC`

	rows, err := m.DB.Query(statement, userID)
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

func (m *CommentsModel) Get(id int) (*Comment, error) {
	statement := `SELECT c.id, c.content, c.author_id, u.username, c.post_id, c.created_at 
	FROM comments c
	JOIN users u ON c.author_id = u.id
	WHERE c.id = ?`

	c := &Comment{}
	err := m.DB.QueryRow(statement, id).Scan(&c.ID, &c.Content, &c.AuthorID, &c.Username, &c.PostID, &c.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecords
		}
		return nil, err
	}
	return c, nil
}

func (m *CommentsModel) Delete(id int) error {
	statement := `DELETE FROM comments WHERE id = ?`
	_, err := m.DB.Exec(statement, id)
	return err
}
