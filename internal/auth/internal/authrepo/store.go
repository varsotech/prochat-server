package authrepo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (r *Repo) SetAccessToken(ctx context.Context, token TokenData) error {
	slog.Info("store access token")

	_, err := r.redisClient.Set(ctx, token.formatTokenKey(), token.formatTokenValue(), token.TTL).Result()
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

func (r *Repo) SetRefreshToken(ctx context.Context, token TokenData, accessTokenId string) error {
	slog.Info("store refresh token")

	_, err := r.redisClient.Set(ctx, token.formatTokenKey(), token.formatTokenValueWithAccessToken(accessTokenId), token.TTL).Result()
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

func (r *Repo) GetUserIdFromToken(ctx context.Context, tokenId uuid.UUID, tokenType TokenType) (string, bool, error) {
	// TODO: return correct values

	dataStr, err := r.redisClient.Get(ctx, r.formatTokenKey(string(tokenType), tokenId)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to get access token: %w", err)
	}

	return dataStr, true, nil
}

func (r *Repo) DeleteToken(ctx context.Context, tokenType string, tokenId uuid.UUID) error {
	slog.Info("delete token")

	_, err := r.redisClient.Del(ctx, r.formatTokenKey(tokenType, tokenId)).Result()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (r *Repo) formatTokenKey(tokenType string, tokenId uuid.UUID) string {
	return fmt.Sprintf("auth:%s:%s", tokenType, tokenId.String())
}
