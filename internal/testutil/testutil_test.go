package testutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyHasFakePrefixAndRandomHex(t *testing.T) {
	a, b := APIKey(t), APIKey(t)
	require.Regexp(t, regexp.MustCompile(`^test-api-key-[0-9a-f]{32}$`), a)
	require.NotEqual(t, a, b)
}

func TestWindowsCoverOnlyTheRandomPart(t *testing.T) {
	w := Windows(APIKeyPrefix+"0123456789", 8)
	require.Equal(t, []string{"01234567", "12345678", "23456789"}, w)
}

func TestAssertNoLeakPassesOnCleanText(t *testing.T) {
	AssertNoLeak(t, "nothing secret here", APIKey(t), 8, "Test User")
}

func TestU2_TokenHasFakePrefixAndRandomHex(t *testing.T) {
	a, b := Token(t), Token(t)
	require.Regexp(t, regexp.MustCompile(`^test-token-[0-9a-f]{32}$`), a)
	require.NotEqual(t, a, b)
}

func TestU2_WindowsCoverOnlyTheRandomPartOfAToken(t *testing.T) {
	w := Windows(TokenPrefix+"0123456789", 8)
	require.Equal(t, []string{"01234567", "12345678", "23456789"}, w)
}

func TestU2_AssertNoLeakCatchesATokenWindow(t *testing.T) {
	tok := Token(t)
	ft := &failRecorder{TB: t}
	AssertNoLeak(ft, "prefix "+tok[len(TokenPrefix):len(TokenPrefix)+8]+" suffix", tok, 8)
	require.True(t, ft.failed, "an 8-character window of a token must be caught")
}

// failRecorder records Fatalf instead of stopping the test.
type failRecorder struct {
	testing.TB
	failed bool
}

func (f *failRecorder) Helper()               {}
func (f *failRecorder) Fatalf(string, ...any) { f.failed = true }
