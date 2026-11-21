package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kosku-api/app/model"
)

/* ------------------------------------------------------------------ */
/*  Contract                                                          */
/* ------------------------------------------------------------------ */

type TokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	RevokeByHash(ctx context.Context, hash string) error
	RevokeAllByUser(ctx context.Context, userID uint) error
}

/* ------------------------------------------------------------------ */
/*  Implementation                                                    */
/* ------------------------------------------------------------------ */

type tokenPostgresRepository struct {
	pool *pgxpool.Pool
}

const tokenColumns = "id, user_id, token_hash, expires_at, revoked_at, created_at"

func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenPostgresRepository{pool: pool}
}

func scanToken(row pgx.Row) (*model.RefreshToken, error) {
	var t model.RefreshToken
	err := row.Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.ExpiresAt,
		&t.RevokedAt,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tokenPostgresRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("menyimpan refresh token: %w", err)
	}
	return nil
}

func (r *tokenPostgresRepository) FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	query := "SELECT " + tokenColumns + " FROM refresh_tokens WHERE token_hash = $1"
	t, err := scanToken(r.pool.QueryRow(ctx, query, hash))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("mengambil refresh token: %w", err)
	}
	return t, nil
}

func (r *tokenPostgresRepository) RevokeByHash(ctx context.Context, hash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, hash)
	if err != nil {
		return fmt.Errorf("mencabut refresh token: %w", err)
	}
	return nil
}

func (r *tokenPostgresRepository) RevokeAllByUser(ctx context.Context, userID uint) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("mencabut seluruh refresh token user: %w", err)
	}
	return nil
}
