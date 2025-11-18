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
	Categories   []int        `json:"categories"` // Changed from CategoryID int
	CreatedAt    time.Time    `json:"createdAt"`
	LastModified sql.NullTime `json:"lastModified"`
}

// PostModel (service object) interacts with the DB
type PostsModel struct {
	DB *sql.DB
}

func (m *PostsModel) Get(id int) (*Post, error) {
	post := &Post{}
	statement := `SELECT id, title, content, author_id, created_at, last_modified FROM posts WHERE id = ?`

	err := m.DB.QueryRow(statement, id).Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CreatedAt, &post.LastModified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecords
		}
		return nil, err
	}

	// Fetch categories
	categories, err := m.getCategoriesForPost(id)
	if err != nil {
		return nil, err
	}
	post.Categories = categories

	return post, nil
}

func (m *PostsModel) CreateNewPostDB(title, content, author_id string, category_ids []int) (int, error) {
	statement := `INSERT INTO posts (title,content,author_id,created_at)
	VALUES (?,?,?,datetime())`

	result, err := m.DB.Exec(statement, title, content, author_id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, ErrDuplicateRecord
		}
		return 0, err
	}
	id64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	id := int(id64)

	// Insert categories
	for _, catID := range category_ids {
		_, err := m.DB.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`, id, catID)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
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

func (m *PostsModel) UpdatePostDB(title, content, author_id string, category_ids []int, id int) error {
	statement := `UPDATE posts
    SET title = ?, content = ?, author_id = ?, last_modified = datetime()
    WHERE id = ?`

	result, err := m.DB.Exec(statement, title, content, author_id, id)
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

	// Update categories: Delete old ones and insert new ones
	_, err = m.DB.Exec(`DELETE FROM post_categories WHERE post_id = ?`, id)
	if err != nil {
		return err
	}

	for _, catID := range category_ids {
		_, err := m.DB.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`, id, catID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *PostsModel) ListRandom(limit int) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.created_at, p.last_modified
	FROM posts p
	JOIN users u ON p.author_id = u.id
	ORDER BY RANDOM()
	limit ?`

	rows, err := m.DB.Query(statement, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Username, &post.CreatedAt, &post.LastModified)
		if err != nil {
			return nil, err
		}

		categories, err := m.getCategoriesForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (m *PostsModel) ListByCategory(categoryID int, limit int) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.created_at, p.last_modified
	FROM posts p
	JOIN users u ON p.author_id = u.id
	JOIN post_categories pc ON p.id = pc.post_id
	WHERE pc.category_id = ?
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
			&post.CreatedAt, &post.LastModified,
		)
		if err != nil {
			return nil, err
		}

		categories, err := m.getCategoriesForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *PostsModel) ListByAuthor(authorID string) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.created_at, p.last_modified
	FROM posts p
	JOIN users u ON p.author_id = u.id
	WHERE p.author_id = ?
	ORDER BY p.created_at DESC`

	rows, err := m.DB.Query(statement, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Username,
			&post.CreatedAt, &post.LastModified,
		)
		if err != nil {
			return nil, err
		}

		categories, err := m.getCategoriesForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *PostsModel) ListLikedByUser(userID string) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.created_at, p.last_modified
	FROM posts p
	JOIN users u ON p.author_id = u.id
	JOIN reactions r ON p.id = r.post_id
	WHERE r.user_id = ? AND r.type = 1
	ORDER BY r.id DESC`

	rows, err := m.DB.Query(statement, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Username,
			&post.CreatedAt, &post.LastModified,
		)
		if err != nil {
			return nil, err
		}

		categories, err := m.getCategoriesForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *PostsModel) getCategoriesForPost(postID int) ([]int, error) {
	rows, err := m.DB.Query(`SELECT category_id FROM post_categories WHERE post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []int
	for rows.Next() {
		var catID int
		if err := rows.Scan(&catID); err != nil {
			return nil, err
		}
		categories = append(categories, catID)
	}
	return categories, nil
}
