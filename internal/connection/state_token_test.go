package connection

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOAuthStateRoundTrips(t *testing.T) {
	nonce, err := newNonce()
	require.NoError(t, err)
	require.Len(t, nonce, 32)
	s := encodeState("ws-1", nonce)
	ws, got, err := decodeState(s)
	require.NoError(t, err)
	require.Equal(t, "ws-1", ws)
	require.Equal(t, nonce, got)
	raw, err := base64.RawURLEncoding.DecodeString(s)
	require.NoError(t, err, "the state is unpadded base64url")
	require.True(t, bytes.HasPrefix(raw, nonce))
}

func TestOAuthStateTwoTokensForOneWorkspaceDiffer(t *testing.T) {
	a, err := newNonce()
	require.NoError(t, err)
	b, err := newNonce()
	require.NoError(t, err)
	require.NotEqual(t, encodeState("ws-1", a), encodeState("ws-1", b))
}

func TestOAuthStateRejectsBadInput(t *testing.T) {
	nonce := bytes.Repeat([]byte{7}, 32)
	valid := encodeState("ws-1", nonce)
	cases := map[string]string{
		"empty":               "",
		"not base64":          "%%%%",
		"wrong alphabet":      strings.ReplaceAll(base64.StdEncoding.EncodeToString(append(bytes.Repeat([]byte{0xfb}, 32), "ws-1"...)), "=", ""),
		"padded":              base64.URLEncoding.EncodeToString(append(nonce, "ws-12"...)),
		"truncated":           valid[:40],
		"no workspace":        base64.RawURLEncoding.EncodeToString(nonce),
		"overlong":            base64.RawURLEncoding.EncodeToString(append(nonce, strings.Repeat("w", 200)...)),
		"workspace with path": base64.RawURLEncoding.EncodeToString(append(nonce, "../x"...)),
		"dot workspace":       base64.RawURLEncoding.EncodeToString(append(nonce, ".."...)),
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := decodeState(s)
			require.ErrorIs(t, err, errBadState)
		})
	}
}
