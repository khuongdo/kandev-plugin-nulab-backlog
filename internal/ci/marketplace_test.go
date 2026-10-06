package ci

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const thisRepo = "khuongdo/kandev-plugin-nulab-backlog"

var thisManifest = Manifest{ID: "nulab-backlog", Version: "0.1.0", MinKandevVersion: "0.96.0"}

func TestMarketplaceEntryHasTheRegistryFields(t *testing.T) {
	got, err := MarketplaceEntry(thisManifest, thisRepo, []string{"integrations"})
	require.NoError(t, err)
	require.Equal(t, "  - id: nulab-backlog\n    repo: khuongdo/kandev-plugin-nulab-backlog\n    categories: [integrations]\n", got)
	require.NotContains(t, got, "featured")
}

func TestMarketplaceEntryRejectsBadIDOrRepo(t *testing.T) {
	_, err := MarketplaceEntry(Manifest{ID: "Nulab_Backlog"}, thisRepo, nil)
	require.ErrorContains(t, err, "id")
	for _, repo := range []string{"no-slash", "a/b/c", "/b", "a/"} {
		_, err = MarketplaceEntry(thisManifest, repo, nil)
		require.ErrorContains(t, err, "repo", repo)
	}
}

func registry(entries string) []byte { return []byte("plugins:\n" + entries) }

func TestCheckRegistry(t *testing.T) {
	cases := []struct {
		name       string
		registry   []byte
		wantListed bool
		wantErr    string
	}{
		{"entry matches", registry("  - id: other\n    repo: a/b\n  - id: nulab-backlog\n    repo: " + thisRepo + "\n"), true, ""},
		{"entry has a different id", registry("  - id: backlog\n    repo: " + thisRepo + "\n"), true, "differs from the manifest id"},
		{"id used by another repo", registry("  - id: nulab-backlog\n    repo: someone/else\n"), false, "already used"},
		{"not listed yet", registry("  - id: other\n    repo: a/b\n"), false, ""},
		{"not YAML", []byte("plugins: [unclosed\n"), false, "parse"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			listed, err := CheckRegistry(tc.registry, "nulab-backlog", thisRepo)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantListed, listed)
		})
	}
}
