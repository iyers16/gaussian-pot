package repository

import (
	"database/sql"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

type RoundRepository struct {
	db *sql.DB
}

func NewRoundRepository(db *sql.DB) *RoundRepository {
	return &RoundRepository{db: db}
}

func (r *RoundRepository) Create(questionID int, questionText string, targetValue float64, unit string, mode model.BettingMode) (*model.Round, error) {
	round := &model.Round{}
	err := r.db.QueryRow(
		`INSERT INTO rounds (question_id, question_text, target_value, unit, mode, state)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, question_id, question_text, target_value, unit, mode, state, opened_at`,
		questionID, questionText, targetValue, unit, string(mode), string(model.StateRoundOpen),
	).Scan(&round.ID, &round.QuestionID, &round.QuestionText, &round.TargetValue,
		&round.Unit, &round.Mode, &round.State, &round.OpenedAt)
	return round, err
}

func (r *RoundRepository) GetCurrent() (*model.Round, error) {
	round := &model.Round{}
	err := r.db.QueryRow(
		`SELECT id, question_id, question_text, target_value, unit, mode, state, opened_at, called_at
		 FROM rounds
		 WHERE state != $1
		 ORDER BY opened_at DESC LIMIT 1`,
		string(model.StateIdle),
	).Scan(&round.ID, &round.QuestionID, &round.QuestionText, &round.TargetValue,
		&round.Unit, &round.Mode, &round.State, &round.OpenedAt, &round.CalledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return round, err
}

func (r *RoundRepository) UpdateState(roundID int, state model.GameState) error {
	if state == model.StateRoundCalled || state == model.StateSettling {
		_, err := r.db.Exec(
			`UPDATE rounds SET state = $1, called_at = NOW() WHERE id = $2`,
			string(state), roundID,
		)
		return err
	}
	_, err := r.db.Exec(
		`UPDATE rounds SET state = $1 WHERE id = $2`,
		string(state), roundID,
	)
	return err
}

func (r *RoundRepository) GetByID(roundID int) (*model.Round, error) {
	round := &model.Round{}
	err := r.db.QueryRow(
		`SELECT id, question_id, question_text, target_value, unit, mode, state, opened_at, called_at
		 FROM rounds WHERE id = $1`,
		roundID,
	).Scan(&round.ID, &round.QuestionID, &round.QuestionText, &round.TargetValue,
		&round.Unit, &round.Mode, &round.State, &round.OpenedAt, &round.CalledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return round, err
}
