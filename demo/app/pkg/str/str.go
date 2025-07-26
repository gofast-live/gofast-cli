package str

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
)

// ParseInt32 converts a string to an int32, handling empty strings.
func ParseInt32(s string) (int32, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse '%s' as int32: %w", s, err)
	}
	return int32(n), nil
}

// GenerateRandomBase64String generates a random base64 URL-safe string
func GenerateRandomBase64String() (string, error) {
	// Generate random bytes
	randomBytes, err := generateRandomBytes()
	if err != nil {
		return "", err
	}
	// Encode the random bytes to a base64 URL-safe string
	state := base64.URLEncoding.EncodeToString(randomBytes)
	return state, nil
}

// GenerateRandomHexString generates a random hex string
func GenerateRandomHexString() (string, error) {
	// Generate random bytes
	randomBytes, err := generateRandomBytes()
	if err != nil {
		return "", err
	}
	// Encode the random bytes to a hex string
	hex := hex.EncodeToString(randomBytes)
	return hex, nil
}

func generateRandomBytes() ([]byte, error) {
	numberOfBytes := 32
	randomBytes := make([]byte, numberOfBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, fmt.Errorf("error generating random bytes: %w", err)
	}
	return randomBytes, nil
}
