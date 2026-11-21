package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kosku-api/app/model"
)

type RoomRepository interface {
	FindAll(ctx context.Context) ([]model.Room, error)
	FindAvailable(ctx context.Context) ([]model.Room, error)
	FindByID(ctx context.Context, id uint) (*model.Room, error)
	FindByCode(ctx context.Context, code string) (*model.Room, error)
	Create(ctx context.Context, room *model.Room) error
	Update(ctx context.Context, room *model.Room) error
	Delete(ctx context.Context, id uint) error
	UpdateStatus(ctx context.Context, id uint, status string) error
}

type roomPostgresRepository struct {
	pool *pgxpool.Pool
}

const roomColumns = "id, code, name, price_per_month, facilities, status, created_at"

func NewRoomRepository(pool *pgxpool.Pool) RoomRepository {
	return &roomPostgresRepository{pool: pool}
}

func scanRoom(row pgx.Row) (*model.Room, error) {
	var r model.Room
	err := row.Scan(
		&r.ID,
		&r.Code,
		&r.Name,
		&r.PricePerMonth,
		&r.Facilities,
		&r.Status,
		&r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *roomPostgresRepository) collectRooms(rows pgx.Rows, err error) ([]model.Room, error) {
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar kamar: %w", err)
	}
	defer rows.Close()

	result := []model.Room{}
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row kamar: %w", err)
		}
		result = append(result, *room)
	}
	return result, rows.Err()
}

func (r *roomPostgresRepository) FindAll(ctx context.Context) ([]model.Room, error) {
	return r.collectRooms(r.pool.Query(ctx, "SELECT "+roomColumns+" FROM rooms ORDER BY id DESC"))
}

func (r *roomPostgresRepository) FindAvailable(ctx context.Context) ([]model.Room, error) {
	return r.collectRooms(r.pool.Query(ctx, "SELECT "+roomColumns+" FROM rooms WHERE status = $1 ORDER BY id DESC", model.RoomAvailable))
}

func (r *roomPostgresRepository) FindByID(ctx context.Context, id uint) (*model.Room, error) {
	query := "SELECT " + roomColumns + " FROM rooms WHERE id = $1"
	room, err := scanRoom(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("kamar tidak ditemukan")
		}
		return nil, fmt.Errorf("mengambil kamar: %w", err)
	}
	return room, nil
}

func (r *roomPostgresRepository) FindByCode(ctx context.Context, code string) (*model.Room, error) {
	query := "SELECT " + roomColumns + " FROM rooms WHERE code = $1"
	room, err := scanRoom(r.pool.QueryRow(ctx, query, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("mengambil kamar berdasarkan kode: %w", err)
	}
	return room, nil
}

func (r *roomPostgresRepository) Create(ctx context.Context, room *model.Room) error {
	query := `
		INSERT INTO rooms (code, name, price_per_month, facilities, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		room.Code,
		room.Name,
		room.PricePerMonth,
		room.Facilities,
		room.Status,
	).Scan(&room.ID, &room.CreatedAt)
	if err != nil {
		return fmt.Errorf("menyimpan kamar: %w", err)
	}
	return nil
}

func (r *roomPostgresRepository) Update(ctx context.Context, room *model.Room) error {
	query := `
		UPDATE rooms
		SET code = $1, name = $2, price_per_month = $3, facilities = $4, status = $5
		WHERE id = $6
	`
	tag, err := r.pool.Exec(
		ctx,
		query,
		room.Code,
		room.Name,
		room.PricePerMonth,
		room.Facilities,
		room.Status,
		room.ID,
	)
	if err != nil {
		return fmt.Errorf("memperbarui kamar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("kamar tidak ditemukan")
	}
	return nil
}

func (r *roomPostgresRepository) Delete(ctx context.Context, id uint) error {
	query := "DELETE FROM rooms WHERE id = $1"
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("menghapus kamar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("kamar tidak ditemukan")
	}
	return nil
}

func (r *roomPostgresRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	query := "UPDATE rooms SET status = $1 WHERE id = $2"
	tag, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("memperbarui status kamar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("kamar tidak ditemukan")
	}
	return nil
}
