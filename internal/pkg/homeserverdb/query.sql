-- name: CreateAnonymousUser :one
INSERT INTO users (id, username, display_name)
VALUES ($1, $2, $3)
    RETURNING *;

-- name: CreateUser :one
INSERT INTO users (id, username, display_name, email, password_hash)
VALUES ($1, $2, $3, $4, $5)
    RETURNING *;

-- name: GetUserByLogin :one
SELECT * FROM users WHERE username = @login OR email = @login;

-- name: GetUserById :one
SELECT id, username, display_name, email FROM users WHERE id = $1;

-- name: UpsertUserServer :one
INSERT INTO user_servers (user_id, host)
VALUES ($1, $2)
    ON CONFLICT (user_id, host) DO NOTHING
    RETURNING *;

-- name: GetUserServers :many
SELECT host FROM user_servers WHERE user_id = @user_id;