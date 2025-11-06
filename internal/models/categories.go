package models

import "database/sql"

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CategoriesModel struct {
	DB *sql.DB
}

func (m *CategoriesModel) ListAll() ([]*Category, error) {
	statement := `SELECT id, name FROM categories ORDER BY ASC`

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}
