package ci

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Patterns from kdlbs/kandev plugin-registry/schema.json.
var (
	registryID   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	registryRepo = regexp.MustCompile(`^[^/]+/[^/]+$`)
)

// MarketplaceEntry returns the plugin-registry/plugins.yaml entry for this
// plugin, indented to paste under "plugins:" (AC7.6.1). It never sets
// featured, which is for maintainers only.
func MarketplaceEntry(m Manifest, repo string, categories []string) (string, error) {
	if err := checkEntry(m.ID, repo); err != nil {
		return "", err
	}
	entry := fmt.Sprintf("  - id: %s\n    repo: %s\n", m.ID, repo)
	if len(categories) > 0 {
		entry += fmt.Sprintf("    categories: [%s]\n", strings.Join(categories, ", "))
	}
	return entry, nil
}

func checkEntry(id, repo string) error {
	if !registryID.MatchString(id) {
		return fmt.Errorf("manifest id %q does not match the registry id pattern %s", id, registryID)
	}
	if !registryRepo.MatchString(repo) || strings.HasPrefix(repo, "/") || strings.HasSuffix(repo, "/") {
		return fmt.Errorf("repo %q is not owner/name", repo)
	}
	return nil
}

// CheckRegistry compares a plugins.yaml with this plugin. It fails when the
// entry for repo carries another id (AC7.6.2) or when the manifest id is
// already used by another repo; listed reports whether repo has an entry.
func CheckRegistry(registry []byte, manifestID, repo string) (listed bool, err error) {
	var doc struct {
		Plugins []struct {
			ID   string `yaml:"id"`
			Repo string `yaml:"repo"`
		} `yaml:"plugins"`
	}
	if err := yaml.Unmarshal(registry, &doc); err != nil {
		return false, fmt.Errorf("parse registry: %w", err)
	}
	for _, p := range doc.Plugins {
		switch {
		case p.Repo == repo && p.ID != manifestID:
			return true, fmt.Errorf("registry id %q for %s differs from the manifest id %q", p.ID, repo, manifestID)
		case p.Repo == repo:
			listed = true
		case p.ID == manifestID:
			return false, fmt.Errorf("registry id %q is already used by %s", manifestID, p.Repo)
		}
	}
	return listed, nil
}
