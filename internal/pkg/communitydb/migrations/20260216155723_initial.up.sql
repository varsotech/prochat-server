CREATE TABLE members (
    id UUID PRIMARY KEY,
    user_address TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);