CREATE SCHEMA IF NOT EXISTS users;

SET search_path TO users;

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    name          VARCHAR(255)  NOT NULL,
    password_hash VARCHAR(255)  NOT NULL,
    status        VARCHAR(20)   NOT NULL DEFAULT 'active',
    is_verified   BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS emails (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addr        VARCHAR(255) NOT NULL,
    is_primary  BOOLEAN      NOT NULL DEFAULT FALSE,
    is_verified BOOLEAN      NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,

    CONSTRAINT uq_emails_addr UNIQUE (addr)
);

CREATE INDEX IF NOT EXISTS idx_emails_user_id ON emails(user_id);
CREATE INDEX IF NOT EXISTS idx_emails_addr    ON emails(addr);
CREATE INDEX IF NOT EXISTS idx_users_status   ON users(status);
