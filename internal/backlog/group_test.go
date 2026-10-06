package backlog

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueueGroupForEachCall(t *testing.T) {
	cases := []struct {
		method, path string
		want         Group
	}{
		{http.MethodGet, "/api/v2/users/myself", GroupRead},
		{http.MethodGet, "/api/v2/projects", GroupRead},
		{http.MethodGet, "/api/v2/issues/PROJ-1", GroupRead},
		{http.MethodGet, "/api/v2/issues", GroupSearch},
		{http.MethodGet, "/api/v2/issues/count", GroupSearch},
		{http.MethodPost, "/api/v2/issues", GroupUpdate},
		{http.MethodPatch, "/api/v2/issues/PROJ-1", GroupUpdate},
		{http.MethodDelete, "/api/v2/issues/PROJ-1", GroupUpdate},
		{http.MethodPost, "/api/v2/oauth2/token", GroupUpdate},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			require.Equal(t, tc.want, group(tc.method, tc.path))
		})
	}
}

func TestQueueGroupNamesAreStable(t *testing.T) {
	require.Equal(t, "read", GroupRead.String())
	require.Equal(t, "update", GroupUpdate.String())
	require.Equal(t, "search", GroupSearch.String())
	require.Equal(t, "unknown", Group(99).String())
}
