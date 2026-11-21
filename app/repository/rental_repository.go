package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kosku-api/app/model"
)

type RentalRepository interface {
	Create(ctx context.Context, rental *model.Rental) error
	CreateWithLog(ctx context.Context, rental *model.Rental, logEntry *model.RentalStatusLog) error
	FindByID(ctx context.Context, id uint) (*model.Rental, error)
	FindAll(ctx context.Context) ([]model.Rental, error)
	FindByUserID(ctx context.Context, userID uint) ([]model.Rental, error)
	FindLogsByRentalID(ctx context.Context, rentalID uint) ([]model.RentalStatusLog, error)
	ExistsApprovedByRoom(ctx context.Context, roomID uint) (bool, error)
	ExistsPendingByUserAndRoom(ctx context.Context, userID, roomID uint) (bool, error)
	ApproveTx(ctx context.Context, rentalID, roomID, actorID uint) error
	RejectTx(ctx context.Context, rentalID, actorID uint) error
	CancelTx(ctx context.Context, rentalID, roomID, actorID uint) error
}

type rentalPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRentalRepository(pool *pgxpool.Pool) RentalRepository {
	return &rentalPostgresRepository{pool: pool}
}

const rentalJoinSelect = `
	SELECT 
		r.id, r.user_id, r.room_id, r.start_date, r.duration_months, r.status, r.created_at,
		u.id, u.name, u.email, u.phone, u.role,
		rm.id, rm.code, rm.name, rm.price_per_month, rm.facilities, rm.status, rm.created_at
	FROM rentals r
	JOIN users u ON u.id = r.user_id
	JOIN rooms rm ON rm.id = r.room_id
`

func scanRentalWithJoin(row pgx.Row) (*model.Rental, error) {
	var r model.Rental
	var u model.User
	var rm model.Room
	err := row.Scan(
		&r.ID,
		&r.UserID,
		&r.RoomID,
		&r.StartDate,
		&r.DurationMonths,
		&r.Status,
		&r.CreatedAt,
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Phone,
		&u.Role,
		&rm.ID,
		&rm.Code,
		&rm.Name,
		&rm.PricePerMonth,
		&rm.Facilities,
		&rm.Status,
		&rm.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	r.User = u
	r.Room = rm
	return &r, nil
}

func (r *rentalPostgresRepository) collectRentals(rows pgx.Rows, err error) ([]model.Rental, error) {
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar sewa: %w", err)
	}
	defer rows.Close()

	result := []model.Rental{}
	for rows.Next() {
		rental, err := scanRentalWithJoin(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row sewa: %w", err)
		}
		result = append(result, *rental)
	}
	return result, rows.Err()
}

