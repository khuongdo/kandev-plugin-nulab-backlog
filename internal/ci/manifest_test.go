package ci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeManifest(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "manifest.yaml")
	require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	return p
}

func TestReadManifestReturnsIdentityAndMinimumVersion(t *testing.T) {
	p := writeManifest(t, "id: \"demo\"\nversion: \"1.2.3\"\nmin_kandev_version: \"0.96.0\"\n")
	m, err := ReadManifest(p)
	require.NoError(t, err)
	require.Equal(t, Manifest{ID: "demo", Version: "1.2.3", MinKandevVersion: "0.96.0"}, m)
}

func TestReadManifestRefusesMissingFields(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"missing id", "version: 1.2.3\nmin_kandev_version: 0.96.0\n", "id"},
		{"missing version", "id: demo\nmin_kandev_version: 0.96.0\n", "version"},
		{"missing min_kandev_version", "id: demo\nversion: 1.2.3\n", "min_kandev_version"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadManifest(writeManifest(t, tc.body))
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestReadManifestRefusesNonReleaseMinimumVersion(t *testing.T) {
	for _, v := range []string{"0.96", "v0.96.0", "0.96.0-rc.1", "latest"} {
		t.Run(v, func(t *testing.T) {
			_, err := ReadManifest(writeManifest(t, "id: demo\nversion: 1.2.3\nmin_kandev_version: \""+v+"\"\n"))
			require.ErrorContains(t, err, "min_kandev_version")
		})
	}
}

func TestReadManifestRefusesMissingFileAndBadYAML(t *testing.T) {
	_, err := ReadManifest(filepath.Join(t.TempDir(), "absent.yaml"))
	require.Error(t, err)
	_, err = ReadManifest(writeManifest(t, "id: [unclosed\n"))
	require.Error(t, err)
}

func TestReadManifestReadsTheRepositoryManifest(t *testing.T) {
	m, err := ReadManifest("../../manifest.yaml")
	require.NoError(t, err)
	require.Equal(t, "nulab-backlog", m.ID)
	require.Equal(t, "0.96.0", m.MinKandevVersion)
	require.NotEmpty(t, m.Version)
}
