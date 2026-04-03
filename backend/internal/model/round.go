package model

import "time"

type GameState string

const (
	StateIdle        GameState = "IDLE"
	StateRoundOpen   GameState = "ROUND_OPEN"
	StateRoundCalled GameState = "ROUND_CALLED"
	StateSettling    GameState = "SETTLING"
)

type BettingMode string

const (
	ModeSniper   BettingMode = "sniper"
	ModeSocial   BettingMode = "social"
	ModeVolatile BettingMode = "volatile"
)

type Round struct {
	ID           int         `json:"id"`
	QuestionID   int         `json:"question_id"`
	QuestionText string      `json:"question_text"`
	TargetValue  float64     `json:"target_value"`
	Unit         string      `json:"unit"`
	Mode         BettingMode `json:"mode"`
	State        GameState   `json:"state"`
	OpenedAt     time.Time   `json:"opened_at"`
	CalledAt     *time.Time  `json:"called_at,omitempty"`
}
