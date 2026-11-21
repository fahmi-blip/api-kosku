CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_refresh_tokens_user
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT uni_refresh_tokens_token_hash
        UNIQUE (token_hash)
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id
    ON refresh_tokens(user_id);

UPDATE refresh_tokens
SET created_at = COALESCE(created_at, CURRENT_TIMESTAMP)
WHERE created_at IS NULL;

ALTER TABLE refresh_tokens
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN created_at SET NOT NULL;
    