package backlog

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestProjectsCallsTheProjectListWithEachCredentialKind(t *testing.T) {
	for name, mk := range map[string]func(string) Credentials{
		"api key":      func(s string) Credentials { return Credentials{SpaceHost: spaceHost, APIKey: s} },
		"access token": func(s string) Credentials { return Credentials{SpaceHost: spaceHost, AccessToken: s} },
	} {
		t.Run(name, func(t *testing.T) {
			var got *http.Request
			c, _ := fakeBacklog(t, func(w http.ResponseWriter, r *http.Request) {
				got = r.Clone(context.Background())
				_, _ = w.Write(readFixture(t, "projects_ok.json"))
			})
			projects, err := c.Projects(context.Background(), mk(testutil.Token(t)))
			require.NoError(t, err)
			require.Equal(t, http.MethodGet, got.Method)
			require.Equal(t, "/api/v2/projects", got.URL.Path)
			require.Len(t, projects, 2)
			require.Equal(t, "PROJ", projects[0].Key)
		})
	}
}

func TestProjectsEmptyList(t *testing.T) {
	c, _ := fakeBacklog(t, serveFixture(t, 200, "projects_empty.json"))
	projects, err := c.Projects(context.Background(), creds(testutil.APIKey(t)))
	require.NoError(t, err)
	require.Empty(t, projects)
}

func TestProjectsErrorsNeverEchoTheBody(t *testing.T) {
	for status, kind := range map[int]Kind{401: KindUnauthorized, 403: KindForbidden, 500: KindUnreachable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			bait := testutil.Token(t)
			body := bytes.ReplaceAll(readFixture(t, "error_500_bait.json"), []byte("{{BAIT}}"), []byte(bait))
			c, _ := fakeBacklog(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write(body)
			})
			ctx, logs := oauthLogs()
			_, err := c.Projects(ctx, creds(testutil.APIKey(t)))
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, kind, be.Kind)
			testutil.AssertNoLeak(t, err.Error()+logs.String(), bait)
		})
	}
}
