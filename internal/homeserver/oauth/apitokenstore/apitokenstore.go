package apitokenstore

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

	AccessTokenMaxAge  = int32(60 * 60 * 24)      // 24 hours
	RefreshTokenMaxAge = int32(60 * 60 * 24 * 90) // 90 Days
)

type AccessTokenData struct {
	UserId       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
}

type RefreshTokenData struct {
	UserId      uuid.UUID `json:"user_id"`
	AccessToken string    `json:"access_token"`
}

var RefreshTokenNotFoundError = errors.New("refresh token not found")

type IssueTokenPairResult struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresIn  int32
	RefreshTokenExpiresIn int32
}

type TokenStore struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *TokenStore {
	return &TokenStore{redisClient: redisClient}
}

func (r *TokenStore) IssueTokenPair(ctx context.Context, userId uuid.UUID) (IssueTokenPairResult, error) {
	accessTokenBytes := make([]byte, accessTokenLength)
	_, _ = rand.Read(accessTokenBytes)
	accessToken := base64.StdEncoding.EncodeToString(accessTokenBytes)

	refreshTokenBytes := make([]byte, refreshTokenLength)
	_, _ = rand.Read(refreshTokenBytes)
	refreshToken := base64.StdEncoding.EncodeToString(refreshTokenBytes)

	err := r.storeAccessToken(ctx, userId, accessToken, refreshToken)
	if err != nil {
		return IssueTokenPairResult{}, err
	}

	err = r.storeRefreshToken(ctx, userId, accessToken, refreshToken)
	if err != nil {
		return IssueTokenPairResult{}, err
	}

	return IssueTokenPairResult{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  AccessTokenMaxAge,
		RefreshTokenExpiresIn: RefreshTokenMaxAge,
	}, nil
}

func (r *TokenStore) storeAccessToken(ctx context.Context, userId uuid.UUID, accessToken, refreshToken string) error {
	data := AccessTokenData{
		UserId:       userId,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal access token data: %w", err)
	}

	ttl := time.Duration(AccessTokenMaxAge) * time.Second

	_, err = r.redisClient.Set(ctx, r.formatAccessToken(accessToken), dataStr, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	return nil
}

func (r *TokenStore) storeRefreshToken(ctx context.Context, userId uuid.UUID, accessToken, refreshToken string) error {
	data := RefreshTokenData{
		UserId:      userId,
		AccessToken: accessToken,
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	ttl := time.Duration(RefreshTokenMaxAge) * time.Second

	_, err = r.redisClient.Set(ctx, r.formatRefreshToken(refreshToken), dataStr, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

func (r *TokenStore) GetRefreshTokenData(ctx context.Context, refreshToken string) (*RefreshTokenData, bool, error) {
	dataStr, err := r.redisClient.Get(ctx, r.formatRefreshToken(refreshToken)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var refreshTokenData RefreshTokenData
	err = json.Unmarshal([]byte(dataStr), &refreshTokenData)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	return &refreshTokenData, true, nil
}

func (r *TokenStore) GetAccessTokenData(ctx context.Context, accessToken string) (AccessTokenData, bool, error) {
	dataStr, err := r.redisClient.Get(ctx, r.formatAccessToken(accessToken)).Result()
	if errors.Is(err, redis.Nil) {
		return AccessTokenData{}, false, nil
	}
	if err != nil {
		return AccessTokenData{}, false, fmt.Errorf("failed to get access token: %w", err)
	}

	var accessTokenData AccessTokenData
	err = json.Unmarshal([]byte(dataStr), &accessTokenData)
	if err != nil {
		return AccessTokenData{}, false, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	return accessTokenData, true, nil
}

type RefreshTokenPairResult struct {
	AccessToken  string
	RefreshToken string
}

// RefreshTokenPair issues new access and refresh tokens, then deletes the old access and refresh token.
// If the refresh token was not found, RefreshTokenNotFoundError is returned.
func (r *TokenStore) RefreshTokenPair(ctx context.Context, oldRefreshToken string) (RefreshTokenPairResult, error) {
	refreshTokenData, found, err := r.GetRefreshTokenData(ctx, oldRefreshToken)
	if err != nil {
		return RefreshTokenPairResult{}, fmt.Errorf("failed to get user id from refresh token: %w", err)
	}

	if !found {
		return RefreshTokenPairResult{}, RefreshTokenNotFoundError
	}

	issueTokenPairResult, err := r.IssueTokenPair(ctx, refreshTokenData.UserId)
	if err != nil {
		return RefreshTokenPairResult{}, fmt.Errorf("failed to issue token pair: %w", err)
	}

	err = r.DeleteTokenPair(ctx, refreshTokenData.AccessToken, oldRefreshToken)
	if err != nil {
		return RefreshTokenPairResult{}, fmt.Errorf("failed to delete token pair: %w", err)
	}

	return RefreshTokenPairResult{
		AccessToken:  issueTokenPairResult.AccessToken,
		RefreshToken: issueTokenPairResult.RefreshToken,
	}, nil
}

func (r *TokenStore) DeleteTokenPair(ctx context.Context, accessToken, refreshToken string) error {
	_, err := r.redisClient.Del(ctx, r.formatAccessToken(accessToken)).Result()
	if err != nil && !errors.Is(err, redis.Nil) { // Allow access token to not exist on deletion
		return fmt.Errorf("failed to delete access token: %w", err)
	}

	_, err = r.redisClient.Del(ctx, r.formatRefreshToken(refreshToken)).Result()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (r *TokenStore) formatAccessToken(token string) string {
	return fmt.Sprintf("oauth:access:%s", token)
}

func (r *TokenStore) formatRefreshToken(token string) string {
	return fmt.Sprintf("oauth:refresh:%s", token)
}
