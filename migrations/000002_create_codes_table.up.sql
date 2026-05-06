CREATE TABLE IF NOT EXISTS codes
(
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(10) NOT NULL,
    receiver   VARCHAR(255) NOT NULL,
    purpose VARCHAR(30) NOT NULL DEFAULT 'email_verification',
    expired_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '10 minutes',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_code_id ON codes (id);
CREATE INDEX IF NOT EXISTS idx_codes_receiver ON codes (receiver);
