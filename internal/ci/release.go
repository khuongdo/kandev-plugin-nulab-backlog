package ci

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// tagPattern is a vX.Y.Z semver tag with an optional pre-release suffix and
// no build metadata (team Deployment).
var tagPattern = regexp.MustCompile(`^v((?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?)$`)

// Tag is a parsed release tag.
type Tag struct {
	Name       string // the tag as given, for example v0.1.0
	Version    string // the version without "v", for example 0.1.0
	Prerelease bool   // true when the tag has a -suffix
}

// ParseTag parses a vX.Y.Z[-pre] release tag.
func ParseTag(name string) (Tag, error) {
	m := tagPattern.FindStringSubmatch(name)
	if m == nil {
		return Tag{}, fmt.Errorf("tag %q is not a vX.Y.Z semver tag", name)
	}
	return Tag{Name: name, Version: m[1], Prerelease: m[2] != ""}, nil
}

// Release is one GitHub Release, as `gh release list --json
// tagName,isDraft,isPrerelease` prints it.
type Release struct {
	TagName      string `json:"tagName"`
	IsDraft      bool   `json:"isDraft"`
	IsPrerelease bool   `json:"isPrerelease"`
}

// ParseReleases parses the `gh release list` JSON. Anything but a JSON array,
// for example the empty output of a failed gh call, is an error: an unknown
// list must never read as "no Release" (fail closed).
func ParseReleases(raw string) ([]Release, error) {
	var rs []Release
	if err := json.Unmarshal([]byte(raw), &rs); err != nil || rs == nil {
		return nil, errUnknownReleases
	}
	return rs, nil
}

var errUnknownReleases = errors.New("could not list GitHub Releases; refusing to release without knowing which exist")

// Facts are the release preflight inputs. The git and gh facts are passed as
// plain values, so the check itself runs no process.
type Facts struct {
	Tag             string    // the pushed tag, for example v0.1.0
	ManifestVersion string    // manifest.yaml version
	OnMain          bool      // the tagged commit is an ancestor of origin/main
	Releases        []Release // every GitHub Release; nil means the list is unknown
	ManualChecksDir string    // docs/manual-checks
}

// firstReleaseResult is the passing Result row of a manual-check record
// (docs/manual-checks/TEMPLATE.md).
var firstReleaseResult = regexp.MustCompile(`(?m)^\|\s*Result\s*\|\s*pass\s*\|\s*$`)

// CheckRelease refuses a release whose tag is malformed, differs from the
// manifest version, is not on main or already has a Release; and, for the
// first stable release (no published, non-pre-release GitHub Release yet),
// one without a passing first-release record (AC7.5.2, AC7.5.3). An unknown
// Release list is refused. Every refusal is one line.
func CheckRelease(f Facts) error {
	tag, err := ParseTag(f.Tag)
	if err != nil {
		return err
	}
	switch {
	case f.Releases == nil:
		return errUnknownReleases
	case tag.Version != f.ManifestVersion:
		return fmt.Errorf("tag %s does not match the manifest version %q", f.Tag, f.ManifestVersion)
	case !f.OnMain:
		return fmt.Errorf("tag %s points at a commit that is not on main", f.Tag)
	case hasRelease(f.Releases, f.Tag):
		return fmt.Errorf("tag %s already has a GitHub Release; a released tag is never overwritten", f.Tag)
	}
	if tag.Prerelease || hasPublishedStableRelease(f.Releases, f.Tag) {
		return nil
	}
	ok, err := FirstReleaseRecorded(f.ManualChecksDir)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("tag %s is the first release and %s has no *-first-release.md record with Result pass", f.Tag, f.ManualChecksDir)
	}
	return nil
}

// hasRelease reports whether any Release, a draft included, has the tag.
func hasRelease(rs []Release, tag string) bool {
	for _, r := range rs {
		if r.TagName == tag {
			return true
		}
	}
	return false
}

// hasPublishedStableRelease reports whether a Release other than current is
// neither a draft nor a pre-release. A tag the preflight refused has no
// Release, so it never counts (R-01).
func hasPublishedStableRelease(rs []Release, current string) bool {
	for _, r := range rs {
		if r.TagName != current && !r.IsDraft && !r.IsPrerelease {
			return true
		}
	}
	return false
}

// FirstReleaseRecorded reports whether dir holds a *-first-release.md
// manual-check record whose Result is pass. A missing dir is "no record".
func FirstReleaseRecorded(dir string) (bool, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*-first-release.md"))
	if err != nil {
		return false, fmt.Errorf("list manual-check records: %w", err)
	}
	for _, p := range paths {
		raw, err := os.ReadFile(p) //nolint:gosec // G304: records live in the repository's docs/manual-checks
		if err != nil {
			return false, fmt.Errorf("read manual-check record: %w", err)
		}
		if firstReleaseResult.Match(raw) {
			return true, nil
		}
	}
	return false, nil
}
