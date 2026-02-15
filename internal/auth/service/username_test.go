package service

import (
	"testing"
)

// TestGenerateUsername generates usernames a significant amount to ensure there's no panic
func TestGenerateUsername(t *testing.T) {
	for i := 0; i < 10000; i++ {
		_ = GenerateUsername()
	}
}
