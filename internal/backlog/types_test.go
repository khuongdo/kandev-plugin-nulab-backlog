package backlog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name)) //nolint:gosec // fixture names are test constants
	require.NoError(t, err)
	return b
}

func TestParseUser(t *testing.T) {
	cases := []struct {
		name    string
		fixture string
		body    []byte
		want    User
		wantErr bool
	}{
		{name: "maps id, userId and name", fixture: "myself_ok.json", want: User{ID: 1234, UserID: "test.user", Name: "Test User"}},
		{name: "rejects a missing id", fixture: "myself_missing_id.json", wantErr: true},
		{name: "rejects an empty name", fixture: "myself_empty_name.json", wantErr: true},
		{name: "rejects a non-numeric id", fixture: "myself_string_id.json", wantErr: true},
		{name: "rejects an empty body", body: []byte{}, wantErr: true},
		{name: "rejects invalid JSON", body: []byte("<html>"), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body
			if tc.fixture != "" {
				body = readFixture(t, tc.fixture)
			}
			got, err := parseUser(body)
			if tc.wantErr {
				var be *Error
				require.True(t, errors.As(err, &be))
				require.Equal(t, KindUnreachable, be.Kind)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestKindForStatus(t *testing.T) {
	cases := []struct {
		status int
		kind   Kind
		isErr  bool
	}{
		{200, 0, false},
		{401, KindUnauthorized, true},
		{403, KindForbidden, true},
		{404, KindNotFound, true},
		{409, KindConflict, true},
		{400, KindInvalid, true},
		{422, KindInvalid, true},
		{429, KindRateLimited, true},
		{418, KindUnreachable, true},
		{302, KindUnreachable, true},
		{503, KindUnreachable, true},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("status %d", tc.status), func(t *testing.T) {
			kind, isErr := kindForStatus(tc.status)
			require.Equal(t, tc.isErr, isErr)
			if tc.isErr {
				require.Equal(t, tc.kind, kind)
			}
		})
	}
}

func TestErrorTextShowsOnlyKindAndStatus(t *testing.T) {
	err := &Error{Kind: KindUnauthorized, Status: 401, Class: "http"}
	require.Equal(t, "backlog: unauthorized (status 401)", err.Error())
	err = &Error{Kind: KindUnreachable, Class: "timeout"}
	require.Equal(t, "backlog: unreachable (status 0)", err.Error())
}

func TestKindStringsAreStable(t *testing.T) {
	require.Equal(t, "forbidden", KindForbidden.String())
	require.Equal(t, "rate_limited", KindRateLimited.String())
	require.Equal(t, "unknown", Kind(99).String())
}

func TestCredentialsHideSecretsWhenFormatted(t *testing.T) {
	key, token := testutil.APIKey(t), testutil.APIKey(t)
	c := Credentials{SpaceHost: "example-space.backlog.com", APIKey: key, AccessToken: token}
	testutil.AssertNoLeak(t, c.String()+fmt.Sprintf("%v %+v %#v", c, c, c), key, 8, token)
	require.Contains(t, c.String(), "example-space.backlog.com")
}

func TestCallClassValues(t *testing.T) {
	require.NotEqual(t, Interactive, Background)
}
