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
