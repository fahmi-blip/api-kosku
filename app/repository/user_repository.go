package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kosku-api/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

const userColumns = "id, name, email, password_hash, phone, role, created_at"

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Phone,
		&u.Role,
		&u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userPostgresRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (name, email, password_hash, phone, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("menyimpan pengguna: %w", err)
	}
	return nil
}

func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := "SELECT " + userColumns + " FROM users WHERE email = $1"
	u, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("mengambil pengguna berdasarkan email: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	query := "SELECT " + userColumns + " FROM users WHERE id = $1"
	u, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("mengambil pengguna berdasarkan id: %w", err)
	}
	return u, nil
}
