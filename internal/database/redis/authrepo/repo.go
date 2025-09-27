package authrepo

import (
	"github.com/redis/go-redis/v9"
)

type Repo struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *Repo {
	return &Repo{redisClient: redisClient}
}
