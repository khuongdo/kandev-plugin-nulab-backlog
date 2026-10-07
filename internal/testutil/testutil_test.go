package testutil

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyHasFakePrefixAndRandomPart(t *testing.T) {
	a, b := APIKey(t), APIKey(t)
	require.Regexp(t, regexp.MustCompile(`^test-api-key-[GHJKLMNPQRSVWXYZ]{32}$`), a)
	require.NotEqual(t, a, b)
}

func TestWindowsCoverOnlyTheRandomPart(t *testing.T) {
	w := Windows(APIKeyPrefix+"0123456789", 8)
	require.Equal(t, []string{"01234567", "12345678", "23456789"}, w)
}

func TestAssertNoLeakPassesOnCleanText(t *testing.T) {
	AssertNoLeak(t, "nothing secret here", APIKey(t), "Test User")
}

func TestU2_TokenHasFakePrefixAndRandomPart(t *testing.T) {
	a, b := Token(t), Token(t)
	require.Regexp(t, regexp.MustCompile(`^test-token-[GHJKLMNPQRSVWXYZ]{32}$`), a)
	require.NotEqual(t, a, b)
}

func TestU2_WindowsCoverOnlyTheRandomPartOfAToken(t *testing.T) {
	w := Windows(TokenPrefix+"0123456789", 8)
	require.Equal(t, []string{"01234567", "12345678", "23456789"}, w)
}

func TestU2_AssertNoLeakCatchesATokenWindow(t *testing.T) {
	tok := Token(t)
	ft := &failRecorder{TB: t}
	AssertNoLeak(ft, "prefix "+tok[len(TokenPrefix):len(TokenPrefix)+8]+" suffix", tok)
	require.True(t, ft.failed, "an 8-character window of a token must be caught")
}

// NFR3.2: no 4-character substring of a secret may appear (T-SEC-02).
func TestAssertNoLeak_Catches4CharWindow(t *testing.T) {
	for _, secret := range []string{APIKey(t), Token(t)} {
		random := secret[strings.LastIndex(secret, "-")+1:]
		for i := 0; i+4 <= len(random); i++ {
			ft := &failRecorder{TB: t}
			AssertNoLeak(ft, "x "+random[i:i+4]+" y", secret)
			require.True(t, ft.failed, "window at %d must be caught", i)
		}
	}
}

func TestAssertNoLeak_PassesWithout4Char(t *testing.T) {
	key := APIKeyPrefix + "0123456789abcdef"
	ft := &failRecorder{TB: t}
	AssertNoLeak(ft, "012 123 9ab bcd cdef0", key)
	require.True(t, ft.failed, "cdef is a 4-character window")
	ft = &failRecorder{TB: t}
	AssertNoLeak(ft, "012 123 9ab bcd def "+APIKeyPrefix+TokenPrefix, key)
	require.False(t, ft.failed, "3-character pieces and the fixed prefixes are not leaks")
}

// The random part shares no character with digits, hex IDs, timestamps or
// lowercase log text, so a 4-character window cannot match fixture text by
// chance (T-SEC-02).
func TestSecretRandomPartAvoidsHexAndDigits(t *testing.T) {
	for range 100 {
		for _, secret := range []string{APIKey(t), Token(t)} {
			random := secret[strings.LastIndex(secret, "-")+1:]
			require.Len(t, random, 32)
			require.NotRegexp(t, `[0-9a-z]`, random)
		}
	}
}

// failRecorder records Fatalf instead of stopping the test.
type failRecorder struct {
	testing.TB
	failed bool
}

func (f *failRecorder) Helper()               {}
func (f *failRecorder) Fatalf(string, ...any) { f.failed = true }
