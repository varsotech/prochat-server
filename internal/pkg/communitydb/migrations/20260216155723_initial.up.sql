CREATE TABLE members (
    id UUID PRIMARY KEY,
    user_address TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE communities (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    is_default BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX communities_is_default_idx
    ON communities (is_default)
    WHERE is_default = TRUE;

CREATE TABLE community_members (
    id UUID PRIMARY KEY,
    member_id UUID NOT NULL,
    community_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX community_members_member_id_community_id_idx
    ON community_members (member_id, community_id);