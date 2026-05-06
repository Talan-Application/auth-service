package postgres

import (
	"context"

	"github.com/Talan-Application/auth-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CodeRepository struct {
	pool *pgxpool.Pool
}

func NewCodeRepository(pool *pgxpool.Pool) *CodeRepository {
	return &CodeRepository{pool: pool}
}

func (r *CodeRepository) Create(ctx context.Context, code *domain.Code) error {
	query := `INSERT INTO codes()`
}
