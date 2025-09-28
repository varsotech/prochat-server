package http

import (
	"context"
	"github.com/google/uuid"
)

// define an unexported type for the key to avoid collisions, see: https://go.dev/blog/context
type contextKey string

const userIDKey contextKey = "userID"

func contextSetUserId(ctx context.Context, userId uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userId)
}

// contextGetUserId retrieves the userId from the context, if present.
func contextGetUserId(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return uuid.UUID{}, false
	}
	uid, ok := val.(uuid.UUID)
	return uid, ok
}
