package domain

import "time"

const (
	CodePurposeEmailVerification = "email_verification"
	CodePurposeLoginOTP          = "login_otp"
)

type Code struct {
	ID        int64
	Code      string
	ExpiredAt time.Time
	Purpose   string
	Receiver  string
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}
