package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/mattn/go-sqlite3"
)

type Post struct {
	ID           int          `json:"id"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	AuthorID     string       `json:"authorId"`
	AuthorName   string       `json:"username"`
	ImageURL     string       `json:"imageUrl"`
	Categories   []int        `json:"categories"`
	CreatedAt    time.Time    `json:"createdAt"`
	LastModified sql.NullTime `json:"lastModified"`
	LikeCount    int          `json:"likeCount"`
	DislikeCount int          `json:"dislikeCount"`
	CommentCount int          `json:"commentCount"`
}

type PostsModel struct {
	DB *sql.DB
}

func (m *PostsModel) CreateNewPostDB(post Post) (int64, error) {
	query := `INSERT INTO posts (title, content, author_id, image_url, created_at) VALUES (?, ?, ?, ?, ?)`
	result, err := m.DB.Exec(query, post.Title, post.Content, post.AuthorID, post.ImageURL, time.Now())
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert categories
	for _, catID := range post.Categories {
		_, err := m.DB.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
		if err != nil {
			return 0, err
		}
	}

	return postID, nil
}

func (m *PostsModel) Get(id int) (*Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.author_id, u.username, p.image_url, p.created_at, p.last_modified,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = 1) as likes,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = -1) as dislikes,
		(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comments
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = ?
	`
	row := m.DB.QueryRow(query, id)

	var post Post
	var lastModified sql.NullTime
	var imageURL sql.NullString

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &imageURL, &post.CreatedAt, &lastModified, &post.LikeCount, &post.DislikeCount, &post.CommentCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecords
		}
		return nil, err
	}

	if lastModified.Valid {
		post.LastModified.Time = lastModified.Time
		post.LastModified.Valid = true
	}
	if imageURL.Valid {
		post.ImageURL = imageURL.String
	}

	// Fetch categories
	catQuery := `SELECT category_id FROM post_categories WHERE post_id = ?`
	rows, err := m.DB.Query(catQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var catID int
		if err := rows.Scan(&catID); err != nil {
			return nil, err
		}
		post.Categories = append(post.Categories, catID)
	}

	return &post, nil
}

func (m *PostsModel) ListAll() ([]*Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.author_id, u.username, p.image_url, p.created_at, p.last_modified,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = 1) as likes,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = -1) as dislikes,
		(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comments
		FROM posts p
		JOIN users u ON p.author_id = u.id
		ORDER BY COALESCE(p.last_modified, p.created_at) DESC
	`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		var lastModified sql.NullTime
		var imageURL sql.NullString
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &imageURL, &post.CreatedAt, &lastModified, &post.LikeCount, &post.DislikeCount, &post.CommentCount)
		if err != nil {
			return nil, err
		}
		if lastModified.Valid {
			post.LastModified.Time = lastModified.Time
			post.LastModified.Valid = true
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}

		// Fetch categories (N+1 problem, but acceptable for small scale)
		catQuery := `SELECT category_id FROM post_categories WHERE post_id = ?`
		catRows, err := m.DB.Query(catQuery, post.ID)
		if err == nil {
			for catRows.Next() {
				var catID int
				catRows.Scan(&catID)
				post.Categories = append(post.Categories, catID)
			}
			catRows.Close()
		}

		posts = append(posts, &post)
	}
	return posts, nil
}

func (m *PostsModel) UpdateLastModified(postID int) error {
	statement := `UPDATE posts SET last_modified = datetime('now') WHERE id = ?`
	_, err := m.DB.Exec(statement, postID)
	return err
}

func (m *PostsModel) ListByAuthor(authorID string) ([]*Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.author_id, u.username, p.image_url, p.created_at, p.last_modified,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = 1) as likes,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = -1) as dislikes,
		(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comments
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.author_id = ?
		ORDER BY p.created_at DESC
	`
	rows, err := m.DB.Query(query, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		var lastModified sql.NullTime
		var imageURL sql.NullString
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &imageURL, &post.CreatedAt, &lastModified, &post.LikeCount, &post.DislikeCount, &post.CommentCount)
		if err != nil {
			return nil, err
		}
		if lastModified.Valid {
			post.LastModified.Time = lastModified.Time
			post.LastModified.Valid = true
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}

		// Fetch categories
		catQuery := `SELECT category_id FROM post_categories WHERE post_id = ?`
		catRows, err := m.DB.Query(catQuery, post.ID)
		if err == nil {
			for catRows.Next() {
				var catID int
				catRows.Scan(&catID)
				post.Categories = append(post.Categories, catID)
			}
			catRows.Close()
		}

		posts = append(posts, &post)
	}
	return posts, nil
}

func (m *PostsModel) ListLikedByUser(userID string) ([]*Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.author_id, u.username, p.image_url, p.created_at, p.last_modified,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = 1) as likes,
		(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = -1) as dislikes,
		(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comments
		FROM posts p
		JOIN users u ON p.author_id = u.id
		JOIN reactions r ON p.id = r.post_id
		WHERE r.user_id = ? AND r.type = 1
		ORDER BY r.id DESC
	`
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		var lastModified sql.NullTime
		var imageURL sql.NullString
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &imageURL, &post.CreatedAt, &lastModified, &post.LikeCount, &post.DislikeCount, &post.CommentCount)
		if err != nil {
			return nil, err
		}
		if lastModified.Valid {
			post.LastModified.Time = lastModified.Time
			post.LastModified.Valid = true
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}

		// Fetch categories
		catQuery := `SELECT category_id FROM post_categories WHERE post_id = ?`
		catRows, err := m.DB.Query(catQuery, post.ID)
		if err == nil {
			for catRows.Next() {
				var catID int
				catRows.Scan(&catID)
				post.Categories = append(post.Categories, catID)
			}
			catRows.Close()
		}

		posts = append(posts, &post)
	}
	return posts, nil
}

// UpdatePostDB updates a post in the database
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

// DeletePostDB deletes a post from the database
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

func (m *PostsModel) ListRandom(limit int) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.image_url, p.created_at, p.last_modified,
	(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = 1) as likes,
	(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = -1) as dislikes,
	(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comments
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
		var post Post
		var lastModified sql.NullTime
		var imageURL sql.NullString
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &imageURL, &post.CreatedAt, &lastModified, &post.LikeCount, &post.DislikeCount, &post.CommentCount)
		if err != nil {
			return nil, err
		}
		if lastModified.Valid {
			post.LastModified.Time = lastModified.Time
			post.LastModified.Valid = true
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}

		categories, err := m.getCategoriesForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (m *PostsModel) ListByCategory(categoryID int, limit int) ([]*Post, error) {
	statement := `SELECT p.id, p.title, p.content, p.author_id, u.username, p.image_url, p.created_at, p.last_modified,
	(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = 1) as likes,
	(SELECT COUNT(*) FROM reactions WHERE post_id = p.id AND type = -1) as dislikes,
	(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comments
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
		var post Post
		var lastModified sql.NullTime
		var imageURL sql.NullString
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &imageURL,
			&post.CreatedAt, &lastModified, &post.LikeCount, &post.DislikeCount, &post.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		if lastModified.Valid {
			post.LastModified.Time = lastModified.Time
			post.LastModified.Valid = true
		}
		if imageURL.Valid {
			post.ImageURL = imageURL.String
		}

		categories, err := m.getCategoriesForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, &post)
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
