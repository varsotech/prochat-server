package argon2

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"log/slog"
	"strings"
)

type Argon2Params struct {
	Memory     uint32 // in KB
	Time       uint32 // iterations
	Threads    uint8
	SaltLength uint32
	KeyLength  uint32
}

var defaultParams = &Argon2Params{
	Memory:     64 * 1024, // 64 MB
	Time:       3,
	Threads:    2,
	SaltLength: 16,
	KeyLength:  32,
}

// Hash generates an argon2id hash and returns an encoded string that contains all params.
// Format: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<salt_b64>$<hash_b64>
func Hash(password string, p *Argon2Params) (string, error) {
	if p == nil {
		p = defaultParams
	}

	salt := make([]byte, p.SaltLength)
	_, _ = rand.Read(salt)

	hash := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Threads, b64Salt, b64Hash)

	return encoded, nil
}

// Compare verifies a password against the encoded hash (from DB).
func Compare(password string, encodedHash string) (bool, error) {
	params, salt, expectedHash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLength)

	// constant time compare
	if len(computedHash) != len(expectedHash) {
		slog.Error("unexpected hash length computed")
		return false, nil
	}
	var diff byte
	for i := 0; i < len(computedHash); i++ {
		diff |= computedHash[i] ^ expectedHash[i]
	}
	return diff == 0, nil
}

func decodeHash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible hash version")
	}

	p := &Argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
