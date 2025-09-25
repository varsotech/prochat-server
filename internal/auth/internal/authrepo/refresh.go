package authrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type RefreshTokenData struct {
	UserID string `json:"user_id"`
}

func (r *Repo) CreateRefreshToken(ctx context.Context, userId uuid.UUID) (uuid.UUID, error) {
	const refreshTokenTTL = 24 * time.Hour * 90 // 90 days

	refreshToken := uuid.New()

	data := RefreshTokenData{UserID: userId.String()}
	dataStr, err := json.Marshal(data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	_, err = r.redisClient.Set(ctx, r.formatRefreshTokenKey(refreshToken), dataStr, refreshTokenTTL).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken, nil
}

func (r *Repo) GetRefreshToken(ctx context.Context, refreshToken uuid.UUID) (*RefreshTokenData, bool, error) {
	dataStr, err := r.redisClient.Get(ctx, r.formatRefreshTokenKey(refreshToken)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var data RefreshTokenData
	err = json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	return &data, true, nil
}

func (r *Repo) formatRefreshTokenKey(refreshToken uuid.UUID) string {
	return fmt.Sprintf("auth:refresh_token:%s", refreshToken.String())
}
