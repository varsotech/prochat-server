package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"log/slog"
	"strconv"
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

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// hashPassword generates an argon2id hash and returns an encoded string that contains all params.
// Format: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<salt_b64>$<hash_b64>
func hashPassword(password string, p *Argon2Params) (string, error) {
	if p == nil {
		p = defaultParams
	}

	salt, err := generateRandomBytes(p.SaltLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %v", err)
	}

	hash := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Threads, b64Salt, b64Hash)

	return encoded, nil
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

// parseEncodedHash extracts params, salt and hash from the encoded string.
func parseEncodedHash(encoded string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// parts: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<hash>"]
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("invalid encoded argon2id format")
	}
	if parts[1] != "argon2id" {
		return nil, nil, nil, errors.New("unsupported algorithm")
	}

	// parse params in parts[3]
	var memory uint32
	var timeParam uint32
	var threads uint8

	paramPairs := strings.Split(parts[3], ",")
	for _, pair := range paramPairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, nil, nil, fmt.Errorf("invalid param: %s", pair)
		}
		k := kv[0]
		v := kv[1]
		switch k {
		case "m":
			mem, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return nil, nil, nil, err
			}
			memory = uint32(mem)
		case "t":
			t, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return nil, nil, nil, err
			}
			timeParam = uint32(t)
		case "p":
			p, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return nil, nil, nil, err
			}
			threads = uint8(p)
		default:
			// ignore unknown params
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}

	params := &Argon2Params{
		Memory:     memory,
		Time:       timeParam,
		Threads:    threads,
		SaltLength: uint32(len(salt)),
		KeyLength:  uint32(len(hash)),
	}

	return params, salt, hash, nil
}

// comparePassword verifies a password against the encoded hash (from DB).
func comparePassword(password string, encodedHash string) (bool, error) {
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
