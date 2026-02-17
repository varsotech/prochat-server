-- name: GetMemberByUserAddress :one
SELECT * FROM members WHERE user_address = @user_address;

-- name: UpsertMember :one
INSERT INTO members (id, user_address)
VALUES ($1, $2)
    ON CONFLICT (user_address) DO NOTHING
    RETURNING *;

-- name: InsertCommunity :one
INSERT INTO communities (id, name)
VALUES ($1, $2)
    RETURNING *;

-- name: UpsertDefaultCommunity :one
INSERT INTO communities (id, name, is_default)
VALUES ($1, $2, true)
    ON CONFLICT (is_default)
    WHERE is_default = TRUE
    DO NOTHING
    RETURNING *;

-- name: GetDefaultCommunity :one
SELECT * FROM communities WHERE is_default = true;

-- name: GetMemberCommunities :many
SELECT c.* FROM community_members INNER JOIN communities c ON c.id = community_members.community_id WHERE community_members.member_id = $1;

-- name: UpsertCommunityMember :one
INSERT INTO community_members (id, member_id, community_id)
VALUES ($1, $2, $3)
    ON CONFLICT (member_id, community_id) DO NOTHING
    RETURNING *;