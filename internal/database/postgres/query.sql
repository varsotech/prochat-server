-- name: CreateAnonymousUser :one
INSERT INTO users (id, username) 
VALUES ($1, $2) 
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (id, username, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;
