package connection

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestValidateAPIKey(t *testing.T) {
	key := testutil.APIKey(t)
	cases := []struct {
		name  string
		input string
		want  string // empty means rejected
	}{
		{"accepts a valid key", key, key},
		{"trims surrounding whitespace", "  " + key + "\n", key},
		{"accepts exactly 256 characters", strings.Repeat("k", 256), strings.Repeat("k", 256)},
		{"rejects an empty key", "", ""},
		{"rejects whitespace only", " \t ", ""},
		{"rejects 257 characters", strings.Repeat("k", 257), ""},
		{"rejects an inner space", "abc def", ""},
		{"rejects a control character", "abc\x01def", ""},
		{"rejects a non-ASCII character", "abcédef", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateAPIKey(tc.input)
			if tc.want == "" {
				require.True(t, errors.Is(err, ErrInvalidAPIKey))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestInvalidAPIKeyErrorNeverContainsTheInput(t *testing.T) {
	input := testutil.APIKey(t) + " x"
	_, err := ValidateAPIKey(input)
	require.Error(t, err)
	testutil.AssertNoLeak(t, err.Error(), input[:45])
}
