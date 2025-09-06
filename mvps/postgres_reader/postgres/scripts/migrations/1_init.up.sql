-- PostgreSQL complete schema migration

-- Model metadata table
CREATE TABLE IF NOT EXISTS meters (
    id TEXT PRIMARY KEY,
    ipv6 TEXT NOT NULL,
    port INTEGER NOT NULL,
    system_title TEXT NOT NULL,
    auth_password TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    block_cipher_key TEXT NOT NULL
);
