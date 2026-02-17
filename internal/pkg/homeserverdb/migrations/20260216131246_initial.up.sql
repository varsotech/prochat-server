CREATE TABLE user_servers (
    id bigserial PRIMARY KEY,
    user_id UUID NOT NULL,
    host TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX user_servers_user_id_host_idx
    ON user_servers (user_id, host);

