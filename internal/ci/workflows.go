package ci

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"

	"gopkg.in/yaml.v3"
)

// pinnedUses is owner/repo[/path]@<full 40-hex commit SHA> (team Deployment).
var pinnedUses = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(?:/[^@\s]+)?@[0-9a-f]{40}$`)

// The only job allowed any write permission: it creates the Release and the
// build provenance attestation.
const (
	publishFile = "release.yml"
	publishJob  = "publish"
)

type workflow struct {
	On          yaml.Node `yaml:"on"`
	Permissions yaml.Node `yaml:"permissions"`
	Jobs        map[string]struct {
		Permissions yaml.Node `yaml:"permissions"`
		Uses        string    `yaml:"uses"`
		Steps       []struct {
			Uses string `yaml:"uses"`
		} `yaml:"steps"`
	} `yaml:"jobs"`
}

// CheckWorkflows enforces the workflow policy on every *.yml and *.yaml file
// in dir: every action pinned to a full commit SHA; top-level permissions
// exactly contents: read; no pull_request_target trigger; and no write
// permission in any job except publish in release.yml.
func CheckWorkflows(dir string) error {
	paths, err := filepath.Glob(filepath.Join(dir, "*.y*ml"))
	if err != nil {
		return fmt.Errorf("list workflows: %w", err)
	}
	if len(paths) == 0 {
		return fmt.Errorf("no workflow files in %s", dir)
	}
	var errs []error
	for _, p := range paths {
		errs = append(errs, checkWorkflow(p)...)
	}
	return errors.Join(errs...)
}

func checkWorkflow(path string) []error {
	name := filepath.Base(path)
	raw, err := os.ReadFile(path) //nolint:gosec // G304: the repository's own workflow files
	if err != nil {
		return []error{fmt.Errorf("%s: %w", name, err)}
	}
	var wf workflow
	if err := yaml.Unmarshal(raw, &wf); err != nil {
		return []error{fmt.Errorf("%s: parse: %w", name, err)}
	}
	var errs []error
	fail := func(format string, args ...any) { errs = append(errs, fmt.Errorf(name+": "+format, args...)) }
	if !onlyContentsRead(&wf.Permissions) {
		fail("top-level permissions must be exactly contents: read")
	}
	if triggers(&wf.On, "pull_request_target") {
		fail("pull_request_target is not allowed; use pull_request")
	}
	jobs := make([]string, 0, len(wf.Jobs))
	for j := range wf.Jobs {
		jobs = append(jobs, j)
	}
	sort.Strings(jobs)
	for _, j := range jobs {
		job := wf.Jobs[j]
		if grantsWrite(&job.Permissions) && (name != publishFile || j != publishJob) {
			fail("job %s has a write permission; only %s in %s may", j, publishJob, publishFile)
		}
		uses := []string{job.Uses}
		for _, s := range job.Steps {
			uses = append(uses, s.Uses)
		}
		for _, u := range uses {
			if u != "" && !pinnedUses.MatchString(u) {
				fail("job %s: %q is not pinned to a full commit SHA", j, u)
			}
		}
	}
	return errs
}

func onlyContentsRead(n *yaml.Node) bool {
	return n.Kind == yaml.MappingNode && len(n.Content) == 2 &&
		n.Content[0].Value == "contents" && n.Content[1].Value == "read"
}

func grantsWrite(n *yaml.Node) bool {
	switch n.Kind {
	case yaml.ScalarNode:
		return n.Value == "write-all"
	case yaml.MappingNode:
		for i := 1; i < len(n.Content); i += 2 {
			if n.Content[i].Value == "write" {
				return true
			}
		}
	}
	return false
}

// triggers reports whether the on: node (a string, a list or a map) names event.
func triggers(n *yaml.Node, event string) bool {
	switch n.Kind {
	case yaml.ScalarNode:
		return n.Value == event
	case yaml.SequenceNode:
		return slices.ContainsFunc(n.Content, func(c *yaml.Node) bool { return c.Value == event })
	case yaml.MappingNode:
		for i := 0; i < len(n.Content); i += 2 {
			if n.Content[i].Value == event {
				return true
			}
		}
	}
	return false
}
