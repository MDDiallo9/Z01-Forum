package models

import (
	"database/sql"
	"errors"

	/* "errors"
	"time" */
	"github.com/mattn/go-sqlite3"
)

type PostsModel struct {
	DB *sql.DB
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
