package str_test

import (
	"app/pkg/str"
	"encoding/base64"
	"testing"
)

func TestParseInt32(t *testing.T) {
	t.Parallel()
	// Test case 1: Valid integer string
	intStr := "12345"
	expected := int32(12345)
	result, err := str.ParseInt32(intStr)
	if err != nil {
		t.Errorf("unexpected error for input %s: %v", intStr, err)
	}
	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}

	// Test case 2: Empty string
	emptyStr := ""
	result, err = str.ParseInt32(emptyStr)
	if err != nil {
		t.Errorf("unexpected error for empty string: %v", err)
	}
	if result != 0 {
		t.Errorf("expected 0 for empty string, got %d", result)
	}

	// Test case 3: Invalid integer string
	invalidStr := "abc"
	_, err = str.ParseInt32(invalidStr)
	if err == nil {
		t.Error("expected an error for invalid integer string, but got none")
	}
}

func TestGenerateRandomBase64String(t *testing.T) {
	t.Parallel()
	// Test case: Generate a random state with a specific length
	length := 32
	state, err := str.GenerateRandomBase64String()
	if err != nil {
		t.Errorf("error generating random state: %v", err)
	}
	expectedLength := base64.StdEncoding.EncodedLen(length)
	if len(state) != expectedLength {
		t.Errorf("expected state length of %d, but got %d", expectedLength, len(state))
	}
}

func TestGenerateRandomHexString(t *testing.T) {
	t.Parallel()
	// Test case: Generate a random string with a specific length
	length := 32
	randomString, err := str.GenerateRandomHexString()
	if err != nil {
		t.Errorf("error generating random string: %v", err)
	}
	if len(randomString) != length*2 { // Hex encoding doubles the length
		t.Errorf("expected string length of %d, but got %d", length*2, len(randomString))
	}
}

