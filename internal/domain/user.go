package domain

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)
