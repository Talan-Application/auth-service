CREATE TABLE IF NOT EXISTS codes
(
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(10) NOT NULL,
    receiver   VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_code_id ON users (id);
