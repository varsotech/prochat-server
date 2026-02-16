package clientmetadata

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	redisClient *redis.Client
}

func NewCache(redisClient *redis.Client) *RedisCache {
	return &RedisCache{
		redisClient: redisClient,
	}
}

type ClientMetadata struct {
	Response      *Response `json:"response"`
	CachedLogoUrl string    `json:"cached_logo_url"`
}

func (r *RedisCache) SetClientMetadata(ctx context.Context, storedClientMetadata *ClientMetadata, ttl time.Duration) error {
	data, err := json.Marshal(storedClientMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal client metadata: %w", err)
	}

	_, err = r.redisClient.Set(ctx, r.clientMetadataKey(storedClientMetadata.Response.ClientID), data, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to set client metadata: %w", err)
	}

	return nil
}

// GetClientMetadata tries to retrieve client metadata from cache. Returns
func (r *RedisCache) GetClientMetadata(ctx context.Context, clientId string) (*ClientMetadata, bool, error) {
	result, err := r.redisClient.Get(ctx, r.clientMetadataKey(clientId)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to retrieve client metadata from cache: %w", err)
	}

	var clientMetadata ClientMetadata
	err = json.Unmarshal([]byte(result), &clientMetadata)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal client metadata: %w", err)
	}

	return &clientMetadata, true, nil
}

func (r *RedisCache) clientMetadataKey(clientId string) string {
	// Hash the client ID to avoid storing a big key for a very long URL
	clientIdHash := sha256.Sum256([]byte(clientId))
	return fmt.Sprintf("clientmetadata:%x", clientIdHash)
}
