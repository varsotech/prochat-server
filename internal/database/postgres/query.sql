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