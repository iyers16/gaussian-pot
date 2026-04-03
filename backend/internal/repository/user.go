package repository

import (
	"database/sql"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, username, role, credits, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Role, &u.Credits, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) Create(username string, role model.Role) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`INSERT INTO users (username, role, credits) VALUES ($1, $2, 1000.00)
		 RETURNING id, username, role, credits, created_at`,
		username, string(role),
	).Scan(&u.ID, &u.Username, &u.Role, &u.Credits, &u.CreatedAt)
	return u, err
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.Query(
		`SELECT id, username, role, credits, created_at FROM users ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Credits, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) UpdateCredits(userID int, credits float64) error {
	_, err := r.db.Exec(
		`UPDATE users SET credits = $1 WHERE id = $2`,
		credits, userID,
	)
	return err
}

func (r *UserRepository) AddCredits(userID int, delta float64) error {
	_, err := r.db.Exec(
		`UPDATE users SET credits = credits + $1 WHERE id = $2`,
		delta, userID,
	)
	return err
}

func (r *UserRepository) GetByID(userID int) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, username, role, credits, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Username, &u.Role, &u.Credits, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}
