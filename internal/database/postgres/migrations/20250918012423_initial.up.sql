CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    create_time TIMESTAMPTZ DEFAULT now()
);
