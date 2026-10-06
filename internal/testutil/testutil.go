// Package testutil holds helpers shared by the plugin's tests. It never holds
// a real credential: every key is generated per test run (NFR4.1).
package testutil

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
)

// APIKeyPrefix starts every generated test key.
const APIKeyPrefix = "test-api-key-" //nolint:gosec // G101: prefix of a generated fake key, not a credential

// TokenPrefix starts every generated fake token, client secret, auth code
// and state nonce.
const TokenPrefix = "test-token-" //nolint:gosec // G101: prefix of a generated fake token, not a credential

// APIKey returns "test-api-key-" plus 32 random hex characters.
func APIKey(t testing.TB) string {
	t.Helper()
	return APIKeyPrefix + randomHex(t)
}

// Token returns "test-token-" plus 32 random hex characters. Use it for every
// fake access token, refresh token, client secret and auth code.
func Token(t testing.TB) string {
	t.Helper()
	return TokenPrefix + randomHex(t)
}

func randomHex(t testing.TB) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("generate test secret: %v", err)
	}
	return hex.EncodeToString(b)
}

// Windows returns every n-character window of the random part of key (an
// APIKey or a Token). Leak
// tests search for these so a partial key is caught too, without matching
// the shared prefix.
func Windows(key string, n int) []string {
	random := strings.TrimPrefix(strings.TrimPrefix(key, APIKeyPrefix), TokenPrefix)
	var out []string
	for i := 0; i+n <= len(random); i++ {
		out = append(out, random[i:i+n])
	}
	return out
}

// AssertNoLeak fails the test when text contains the key, any n-character
// window of its random part, or any of the extra forbidden strings.
func AssertNoLeak(t testing.TB, text, key string, n int, forbidden ...string) {
	t.Helper()
	if key != "" && strings.Contains(text, key) {
		t.Fatalf("text contains the full API key")
	}
	for _, w := range Windows(key, n) {
		if strings.Contains(text, w) {
			t.Fatalf("text contains part of the API key (window %d chars)", n)
		}
	}
	for _, f := range forbidden {
		if f != "" && strings.Contains(text, f) {
			t.Fatalf("text contains a forbidden value")
		}
	}
}
