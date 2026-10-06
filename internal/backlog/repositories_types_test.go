package backlog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU4_RepoParse_DecodesTheList(t *testing.T) {
	got, err := parseRepositories(readFixture(t, "repositories_ok.json"), spaceHost)
	require.NoError(t, err)
	require.Equal(t, []Repository{
		{ID: 11, ProjectID: 101, Name: "web-app", HTTPURL: "https://example-space.backlog.com/git/PROJ/web-app.git"},
		{ID: 12, ProjectID: 101, Name: "api", HTTPURL: "https://example-space.backlog.com/git/PROJ/api.git"},
	}, got)
}

func TestU4_RepoParse_AcceptsAnEmptyList(t *testing.T) {
	got, err := parseRepositories(readFixture(t, "repositories_empty.json"), spaceHost)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestU4_RepoParse_RejectsUnusableBodies(t *testing.T) {
	for name, body := range map[string]string{
		"a non-numeric id": `[{"id":"11","projectId":101,"name":"web-app","httpUrl":"https://example-space.backlog.com/git/PROJ/web-app.git"}]`,
		"an empty name":    `[{"id":11,"projectId":101,"name":"","httpUrl":"https://example-space.backlog.com/git/PROJ/web-app.git"}]`,
		"a plain http URL": `[{"id":11,"projectId":101,"name":"web-app","httpUrl":"http://example-space.backlog.com/git/PROJ/web-app.git"}]`,
		"another host":     `[{"id":11,"projectId":101,"name":"web-app","httpUrl":"https://evil.example.com/git/PROJ/web-app.git"}]`,
		"user info in URL": `[{"id":11,"projectId":101,"name":"web-app","httpUrl":"https://someone@example-space.backlog.com/git/PROJ/web-app.git"}]`,
		"not a list":       `{"id":11}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseRepositories([]byte(body), spaceHost)
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, KindUnreachable, be.Kind)
			require.Equal(t, "body", be.Class)
		})
	}
}
