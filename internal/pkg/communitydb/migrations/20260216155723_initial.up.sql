CREATE TABLE members (
    id UUID PRIMARY KEY,
    user_address TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
