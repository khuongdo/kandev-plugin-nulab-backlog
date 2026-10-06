package ci

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// releaseVersion is Kandev's release version shape for min_kandev_version:
// numeric X.Y.Z with no "v", suffix or build metadata.
var releaseVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// Manifest holds the manifest.yaml facts the CI and release checks use.
type Manifest struct {
	ID               string `yaml:"id"`
	Version          string `yaml:"version"`
	MinKandevVersion string `yaml:"min_kandev_version"`
}

// ReadManifest reads path and fails unless id, version and a numeric X.Y.Z
// min_kandev_version are present.
func ReadManifest(path string) (Manifest, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // G304: the path is the repository's own manifest
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	switch {
	case m.ID == "":
		return Manifest{}, fmt.Errorf("manifest: id is missing")
	case m.Version == "":
		return Manifest{}, fmt.Errorf("manifest: version is missing")
	case m.MinKandevVersion == "":
		return Manifest{}, fmt.Errorf("manifest: min_kandev_version is missing")
	case !releaseVersion.MatchString(m.MinKandevVersion):
		return Manifest{}, fmt.Errorf("manifest: min_kandev_version %q is not a numeric X.Y.Z release version", m.MinKandevVersion)
	}
	return m, nil
}
