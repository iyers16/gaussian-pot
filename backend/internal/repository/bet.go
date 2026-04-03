package repository

import (
	"database/sql"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

type BetRepository struct {
	db *sql.DB
}

func NewBetRepository(db *sql.DB) *BetRepository {
	return &BetRepository{db: db}
}

func (r *BetRepository) Create(roundID, userID int, username string, guess, wager float64) (*model.Bet, error) {
	bet := &model.Bet{}
	err := r.db.QueryRow(
		`INSERT INTO bets (round_id, user_id, username, guess, wager)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, round_id, user_id, username, guess, wager, payout, placed_at`,
		roundID, userID, username, guess, wager,
	).Scan(&bet.ID, &bet.RoundID, &bet.UserID, &bet.Username,
		&bet.Guess, &bet.Wager, &bet.Payout, &bet.PlacedAt)
	return bet, err
}

func (r *BetRepository) GetByRound(roundID int) ([]model.Bet, error) {
	rows, err := r.db.Query(
		`SELECT id, round_id, user_id, username, guess, wager, payout, placed_at
		 FROM bets WHERE round_id = $1 ORDER BY placed_at ASC`,
		roundID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bets []model.Bet
	for rows.Next() {
		var b model.Bet
		if err := rows.Scan(&b.ID, &b.RoundID, &b.UserID, &b.Username,
			&b.Guess, &b.Wager, &b.Payout, &b.PlacedAt); err != nil {
			return nil, err
		}
		bets = append(bets, b)
	}
	return bets, nil
}

func (r *BetRepository) GetByUserAndRound(userID, roundID int) (*model.Bet, error) {
	bet := &model.Bet{}
	err := r.db.QueryRow(
		`SELECT id, round_id, user_id, username, guess, wager, payout, placed_at
		 FROM bets WHERE user_id = $1 AND round_id = $2`,
		userID, roundID,
	).Scan(&bet.ID, &bet.RoundID, &bet.UserID, &bet.Username,
		&bet.Guess, &bet.Wager, &bet.Payout, &bet.PlacedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return bet, err
}

func (r *BetRepository) UpdatePayout(betID int, payout float64) error {
	_, err := r.db.Exec(
		`UPDATE bets SET payout = $1 WHERE id = $2`,
		payout, betID,
	)
	return err
}
