package authrepo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (r *Repo) StoreToken(ctx context.Context, key, value string, ttl time.Duration) error {
	slog.Info("store token")

	// TODO: the value is currently just the userId, and not {"user_id": <uuid>}
	_, err := r.redisClient.Set(ctx, key, value, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

func (r *Repo) GetToken(ctx context.Context, tokenId uuid.UUID, tokenType string) (string, bool, error) {
	dataStr, err := r.redisClient.Get(ctx, r.formatTokenKey(tokenType, tokenId)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to get access token: %w", err)
	}

	// TODO: return correct value
	return dataStr, true, nil
}

func (r *Repo) formatTokenKey(tokenType string, tokenId uuid.UUID) string {
	return fmt.Sprintf("auth:%s:%s", tokenType, tokenId.String())
}
