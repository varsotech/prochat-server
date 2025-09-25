package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/varsotech/prochat-server/internal/auth/internal/authrepo"
)

func createTokenPair(userId uuid.UUID) (authrepo.TokenData, authrepo.TokenData) {
	accessTokenTTL := time.Duration(authrepo.AccessTokenMaxAge) * time.Second
	refreshTokenTTL := time.Duration(authrepo.RefreshTokenMaxAge) * time.Second

	accessTokenId := uuid.New()
	refreshTokenId := uuid.New()

	accessTokenData := authrepo.TokenData{
		Type:   authrepo.AccessTokenType,
		Id:     accessTokenId.String(),
		TTL:    accessTokenTTL,
		UserId: userId.String(),
	}

	refreshTokenData := authrepo.TokenData{
		Type:   authrepo.RefreshTokenType,
		Id:     refreshTokenId.String(),
		TTL:    refreshTokenTTL,
		UserId: userId.String(),
	}

	return accessTokenData, refreshTokenData
}
