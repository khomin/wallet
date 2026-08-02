CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    image_url TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_name ON users(name);  