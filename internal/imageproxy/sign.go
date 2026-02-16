package imageproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"path"
)

type Signer struct {
	baseUrl    string
	secretKey  string
	secretSalt string
}

func NewSigner(imageProxyConfig *Config) *Signer {
	return &Signer{
		baseUrl:    imageProxyConfig.ImageProxyBaseUrl,
		secretKey:  imageProxyConfig.ImageProxySecretKey,
		secretSalt: imageProxyConfig.ImageProxySecretSalt,
	}
}

func (s *Signer) GenerateSignature(inputUrl string) string {
	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(inputUrl + s.secretSalt))
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateSignedURL constructs a signed cache URL path from a url
// TODO: What happens if the secret needs to be rotated?
func (s *Signer) GenerateSignedURL(inputUrl string) string {
	return path.Join(s.baseUrl, "external", s.GenerateSignature(inputUrl), url.PathEscape(inputUrl))
}
