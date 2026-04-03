package repository

import (
	"database/sql"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

type NumberRepository struct {
	db *sql.DB
}

func NewNumberRepository(db *sql.DB) *NumberRepository {
	return &NumberRepository{db: db}
}

func (r *NumberRepository) Save(value int) (*model.Number, error) {
	number := &model.Number{}
	err := r.db.QueryRow(
		"INSERT INTO numbers (value) VALUES ($1) RETURNING id, value, created_at",
		value,
	).Scan(&number.ID, &number.Value, &number.CreatedAt)
	return number, err
}

func (r *NumberRepository) GetAll() ([]model.Number, error) {
	rows, err := r.db.Query("SELECT id, value, created_at FROM numbers ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var numbers []model.Number
	for rows.Next() {
		var n model.Number
		rows.Scan(&n.ID, &n.Value, &n.CreatedAt)
		numbers = append(numbers, n)
	}
	return numbers, nil
}
