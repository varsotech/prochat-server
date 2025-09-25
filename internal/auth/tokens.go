package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	AccessTokenType  TokenType = "access_token"
	RefreshTokenType TokenType = "refresh_token"

	accessTokenMaxAge  = 60 * 60 * 24      // 24 hours
	refreshTokenMaxAge = 60 * 60 * 24 * 90 // 90 Days
)

type TokenData struct {
	Type       TokenType     `json:"type"`
	Identifier string        `json:"identifier"`
	TTL        time.Duration `json:"ttl"`
	UserId     string        `json:"user_id"`
}

func createTokenPair(userId uuid.UUID) (TokenData, TokenData) {
	accessTokenTTL := time.Duration(accessTokenMaxAge) * time.Second
	refreshTokenTTL := time.Duration(refreshTokenMaxAge) * time.Second

	accessTokenId := uuid.New()
	refreshTokenId := uuid.New()

	accessTokenData := TokenData{
		Type:       AccessTokenType,
		Identifier: accessTokenId.String(),
		TTL:        accessTokenTTL,
		UserId:     userId.String(),
	}

	refreshTokenData := TokenData{
		Type:       RefreshTokenType,
		Identifier: refreshTokenId.String(),
		TTL:        refreshTokenTTL,
		UserId:     userId.String(),
	}

	return accessTokenData, refreshTokenData
}

func (tokenData *TokenData) authString() string {
	return fmt.Sprintf("auth:%s:%s", tokenData.Type, tokenData.Identifier)
}
