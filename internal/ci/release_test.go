package ci

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTag(t *testing.T) {
	cases := []struct {
		tag        string
		ok         bool
		version    string
		prerelease bool
	}{
		{"v0.1.0", true, "0.1.0", false},
		{"v1.20.3", true, "1.20.3", false},
		{"v0.0.1-rc.1", true, "0.0.1-rc.1", true},
		{"0.1.0", false, "", false},
		{"v1.2", false, "", false},
		{"v01.2.3", false, "", false},
		{"v1.2.3+build", false, "", false},
		{"v1.2.3 ", false, "", false},
		{"refs/tags/v1.2.3", false, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			got, err := ParseTag(tc.tag)
			if !tc.ok {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, Tag{Name: tc.tag, Version: tc.version, Prerelease: tc.prerelease}, got)
		})
	}
}

func passingRecordDir(t *testing.T, result string) string {
	t.Helper()
	return writeTree(t, map[string]string{
		"2026-10-20-first-release.md": "| Field | Value |\n|-------|-------|\n| Check name | first-release |\n| Result | " + result + " |\n",
	})
}

func validFacts(t *testing.T) Facts {
	return Facts{Tag: "v0.1.0", ManifestVersion: "0.1.0", OnMain: true, Releases: []Release{}, ManualChecksDir: passingRecordDir(t, "pass")}
}

func TestCheckReleasePassesForAValidFirstRelease(t *testing.T) {
	require.NoError(t, CheckRelease(validFacts(t)))
}

func TestCheckReleaseRefusals(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Facts)
		want   string
	}{
		{"malformed tag", func(f *Facts) { f.Tag = "0.1.0" }, "not a vX.Y.Z"},
		{"tag differs from manifest", func(f *Facts) { f.Tag = "v0.1.1" }, "manifest version"},
		{"commit not on main", func(f *Facts) { f.OnMain = false }, "not on main"},
		{"release already exists", func(f *Facts) { f.Releases = []Release{{TagName: "v0.1.0"}} }, "already has a GitHub Release"},
		{"draft release already exists", func(f *Facts) { f.Releases = []Release{{TagName: "v0.1.0", IsDraft: true}} }, "already has a GitHub Release"},
		{"release list unknown", func(f *Facts) { f.Releases = nil }, "could not list GitHub Releases"},
		{"first release without a record", func(f *Facts) { f.ManualChecksDir = t.TempDir() }, "first-release"},
		{"first release with a failing record", func(f *Facts) { f.ManualChecksDir = passingRecordDir(t, "fail") }, "first-release"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := validFacts(t)
			tc.mutate(&f)
			err := CheckRelease(f)
			require.ErrorContains(t, err, tc.want)
			require.NotContains(t, err.Error(), "\n", "a refusal names its reason in one line")
		})
	}
}

func TestCheckReleaseSkipsTheRecordForLaterAndPreReleases(t *testing.T) {
	published := []Release{{TagName: "v0.0.1-rc.1", IsPrerelease: true}, {TagName: "v0.1.0"}}
	later := Facts{Tag: "v0.2.0", ManifestVersion: "0.2.0", OnMain: true, Releases: published, ManualChecksDir: t.TempDir()}
	require.NoError(t, CheckRelease(later))
	pre := Facts{Tag: "v0.0.1-rc.1", ManifestVersion: "0.0.1-rc.1", OnMain: true, Releases: []Release{}, ManualChecksDir: t.TempDir()}
	require.NoError(t, CheckRelease(pre))
}

// R-01: a stable tag that the preflight refused (so no Release was created) is
// not an earlier release; nor are drafts and pre-releases.
func TestCheckReleaseStillRequiresTheRecordWithoutAPublishedStableRelease(t *testing.T) {
	cases := map[string][]Release{
		"refused v0.1.0 then retagged v0.1.1": {},
		"only drafts and pre-releases":        {{TagName: "v0.1.0", IsDraft: true}, {TagName: "v0.0.1-rc.1", IsPrerelease: true}},
	}
	for name, releases := range cases {
		t.Run(name, func(t *testing.T) {
			f := Facts{Tag: "v0.1.1", ManifestVersion: "0.1.1", OnMain: true, Releases: releases, ManualChecksDir: t.TempDir()}
			require.ErrorContains(t, CheckRelease(f), "first release")
		})
	}
}

func TestParseReleases(t *testing.T) {
	got, err := ParseReleases(`[{"tagName":"v0.1.0","isDraft":false,"isPrerelease":false},{"tagName":"v0.2.0-rc.1","isDraft":true,"isPrerelease":true}]`)
	require.NoError(t, err)
	require.Equal(t, []Release{{TagName: "v0.1.0"}, {TagName: "v0.2.0-rc.1", IsDraft: true, IsPrerelease: true}}, got)

	got, err = ParseReleases("[]\n")
	require.NoError(t, err)
	require.Empty(t, got)
	require.NotNil(t, got, "an empty list is known, not unknown")

	for _, raw := range []string{"", "null", "HTTP 401: Bad credentials", "{}"} {
		_, err := ParseReleases(raw)
		require.ErrorContains(t, err, "could not list GitHub Releases", "%q", raw)
	}
}

func TestFirstReleaseRecordedFailsOnMissingDir(t *testing.T) {
	ok, err := FirstReleaseRecorded(filepath.Join(t.TempDir(), "absent"))
	require.NoError(t, err)
	require.False(t, ok)
}
