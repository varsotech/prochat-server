package authrepo

import (
	"encoding/json"
	"fmt"
	"time"
)

type TokenType string

const (
	AccessTokenType  TokenType = "access_token"
	RefreshTokenType TokenType = "refresh_token"

	AccessTokenMaxAge  = 60 * 60 * 24      // 24 hours
	RefreshTokenMaxAge = 60 * 60 * 24 * 90 // 90 Days
)

type TokenData struct {
	Type        TokenType     `json:"type"`
	Id          string        `json:"id"`
	TTL         time.Duration `json:"ttl"`
	AccessToken string        `json:"access_token,omitempty"`
	UserId      string        `json:"user_id"`
}

// TODO: rename
func (tokenData *TokenData) formatTokenKey() string {
	return fmt.Sprintf("auth:%s:%s", tokenData.Type, tokenData.Id)
}

func (tokenData *TokenData) formatTokenValue() string {
	data, _ := json.Marshal(map[string]string{"user_id": tokenData.UserId})
	return string(data)
}

func (tokenData *TokenData) formatTokenValueWithAccessToken(accessToken string) string {
	data, _ := json.Marshal(map[string]string{"access_token": accessToken, "user_id": tokenData.UserId})
	return string(data)
}
