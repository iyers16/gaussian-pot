package repository

import (
	"database/sql"

	"github.com/iyers16/gaussian-pot/backend/internal/model"
)

type DebtRepository struct {
	db *sql.DB
}

func NewDebtRepository(db *sql.DB) *DebtRepository {
	return &DebtRepository{db: db}
}

func (r *DebtRepository) Create(roundID, payerID, payeeID int, payerUsername, payeeUsername string, amount float64) (*model.Debt, error) {
	d := &model.Debt{}
	err := r.db.QueryRow(
		`INSERT INTO debts (round_id, payer_id, payee_id, payer_username, payee_username, amount)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, round_id, payer_id, payer_username, payee_id, payee_username,
		           amount, payer_confirmed, payee_confirmed, status, created_at`,
		roundID, payerID, payeeID, payerUsername, payeeUsername, amount,
	).Scan(&d.ID, &d.RoundID, &d.PayerID, &d.PayerUsername, &d.PayeeID, &d.PayeeUsername,
		&d.Amount, &d.PayerConfirmed, &d.PayeeConfirmed, &d.Status, &d.CreatedAt)
	return d, err
}

func (r *DebtRepository) GetByRound(roundID int) ([]model.Debt, error) {
	rows, err := r.db.Query(
		`SELECT id, round_id, payer_id, payer_username, payee_id, payee_username,
		        amount, payer_confirmed, payee_confirmed, status, created_at
		 FROM debts WHERE round_id = $1`,
		roundID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanDebts(rows)
}

func (r *DebtRepository) GetByUser(userID int) ([]model.Debt, error) {
	rows, err := r.db.Query(
		`SELECT id, round_id, payer_id, payer_username, payee_id, payee_username,
		        amount, payer_confirmed, payee_confirmed, status, created_at
		 FROM debts WHERE (payer_id = $1 OR payee_id = $1) AND status = 'PENDING'
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanDebts(rows)
}

func (r *DebtRepository) GetByID(debtID int) (*model.Debt, error) {
	d := &model.Debt{}
	err := r.db.QueryRow(
		`SELECT id, round_id, payer_id, payer_username, payee_id, payee_username,
		        amount, payer_confirmed, payee_confirmed, status, created_at
		 FROM debts WHERE id = $1`,
		debtID,
	).Scan(&d.ID, &d.RoundID, &d.PayerID, &d.PayerUsername, &d.PayeeID, &d.PayeeUsername,
		&d.Amount, &d.PayerConfirmed, &d.PayeeConfirmed, &d.Status, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

func (r *DebtRepository) ConfirmPaid(debtID int) (*model.Debt, error) {
	d := &model.Debt{}
	err := r.db.QueryRow(
		`UPDATE debts
		 SET payer_confirmed = TRUE,
		     status = CASE WHEN payee_confirmed THEN 'SETTLED' ELSE status END
		 WHERE id = $1
		 RETURNING id, round_id, payer_id, payer_username, payee_id, payee_username,
		           amount, payer_confirmed, payee_confirmed, status, created_at`,
		debtID,
	).Scan(&d.ID, &d.RoundID, &d.PayerID, &d.PayerUsername, &d.PayeeID, &d.PayeeUsername,
		&d.Amount, &d.PayerConfirmed, &d.PayeeConfirmed, &d.Status, &d.CreatedAt)
	return d, err
}

func (r *DebtRepository) ConfirmReceived(debtID int) (*model.Debt, error) {
	d := &model.Debt{}
	err := r.db.QueryRow(
		`UPDATE debts
		 SET payee_confirmed = TRUE,
		     status = CASE WHEN payer_confirmed THEN 'SETTLED' ELSE status END
		 WHERE id = $1
		 RETURNING id, round_id, payer_id, payer_username, payee_id, payee_username,
		           amount, payer_confirmed, payee_confirmed, status, created_at`,
		debtID,
	).Scan(&d.ID, &d.RoundID, &d.PayerID, &d.PayerUsername, &d.PayeeID, &d.PayeeUsername,
		&d.Amount, &d.PayerConfirmed, &d.PayeeConfirmed, &d.Status, &d.CreatedAt)
	return d, err
}

func (r *DebtRepository) HasPendingDebts(userID int) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM debts
		 WHERE (payer_id = $1 OR payee_id = $1) AND status = 'PENDING'`,
		userID,
	).Scan(&count)
	return count > 0, err
}

func (r *DebtRepository) HasAnyPending() (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM debts WHERE status = 'PENDING'`).Scan(&count)
	return count > 0, err
}

func (r *DebtRepository) scanDebts(rows *sql.Rows) ([]model.Debt, error) {
	var debts []model.Debt
	for rows.Next() {
		var d model.Debt
		if err := rows.Scan(&d.ID, &d.RoundID, &d.PayerID, &d.PayerUsername, &d.PayeeID, &d.PayeeUsername,
			&d.Amount, &d.PayerConfirmed, &d.PayeeConfirmed, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}
	return debts, nil
}