func (r *rentalPostgresRepository) Create(ctx context.Context, rental *model.Rental) error {
	query := `
		INSERT INTO rentals (user_id, room_id, start_date, duration_months, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		rental.UserID,
		rental.RoomID,
		rental.StartDate,
		rental.DurationMonths,
		rental.Status,
	).Scan(&rental.ID, &rental.CreatedAt)
	if err != nil {
		return fmt.Errorf("menyimpan sewa: %w", err)
	}
	return nil
}

func (r *rentalPostgresRepository) CreateWithLog(ctx context.Context, rental *model.Rental, logEntry *model.RentalStatusLog) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		`INSERT INTO rentals (user_id, room_id, start_date, duration_months, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		rental.UserID,
		rental.RoomID,
		rental.StartDate,
		rental.DurationMonths,
		rental.Status,
	).Scan(&rental.ID, &rental.CreatedAt)
	if err != nil {
		return fmt.Errorf("menyimpan sewa: %w", err)
	}

	logEntry.RentalID = rental.ID
	err = tx.QueryRow(
		ctx,
		`INSERT INTO rental_status_logs (rental_id, actor_id, from_status, to_status, note)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		logEntry.RentalID,
		logEntry.ActorID,
		logEntry.FromStatus,
		logEntry.ToStatus,
		logEntry.Note,
	).Scan(&logEntry.ID, &logEntry.CreatedAt)
	if err != nil {
		return fmt.Errorf("menyimpan log sewa: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *rentalPostgresRepository) FindByID(ctx context.Context, id uint) (*model.Rental, error) {
	query := rentalJoinSelect + " WHERE r.id = $1"
	rental, err := scanRentalWithJoin(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("data sewa tidak ditemukan")
		}
		return nil, fmt.Errorf("mengambil sewa: %w", err)
	}
	return rental, nil
}

func (r *rentalPostgresRepository) FindAll(ctx context.Context) ([]model.Rental, error) {
	query := rentalJoinSelect + " ORDER BY r.id DESC"
	return r.collectRentals(r.pool.Query(ctx, query))
}

func (r *rentalPostgresRepository) FindByUserID(ctx context.Context, userID uint) ([]model.Rental, error) {
	query := rentalJoinSelect + " WHERE r.user_id = $1 ORDER BY r.id DESC"
	return r.collectRentals(r.pool.Query(ctx, query, userID))
}

func (r *rentalPostgresRepository) FindLogsByRentalID(ctx context.Context, rentalID uint) ([]model.RentalStatusLog, error) {
	query := `
		SELECT l.id, l.rental_id, l.actor_id, l.from_status, l.to_status, COALESCE(l.note, ''), l.created_at,
		       COALESCE(u.id, 0), COALESCE(u.name, ''), COALESCE(u.email, ''), COALESCE(u.phone, ''), COALESCE(u.role, '')
		FROM rental_status_logs l
		LEFT JOIN users u ON u.id = l.actor_id
		WHERE l.rental_id = $1
		ORDER BY l.id ASC
	`
	rows, err := r.pool.Query(ctx, query, rentalID)
	if err != nil {
		return nil, fmt.Errorf("mengambil log sewa: %w", err)
	}
	defer rows.Close()

	var logs []model.RentalStatusLog
	for rows.Next() {
		var l model.RentalStatusLog
		var u model.User
		err := rows.Scan(
			&l.ID,
			&l.RentalID,
			&l.ActorID,
			&l.FromStatus,
			&l.ToStatus,
			&l.Note,
			&l.CreatedAt,
			&u.ID,
			&u.Name,
			&u.Email,
			&u.Phone,
			&u.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("membaca row log sewa: %w", err)
		}
		if u.ID != 0 {
			l.Actor = u
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (r *rentalPostgresRepository) ExistsApprovedByRoom(ctx context.Context, roomID uint) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM rentals WHERE room_id = $1 AND status = $2)"
	err := r.pool.QueryRow(ctx, query, roomID, model.StatusApproved).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("memeriksa sewa disetujui: %w", err)
	}
	return exists, nil
}

func (r *rentalPostgresRepository) ExistsPendingByUserAndRoom(ctx context.Context, userID, roomID uint) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM rentals WHERE user_id = $1 AND room_id = $2 AND status = $3)"
	err := r.pool.QueryRow(ctx, query, userID, roomID, model.StatusPending).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("memeriksa sewa pending: %w", err)
	}
	return exists, nil
}

func (r *rentalPostgresRepository) ApproveTx(ctx context.Context, rentalID, roomID, actorID uint) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(
		ctx,
		"UPDATE rentals SET status = $1 WHERE id = $2 AND room_id = $3 AND status = $4",
		model.StatusApproved,
		rentalID,
		roomID,
		model.StatusPending,
	)
	if err != nil {
		return fmt.Errorf("memperbarui status sewa: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("pengajuan sewa tidak valid atau tidak berstatus pending")
	}

	roomTag, err := tx.Exec(
		ctx,
		"UPDATE rooms SET status = $1 WHERE id = $2 AND status = $3",
		model.RoomOccupied,
		roomID,
		model.RoomAvailable,
	)
	if err != nil {
		return fmt.Errorf("memperbarui status kamar: %w", err)
	}
	if roomTag.RowsAffected() != 1 {
		return errors.New("kamar tidak tersedia")
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO rental_status_logs (rental_id, actor_id, from_status, to_status, note)
		 VALUES ($1, $2, $3, $4, $5)`,
		rentalID,
		actorID,
		model.StatusPending,
		model.StatusApproved,
		"Pengajuan disetujui",
	)
	if err != nil {
		return fmt.Errorf("mencatat log status: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *rentalPostgresRepository) RejectTx(ctx context.Context, rentalID, actorID uint) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(
		ctx,
		"UPDATE rentals SET status = $1 WHERE id = $2 AND status = $3",
		model.StatusRejected,
		rentalID,
		model.StatusPending,
	)
	if err != nil {
		return fmt.Errorf("memperbarui status sewa: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("pengajuan sewa tidak valid atau tidak berstatus pending")
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO rental_status_logs (rental_id, actor_id, from_status, to_status, note)
		 VALUES ($1, $2, $3, $4, $5)`,
		rentalID,
		actorID,
		model.StatusPending,
		model.StatusRejected,
		"Pengajuan ditolak",
	)
	if err != nil {
		return fmt.Errorf("mencatat log status: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *rentalPostgresRepository) CancelTx(ctx context.Context, rentalID, roomID, actorID uint) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(
		ctx,
		"SELECT status FROM rentals WHERE id = $1 AND room_id = $2 FOR UPDATE",
		rentalID,
		roomID,
	).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("data sewa tidak ditemukan")
		}
		return fmt.Errorf("mengambil status sewa: %w", err)
	}

	if currentStatus != model.StatusPending && currentStatus != model.StatusApproved {
		return errors.New("sewa tidak dalam status yang dapat dibatalkan")
	}

	tag, err := tx.Exec(ctx, "UPDATE rentals SET status = $1 WHERE id = $2", model.StatusCancelled, rentalID)
	if err != nil || tag.RowsAffected() != 1 {
		return fmt.Errorf("memperbarui status sewa: %w", err)
	}

	if currentStatus == model.StatusApproved {
		roomTag, err := tx.Exec(ctx, "UPDATE rooms SET status = $1 WHERE id = $2 AND status = $3", model.RoomAvailable, roomID, model.RoomOccupied)
		if err != nil || roomTag.RowsAffected() != 1 {
			return fmt.Errorf("mengembalikan status kamar: %w", err)
		}
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO rental_status_logs (rental_id, actor_id, from_status, to_status, note)
		 VALUES ($1, $2, $3, $4, $5)`,
		rentalID,
		actorID,
		currentStatus,
		model.StatusCancelled,
		"Sewa dibatalkan oleh penghuni",
	)
	if err != nil {
		return fmt.Errorf("mencatat log status: %w", err)
	}

	return tx.Commit(ctx)
}
