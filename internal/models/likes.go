package models

import (
	"database/sql"
	"errors"
)

type Like struct {
	ID        int
	Type      int // 1 for like, -1 for dislike
	UserID    string
	PostID    *int
	CommentID *int
}

type LikesModel struct {
	DB *sql.DB
}

// AddLike adds a like or dislike. If a reaction already exists, it updates it.
// If the same reaction exists, it removes it (toggles).
func (m *LikesModel) AddLike(userID string, postID, commentID *int, reactionType int) error {
	// Check if reaction exists
	var currentType int
	var id int
	
	query := `SELECT id, type FROM reactions WHERE user_id = ? AND `
	var args []interface{}
	args = append(args, userID)

	if postID != nil {
		query += `post_id = ?`
		args = append(args, *postID)
	} else if commentID != nil {
		query += `comment_id = ?`
		args = append(args, *commentID)
	} else {
		return errors.New("postID and commentID cannot both be nil")
	}

	err := m.DB.QueryRow(query, args...).Scan(&id, &currentType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No reaction exists, create one
			insertQuery := `INSERT INTO reactions (user_id, post_id, comment_id, type) VALUES (?, ?, ?, ?)`
			// args already has user_id and post_id/comment_id. Need to reconstruct for insert
			insertArgs := []interface{}{userID, postID, commentID, reactionType}
			_, err := m.DB.Exec(insertQuery, insertArgs...)
			return err
		}
		return err
	}

	// Reaction exists
	if currentType == reactionType {
		// Same reaction, remove it (toggle off)
		_, err := m.DB.Exec(`DELETE FROM reactions WHERE id = ?`, id)
		return err
	} else {
		// Different reaction, update it
		_, err := m.DB.Exec(`UPDATE reactions SET type = ? WHERE id = ?`, reactionType, id)
		return err
	}
}

func (m *LikesModel) GetLikesCount(postID, commentID *int) (likes int, dislikes int, err error) {
	query := `SELECT type, COUNT(*) FROM reactions WHERE `
	var args []interface{}

	if postID != nil {
		query += `post_id = ?`
		args = append(args, *postID)
	} else if commentID != nil {
		query += `comment_id = ?`
		args = append(args, *commentID)
	} else {
		return 0, 0, errors.New("postID and commentID cannot both be nil")
	}

	query += ` GROUP BY type`

	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var rType, count int
		if err := rows.Scan(&rType, &count); err != nil {
			return 0, 0, err
		}
		if rType == 1 {
			likes = count
		} else if rType == -1 {
			dislikes = count
		}
	}

	return likes, dislikes, nil
}
