package service

import "context"

type AuthService interface {
	Register(ctx context.Context, email, password string) (*TokenPair, error)
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	ValidateToken(ctx context.Context, accessToken string) (*Claims, error)
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	UserID int64
	Email  string
	Role   string
}
