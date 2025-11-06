package models

import (
	"database/sql"
	"errors"
	"time"

	/* "errors"
	"time" */
	"github.com/mattn/go-sqlite3"
)

// Post (data object) represents a single post record from the database
type Post struct {
	ID           int          `json:"id"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	AuthorID     string       `json:"authorId"`
	Username     string       `json:"username"`
	CategoryID   int          `json:"categoryId"`
	CreatedAt    time.Time    `json:"createdAt"`
	LastModified sql.NullTime `json:"lastModified"`
}

// PostModel (service object) interacts with the DB
type PostsModel struct {
	DB *sql.DB
}

func (m *PostsModel) Get(id int) (*Post, error) {
	post := &Post{}
	statement := `SELECT id, title, content, author_id, category_id FROM posts WHERE id = ?`

	err := m.DB.QueryRow(statement, id).Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CategoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecords
		}
		return nil, err
	}
	return post, nil
}

func (m *PostsModel) CreateNewPostDB(title, content, author_id string, category_id int) (int, error) {
	statement := `INSERT INTO posts (title,content,author_id,category_id,created_at)
	VALUES (?,?,?,?,datetime())`

	result, err := m.DB.Exec(statement, title, content, author_id, category_id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, ErrDuplicateRecord
		}
		// TODO : Handle other errors related to db constraints
		return 0, err
	}
	id64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id64), nil
}

func (m *PostsModel) DeletePostDB(id int) error {
	statement := `DELETE from posts WHERE id = ?`

	result, err := m.DB.Exec(statement, id)
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

func (m *PostsModel) UpdatePostDB(title, content, author_id string, category_id, id int) error {
	statement := `UPDATE posts
    SET title = ?, content = ?, author_id = ?, category_id = ?, last_modified = datetime()
    WHERE id = ?`

	result, err := m.DB.Exec(statement, title, content, author_id, category_id, id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return ErrDuplicateRecord
		}
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNoRecords
	}

	return nil
}

func (m *PostsModel) ListRandom(limit int) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.category_id, p.created_at, p.last_modified
	FROM posts p
	JOIN users u ON p.author_id = u.id
	ORDER BY random
	limit ?`

	rows, err := m.DB.Query(statement, limit)
	if err != nil {
		// if errors.Is(err, sql.ErrNoRows) {
		// 	return nil, ErrNoRecords
		// }
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Username, &post.CategoryID, &post.CreatedAt, &post.LastModified)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (m *PostsModel) ListByCategory(categoryID int, limit int) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.category_id, p.created_at, p.last_modified
	FROM posts p
	JOIN users u ON p.author_id = u.id
	WHERE p.category_id = ?
	ORDER BY p.created_at DESC
	LIMIT ?`

	rows, err := m.DB.Query(statement, categoryID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Username,
			&post.CategoryID, &post.CreatedAt, &post.LastModified,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
