CREATE TABLE IF NOT EXISTS rentals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    room_id BIGINT NOT NULL,
    start_date DATE NOT NULL,
    duration_months INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_rentals_user
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_rentals_room
        FOREIGN KEY (room_id) REFERENCES rooms(id)
);
