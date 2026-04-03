package model

import "time"

type Bet struct {
	ID       int       `json:"id"`
	RoundID  int       `json:"round_id"`
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	Guess    float64   `json:"guess"`
	Wager    float64   `json:"wager"`
	Payout   float64   `json:"payout"`
	PlacedAt time.Time `json:"placed_at"`
}
