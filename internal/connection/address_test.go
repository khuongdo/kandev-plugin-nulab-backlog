package connection

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSpaceAddress(t *testing.T) {
	label63 := strings.Repeat("a", 63)
	label64 := strings.Repeat("a", 64)
	cases := []struct {
		name  string
		input string
		host  string // empty means rejected
	}{
		{"adds the https scheme", "myteam.backlog.com", "myteam.backlog.com"},
		{"lower-cases the host", "MyTeam.Backlog.com", "myteam.backlog.com"},
		{"accepts backlog.jp", "myteam.backlog.jp", "myteam.backlog.jp"},
		{"accepts backlogtool.com", "x.backlogtool.com", "x.backlogtool.com"},
		{"trims and allows a root path", "  https://myteam.backlog.com/  ", "myteam.backlog.com"},
		{"accepts an upper-case https scheme", "HTTPS://myteam.backlog.com", "myteam.backlog.com"},
		{"accepts a 63-character label", label63 + ".backlog.com", label63 + ".backlog.com"},
		{"accepts digits and inner hyphens", "my-team-2.backlog.com", "my-team-2.backlog.com"},
		{"rejects http", "http://myteam.backlog.com", ""},
		{"rejects an unknown suffix", "evil.example.com", ""},
		{"rejects an IP address", "192.168.1.10", ""},
		{"rejects an allowed suffix inside another host", "myteam.backlog.com.evil.io", ""},
		{"rejects a path", "https://myteam.backlog.com/path", ""},
		{"rejects a port", "myteam.backlog.com:8443", ""},
		{"rejects userinfo", "https://a@evil.io", ""},
		{"rejects userinfo on an allowed host", "https://a@myteam.backlog.com", ""},
		{"rejects a missing space label", "backlog.com", ""},
		{"rejects a trailing dot", "myteam.backlog.com.", ""},
		{"rejects more than one label", "a.b.backlog.com", ""},
		{"rejects non-ASCII", "tëam.backlog.com", ""},
		{"rejects a 64-character label", label64 + ".backlog.com", ""},
		{"rejects a leading hyphen", "-team.backlog.com", ""},
		{"rejects a trailing hyphen", "team-.backlog.com", ""},
		{"rejects a query", "https://myteam.backlog.com/?a=1", ""},
		{"rejects a fragment", "https://myteam.backlog.com/#x", ""},
		{"rejects an empty input", "   ", ""},
		{"rejects another scheme written with ://", "ftp://myteam.backlog.com", ""},
		{"rejects an underscore", "my_team.backlog.com", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := ParseSpaceAddress(tc.input)
			if tc.host == "" {
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrInvalidSpaceAddress))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.host, addr.Host)
			require.Equal(t, "https://"+tc.host, addr.BaseURL())
		})
	}
}

func TestInvalidSpaceAddressMessageStatesTheFormat(t *testing.T) {
	_, err := ParseSpaceAddress("evil.example.com")
	require.ErrorContains(t, err, "<space>.backlog.com")
}
