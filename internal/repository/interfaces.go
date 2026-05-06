package repository

import (
	"context"

	"github.com/Talan-Application/auth-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	SetVerified(ctx context.Context, email string) error
}

type CodeRepository interface {
	Create(ctx context.Context, code *domain.Code) error
	FindValid(ctx context.Context, receiver, code, purpose string) (*domain.Code, error)
	MarkUsed(ctx context.Context, id int64) error
}
