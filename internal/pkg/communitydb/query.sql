-- name: GetMemberByUserAddress :one
SELECT * FROM members WHERE user_address = @user_address;

-- name: UpsertMember :one
INSERT INTO members (id, user_address)
VALUES ($1, $2)
    ON CONFLICT (user_address) DO NOTHING
    RETURNING *;