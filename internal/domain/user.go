package domain

import "time"

type User struct {
	ID           int64
	Email        string
	FirstName    string
	LastName     string
	MiddleName   *string
	PhoneNumber  *string
	PasswordHash string
	Role         Role
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role string

const (
	RoleStudent Role = "student"
	RoleCurator Role = "curator"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)
