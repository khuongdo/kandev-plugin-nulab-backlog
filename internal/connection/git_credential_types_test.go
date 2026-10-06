package connection

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestU4_GitCredential_ValidatesTheInput(t *testing.T) {
	pw := testutil.Token(t)
	got, err := ValidateGitCredential(GitCredentialInput{Username: "  lan ", Password: pw})
	require.NoError(t, err)
	require.Equal(t, GitCredentialInput{Username: "lan", Password: pw}, got)

	cases := map[string]struct {
		in    GitCredentialInput
		field string
	}{
		"an empty user name":    {GitCredentialInput{Username: " ", Password: pw}, FieldGitUsername},
		"a long user name":      {GitCredentialInput{Username: strings.Repeat("u", 101), Password: pw}, FieldGitUsername},
		"a control character":   {GitCredentialInput{Username: "la\nn", Password: pw}, FieldGitUsername},
		"an empty password":     {GitCredentialInput{Username: "lan", Password: ""}, FieldGitPassword},
		"a long password":       {GitCredentialInput{Username: "lan", Password: strings.Repeat("p", 257)}, FieldGitPassword},
		"a password of 256 ok":  {GitCredentialInput{Username: "lan", Password: strings.Repeat("p", 256)}, ""},
		"a user name of 100 ok": {GitCredentialInput{Username: strings.Repeat("u", 100), Password: pw}, ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateGitCredential(tc.in)
			if tc.field == "" {
				require.NoError(t, err)
				return
			}
			var fe *FieldError
			require.True(t, errors.As(err, &fe))
			require.Equal(t, tc.field, fe.Field)
			testutil.AssertNoLeak(t, err.Error(), pw, 8)
		})
	}
}

func TestU4_GitSecret_RoundTripsAsJSON(t *testing.T) {
	in := gitSecret{Username: "lan", Password: testutil.Token(t), SpaceHost: "a.backlog.com", Revision: 3}
	b, err := json.Marshal(in) //nolint:gosec // G117: a test round-trip of a generated fake secret
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	require.Equal(t, map[string]any{"username": "lan", "password": in.Password, "spaceHost": "a.backlog.com", "revision": float64(3)}, m)
	var out gitSecret
	require.NoError(t, json.Unmarshal(b, &out))
	require.Equal(t, in, out)
}

func TestU4_GitCredential_HidesThePassword(t *testing.T) {
	pw := testutil.Token(t)
	c := GitCredential{Username: "lan", Password: pw, SpaceHost: "a.backlog.com"}
	for _, s := range []string{fmt.Sprintf("%v", c), fmt.Sprintf("%+v", c), fmt.Sprintf("%#v", c), fmt.Sprint(c)} {
		testutil.AssertNoLeak(t, s, pw, 8)
		require.Contains(t, s, "lan")
	}
}
