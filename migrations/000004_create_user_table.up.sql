CREATE EXTENSION IF NOT EXISTS citext;
CREATE TABLE IF NOT EXISTS users(
    id bigserial PRIMARY KEY,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash BYTEA NOT NULL,
    activated bool NOT NULL,
    version INTEGER NOT NULL DEFAULT 1
);
