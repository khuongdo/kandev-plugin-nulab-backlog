// Package testutil holds helpers shared by the plugin's tests. It never holds
// a real credential: every key is generated per test run (NFR4.1).
package testutil

import (
	"crypto/rand"
	"strings"
	"testing"
)

// APIKeyPrefix starts every generated test key.
const APIKeyPrefix = "test-api-key-" //nolint:gosec // G101: prefix of a generated fake key, not a credential

// TokenPrefix starts every generated fake token, client secret, auth code
// and state nonce.
const TokenPrefix = "test-token-" //nolint:gosec // G101: prefix of a generated fake token, not a credential

// APIKey returns "test-api-key-" plus 32 random characters (see randomPart).
func APIKey(t testing.TB) string {
	t.Helper()
	return APIKeyPrefix + randomPart(t)
}

// Token returns "test-token-" plus 32 random characters. Use it for every
// fake access token, refresh token, client secret and auth code.
func Token(t testing.TB) string {
	t.Helper()
	return TokenPrefix + randomPart(t)
}

// secretAlphabet is 16 uppercase letters with no hex digit and no vowel.
// Fixture and log text is lowercase words, digits, hex IDs and timestamps, so
// a 4-character window of a secret cannot match it by chance (T-SEC-02). H is
// kept and T dropped so "HTTP" cannot be formed.
const secretAlphabet = "GHJKLMNPQRSVWXYZ"

// randomPart returns 32 characters of secretAlphabet (128 random bits).
func randomPart(t testing.TB) string {
	t.Helper()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("generate test secret: %v", err)
	}
	for i := range b {
		b[i] = secretAlphabet[b[i]&0x0f]
	}
	return string(b)
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

// LeakWindow is the length of the partial-secret windows AssertNoLeak
// searches for.
const LeakWindow = 4

// AssertNoLeak fails the test when text contains the key, any LeakWindow
// window of its random part, or any of the extra forbidden strings.
func AssertNoLeak(t testing.TB, text, key string, forbidden ...string) {
	t.Helper()
	if key != "" && strings.Contains(text, key) {
		t.Fatalf("text contains the full API key")
	}
	for _, w := range Windows(key, LeakWindow) {
		if strings.Contains(text, w) {
			t.Fatalf("text contains part of the API key (window %d chars)", LeakWindow)
		}
	}
	for _, f := range forbidden {
		if f != "" && strings.Contains(text, f) {
			t.Fatalf("text contains a forbidden value")
		}
	}
}
