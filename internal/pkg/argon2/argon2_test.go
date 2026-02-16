package argon2

import (
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestDecodeHash(t *testing.T) {
	salt := []byte("1234567890abcdef")                 // 16 bytes
	hash := []byte("12345678901234567890123456789012") // 32 bytes
	
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	validEncoded := fmt.Sprintf("$argon2id$v=%d$m=65536,t=3,p=2$%s$%s", argon2.Version, b64Salt, b64Hash)

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantAlg     string
		wantSalt    []byte
		wantHash    []byte
		wantMem     uint32
		wantTime    uint32
		wantThreads uint8
	}{
		{
			name:        "valid string",
			input:       validEncoded,
			wantErr:     false,
			wantSalt:    salt,
			wantHash:    hash,
			wantMem:     65536,
			wantTime:    3,
			wantThreads: 2,
		},
		{
			name:    "wrong algorithm",
			input:   "$argon2i$v=999$m=65536,t=3,p=2$" + b64Salt + "$" + b64Hash,
			wantErr: true,
		},
		{
			name:    "missing hash",
			input:   "$argon2id$v=19$m=65536,t=3,p=2$" + b64Salt,
			wantErr: true,
		},
		{
			name:    "invalid base64 salt",
			input:   "$argon2id$v=19$m=65536,t=3,p=2$NOT_BASE64$" + b64Hash,
			wantErr: true,
		},
		{
			name:    "invalid base64 hash",
			input:   "$argon2id$v=19$m=65536,t=3,p=2$" + b64Salt + "$NOT_BASE64",
			wantErr: true,
		},
		{
			name:    "invalid numeric parameter",
			input:   "$argon2id$v=19$m=NaN,t=3,p=2$" + b64Salt + "$" + b64Hash,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params, gotSalt, gotHash, err := decodeHash(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check params
			if params.Memory != tc.wantMem {
				t.Errorf("Memory = %d, want %d", params.Memory, tc.wantMem)
			}
			if params.Time != tc.wantTime {
				t.Errorf("Time = %d, want %d", params.Time, tc.wantTime)
			}
			if params.Threads != tc.wantThreads {
				t.Errorf("Threads = %d, want %d", params.Threads, tc.wantThreads)
			}

			// Check salt + hash
			if string(gotSalt) != string(tc.wantSalt) {
				t.Errorf("Salt = %x, want %x", gotSalt, tc.wantSalt)
			}
			if string(gotHash) != string(tc.wantHash) {
				t.Errorf("Hash = %x, want %x", gotHash, tc.wantHash)
			}
		})
	}
}
