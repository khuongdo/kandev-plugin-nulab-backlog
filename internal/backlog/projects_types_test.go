package backlog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectsParseDecodesTheProjectList(t *testing.T) {
	got, err := parseProjects(readFixture(t, "projects_ok.json"))
	require.NoError(t, err)
	require.Equal(t, []Project{
		{ID: 101, Key: "PROJ", Name: "Test Project"},
		{ID: 102, Key: "DEMO", Name: "Demo Project", Archived: true},
	}, got, "archived projects are kept and flagged")
}

func TestProjectsParseAcceptsAnEmptyList(t *testing.T) {
	got, err := parseProjects(readFixture(t, "projects_empty.json"))
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestProjectsParseRejectsUnusableBodies(t *testing.T) {
	for name, body := range map[string]string{
		"a non-numeric id": `[{"id":"101","projectKey":"PROJ","name":"P"}]`,
		"a missing id":     `[{"projectKey":"PROJ","name":"P"}]`,
		"an empty key":     `[{"id":101,"projectKey":"","name":"P"}]`,
		"not a list":       `{"id":101}`,
		"invalid JSON":     `<html>`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseProjects([]byte(body))
			var be *Error
			require.True(t, errors.As(err, &be))
			require.Equal(t, KindUnreachable, be.Kind)
			require.Equal(t, "body", be.Class)
		})
	}
}
