package authrepo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	accessTokenLength  = 32
	refreshTokenLength = 64

	AccessTokenMaxAge  = 60 * 60 * 24      // 24 hours
	RefreshTokenMaxAge = 60 * 60 * 24 * 90 // 90 Days
)

type AccessTokenData struct {
	UserId uuid.UUID `json:"user_id"`
}

type RefreshTokenData struct {
	UserId      uuid.UUID `json:"user_id"`
	AccessToken string    `json:"access_token"`
}

func (r *Repo) IssueAccessToken(ctx context.Context, userId uuid.UUID) (string, error) {
	accessTokenBytes := make([]byte, accessTokenLength)
	_, _ = rand.Read(accessTokenBytes)
	accessToken := base64.StdEncoding.EncodeToString(accessTokenBytes)

	data := AccessTokenData{
		UserId: userId,
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal access token data: %w", err)
	}

	ttl := time.Duration(AccessTokenMaxAge) * time.Second

	_, err = r.redisClient.Set(ctx, r.formatAccessToken(accessToken), dataStr, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("failed to store access token: %w", err)
	}

	return accessToken, nil
}

func (r *Repo) IssueRefreshToken(ctx context.Context, userId uuid.UUID, accessToken string) (string, error) {
	refreshTokenBytes := make([]byte, accessTokenLength)
	_, _ = rand.Read(refreshTokenBytes)
	refreshToken := base64.StdEncoding.EncodeToString(refreshTokenBytes)

	data := RefreshTokenData{
		UserId:      userId,
		AccessToken: accessToken,
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	ttl := time.Duration(RefreshTokenMaxAge) * time.Second

	_, err = r.redisClient.Set(ctx, r.formatRefreshToken(refreshToken), dataStr, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return refreshToken, nil
}

func (r *Repo) GetUserIdFromRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, bool, error) {
	dataStr, err := r.redisClient.Get(ctx, r.formatRefreshToken(refreshToken)).Result()
	if errors.Is(err, redis.Nil) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var refreshTokenData RefreshTokenData
	err = json.Unmarshal([]byte(dataStr), &refreshTokenData)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	return refreshTokenData.UserId, true, nil
}

func (r *Repo) GetUserIdFromAccessToken(ctx context.Context, accessToken string) (uuid.UUID, bool, error) {
	dataStr, err := r.redisClient.Get(ctx, r.formatAccessToken(accessToken)).Result()
	if errors.Is(err, redis.Nil) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("failed to get access token: %w", err)
	}

	var accessTokenData AccessTokenData
	err = json.Unmarshal([]byte(dataStr), &accessTokenData)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	return accessTokenData.UserId, true, nil
}

// RefreshTokenPair issues new access and refresh tokens, then deletes the old access and refresh token.
func (r *Repo) RefreshTokenPair(ctx context.Context, userId uuid.UUID) (string, string, error) {
	accessToken, err := r.IssueAccessToken(ctx, userId)
	if err != nil {
		return "", "", fmt.Errorf("failed to issue access token: %w", err)
	}

	refreshToken, err := r.IssueRefreshToken(ctx, userId, accessToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to issue refresh token: %w", err)
	}

	_, err = r.redisClient.Del(ctx, r.formatAccessToken(accessToken)).Result()
	if !errors.Is(err, redis.Nil) && err != nil {
		return "", "", fmt.Errorf("failed to delete access token: %w", err)
	}

	_, err = r.redisClient.Del(ctx, r.formatRefreshToken(refreshToken)).Result()
	if err != nil {
		return "", "", fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (r *Repo) formatAccessToken(token string) string {
	return fmt.Sprintf("auth:access:%s", token)
}

func (r *Repo) formatRefreshToken(token string) string {
	return fmt.Sprintf("auth:refresh:%s", token)
}
