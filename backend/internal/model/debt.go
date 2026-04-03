package model

import "time"

type DebtStatus string

const (
	DebtPending DebtStatus = "PENDING"
	DebtSettled DebtStatus = "SETTLED"
)

type Debt struct {
	ID             int        `json:"id"`
	RoundID        int        `json:"round_id"`
	PayerID        int        `json:"payer_id"`
	PayerUsername  string     `json:"payer_username"`
	PayeeID        int        `json:"payee_id"`
	PayeeUsername  string     `json:"payee_username"`
	Amount         float64    `json:"amount"`
	PayerConfirmed bool       `json:"payer_confirmed"`
	PayeeConfirmed bool       `json:"payee_confirmed"`
	Status         DebtStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}
