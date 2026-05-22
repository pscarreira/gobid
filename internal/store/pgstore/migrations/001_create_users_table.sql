CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    username VARCHAR(50) NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    bio TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()    
);
---- create above / drop below ----
DROP TABLE IF EXISTS users;