package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Talan-Application/auth-service/internal/domain"
)

type CodeRepository struct {
	pool *pgxpool.Pool
}

func NewCodeRepository(pool *pgxpool.Pool) *CodeRepository {
	return &CodeRepository{pool: pool}
}

func (r *CodeRepository) Create(ctx context.Context, code *domain.Code) error {
	query := `
		INSERT INTO codes (receiver, code, purpose, expired_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id`

	err := r.pool.QueryRow(ctx, query, code.Receiver, code.Code, code.Purpose, code.ExpiredAt).Scan(&code.ID)
	if err != nil {
		return fmt.Errorf("create code: %w", err)
	}
	return nil
}

func (r *CodeRepository) FindValid(ctx context.Context, receiver, code, purpose string) (*domain.Code, error) {
	query := `
		SELECT id, code, receiver, purpose, expired_at, created_at
		FROM codes
		WHERE receiver = $1
		  AND code = $2
		  AND purpose = $3
		  AND deleted_at IS NULL
		  AND expired_at > $4
		ORDER BY created_at DESC
		LIMIT 1`

	c := &domain.Code{}
	err := r.pool.QueryRow(ctx, query, receiver, code, purpose, time.Now()).
		Scan(&c.ID, &c.Code, &c.Receiver, &c.Purpose, &c.ExpiredAt, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCode
		}
		return nil, fmt.Errorf("find valid code: %w", err)
	}
	return c, nil
}

func (r *CodeRepository) MarkUsed(ctx context.Context, id int64) error {
	query := `UPDATE codes SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark code used: %w", err)
	}
	return nil
}
