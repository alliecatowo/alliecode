package bridge

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const workSecretBytes = 32

// NewWorkSecret generates a random bridge work secret.
func NewWorkSecret() (string, error) {
	buf := make([]byte, workSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate work secret: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// WorkSecretHash returns a stable hash of a work secret.
func WorkSecretHash(secret string) (string, error) {
	normalized := strings.TrimSpace(secret)
	if normalized == "" {
		return "", errors.New("work secret cannot be empty")
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:]), nil
}

// VerifyWorkSecret checks if secret matches hash.
func VerifyWorkSecret(secret, expectedHash string) bool {
	hash, err := WorkSecretHash(secret)
	if err != nil {
		return false
	}
	return strings.EqualFold(hash, strings.TrimSpace(expectedHash))
}
