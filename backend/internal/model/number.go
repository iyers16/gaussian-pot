package model

import "time"

type Number struct {
	ID		int	`json:"id"`
	Value		int	`json:"value"`
	CreatedAt	time.Time	`json:"created_at"`
}
