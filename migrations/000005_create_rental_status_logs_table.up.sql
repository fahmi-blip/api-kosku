CREATE TABLE IF NOT EXISTS rental_status_logs (
    id BIGSERIAL PRIMARY KEY,
    rental_id BIGINT NOT NULL,
    actor_id BIGINT NOT NULL,
    from_status VARCHAR(20) NOT NULL,
    to_status VARCHAR(20) NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_rental_status_logs_rental
        FOREIGN KEY (rental_id) REFERENCES rentals(id),
    CONSTRAINT fk_rental_status_logs_actor
        FOREIGN KEY (actor_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_rental_status_logs_rental_id
    ON rental_status_logs(rental_id);

CREATE INDEX IF NOT EXISTS idx_rental_status_logs_actor_id
    ON rental_status_logs(actor_id);
