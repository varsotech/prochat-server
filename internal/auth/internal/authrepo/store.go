package authrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	AccessTokenMaxAge  = 60 * 60 * 24      // 24 hours
	RefreshTokenMaxAge = 60 * 60 * 24 * 90 // 90 Days
)

type AccessTokenData struct {
	UserId uuid.UUID `json:"user_id"`
}

type RefreshTokenData struct {
	UserId      uuid.UUID `json:"user_id"`
	AccessToken uuid.UUID `json:"access_token"`
}

func (r *Repo) IssueAccessToken(ctx context.Context, userId uuid.UUID) (uuid.UUID, error) {
	accessToken := uuid.New()
	data := AccessTokenData{
		UserId: userId,
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal access token data: %w", err)
	}

	ttl := time.Duration(AccessTokenMaxAge) * time.Second

	_, err = r.redisClient.Set(ctx, r.formatAccessToken(accessToken), dataStr, ttl).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to store access token: %w", err)
	}

	return accessToken, nil
}

func (r *Repo) IssueRefreshToken(ctx context.Context, userId, accessToken uuid.UUID) (uuid.UUID, error) {
	refreshToken := uuid.New()
	data := RefreshTokenData{
		UserId:      userId,
		AccessToken: accessToken,
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	ttl := time.Duration(RefreshTokenMaxAge) * time.Second

	_, err = r.redisClient.Set(ctx, r.formatRefreshToken(refreshToken), dataStr, ttl).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, nil
}

func (r *Repo) GetUserIdFromRefreshToken(ctx context.Context, tokenId uuid.UUID) (string, bool, error) {
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

func (r *Repo) formatAccessToken(token uuid.UUID) string {
	return fmt.Sprintf("auth:access:%s", token.String())
}

func (r *Repo) formatRefreshToken(token uuid.UUID) string {
	return fmt.Sprintf("auth:refresh:%s", token.String())
}
