DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'uni_refresh_tokens_token_hash'
          AND conrelid = 'refresh_tokens'::regclass
    ) THEN
        ALTER TABLE refresh_tokens
        ADD CONSTRAINT uni_refresh_tokens_token_hash
        UNIQUE (token_hash);
    END IF;
END
$$;
