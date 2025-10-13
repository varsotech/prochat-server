package imageproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SignURL(secret, url string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(url))
	return hex.EncodeToString(mac.Sum(nil))
}
