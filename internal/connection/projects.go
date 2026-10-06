package connection

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidProjects is returned for a project selection that is not a list
// of at most 100 Backlog project keys on the connected space.
var ErrInvalidProjects = errors.New("projectKeys must be a list of at most 100 Backlog project keys from this space")

// FieldProjectKeys is the input field reported with ErrInvalidProjects.
const FieldProjectKeys = "projectKeys"

const maxProjects = 100

var projectKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,24}$`)

// ValidateProjectKeys checks the decoded projectKeys JSON value: it trims and
// upper-cases each key and removes duplicates, keeping the first order.
func ValidateProjectKeys(raw any) ([]string, error) {
	list, ok := raw.([]any)
	if !ok || len(list) > maxProjects {
		return nil, ErrInvalidProjects
	}
	out := make([]string, 0, len(list))
	seen := map[string]bool{}
	for _, item := range list {
		s, ok := item.(string)
		key := strings.ToUpper(strings.TrimSpace(s))
		if !ok || !projectKeyPattern.MatchString(key) {
			return nil, ErrInvalidProjects
		}
		if !seen[key] {
			seen[key] = true
			out = append(out, key)
		}
	}
	return out, nil
}
