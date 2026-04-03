package model

import "time"

type Role string

const (
	RoleHost   Role = "host"
	RolePlayer Role = "player"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	Credits   float64   `json:"credits"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	Token    string
	UserID   int
	Username string
	Role     Role
}
