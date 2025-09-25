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

type AccessTokenData struct {
	UserID string `json:"user_id"`
}

func (r *Repo) CreateAccessToken(ctx context.Context, userId uuid.UUID) (uuid.UUID, error) {
	const accessTokenTTL = 24 * time.Hour

	accessToken := uuid.New()

	data := AccessTokenData{UserID: userId.String()}
	dataStr, err := json.Marshal(data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal access token data: %w", err)
	}

	_, err = r.redisClient.Set(ctx, r.formatAccessTokenKey(accessToken), dataStr, accessTokenTTL).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to store access token: %w", err)
	}

	return accessToken, nil
}

func (r *Repo) GetAccessToken(ctx context.Context, accessToken uuid.UUID) (*AccessTokenData, bool, error) {
	dataStr, err := r.redisClient.Get(ctx, r.formatAccessTokenKey(accessToken)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get access token: %w", err)
	}

	var data AccessTokenData
	err = json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal access token data: %w", err)
	}

	return &data, true, nil
}

func (r *Repo) formatAccessTokenKey(accessToken uuid.UUID) string {
	return fmt.Sprintf("auth:access_token:%s", accessToken.String())
}
