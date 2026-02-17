package identity

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type Claims struct {
	jwt.RegisteredClaims
}

const identityTokenExpiration = 15 * time.Minute

func NewClaims(host string, userId uuid.UUID) *Claims {
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    host,
			Subject:   userId.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(identityTokenExpiration)),
		},
	}
}

func Parse(tokenString string, publicKeyStr string) (*jwt.Token, error) {
	publicKey, err := getRSAPublicKey([]byte(publicKeyStr))
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate alg is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return publicKey, nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"RS256"}))

	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

// GetUnverifiedIssuer returns the JWT issuer without validating the authenticity of the JWT.
// This is useful for getting an issuer to determine which public key to verify the JWT with.
// Make sure identity is treated as user_id + issuer and never user_id alone.
func GetUnverifiedIssuer(tokenString string) (string, error) {
	parser := jwt.NewParser()

	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to parse claims")
	}

	return claims.GetIssuer()
}

func (i *Claims) Sign(privateKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, i)

	privKey, err := getRSAPrivateKey([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	signedToken, err := token.SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func getRSAPrivateKey(privateKeyBytes []byte) (*rsa.PrivateKey, error) {
	key, err := ssh.ParseRawPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}

func getRSAPublicKey(publicKeyBytes []byte) (*rsa.PublicKey, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse authorized key: %w", err)
	}

	// Convert to crypto.PublicKey
	cryptoPubKey, ok := pubKey.(ssh.CryptoPublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ssh public key")
	}

	// Extract *rsa.PublicKey
	rsaPubKey, ok := cryptoPubKey.CryptoPublicKey().(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an rsa public key")
	}

	return rsaPubKey, nil
}
