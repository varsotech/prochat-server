package oauth

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

const codeLength = 32

type RedisCodeStore struct {
	redisClient *redis.Client
}

func NewCodeStore(redisClient *redis.Client) *RedisCodeStore {
	return &RedisCodeStore{
		redisClient: redisClient,
	}
}

type StoredCode struct {
	Code             string    `json:"code"`
	UserId           uuid.UUID `json:"user_id"`
	ClientId         string    `json:"client_id"`
	RedirectUriParam string    `json:"redirect_uri"`
}

// InsertCode generates a new authorization code and inserts it to redis
func (r *RedisCodeStore) InsertCode(ctx context.Context, userId uuid.UUID, clientId, redirectUri string) (string, error) {
	ttl := 5 * time.Minute

	codeBytes := make([]byte, codeLength)
	_, _ = rand.Read(codeBytes)
	code := base64.URLEncoding.EncodeToString(codeBytes)

	data, err := json.Marshal(StoredCode{
		Code:             code,
		UserId:           userId,
		ClientId:         clientId,
		RedirectUriParam: redirectUri,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal code json: %w", err)
	}

	_, err = r.redisClient.Set(ctx, r.codeKey(code), data, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("failed to set client metadata: %w", err)
	}

	return code, nil
}

var ErrCodeNotFound = errors.New("code not found in redis")

// DeleteCode deletes the authorization code and returns it. Returns ErrCodeNotFound if code was not found.
func (r *RedisCodeStore) DeleteCode(ctx context.Context, code string) (*StoredCode, error) {
	data, err := r.redisClient.Get(ctx, r.codeKey(code)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get code from redis: %w", err)
	}

	_, err = r.redisClient.Del(ctx, r.codeKey(code)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to delete code from redis: %w", err)
	}

	var storedCode StoredCode
	err = json.Unmarshal([]byte(data), &storedCode)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal code: %w", err)
	}

	return &storedCode, nil
}

func (r *RedisCodeStore) codeKey(code string) string {
	return fmt.Sprintf("oauthcode:%x", code)
}
