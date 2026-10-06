package backlog

import "net/http"

// Group is Backlog's rate-limit group of a call (ADR-003).
type Group int

// Rate-limit groups. Search and Update calls are queued per space host.
const (
	GroupRead Group = iota + 1
	GroupUpdate
	GroupSearch
)

var groupNames = map[Group]string{GroupRead: "read", GroupUpdate: "update", GroupSearch: "search"}

func (g Group) String() string {
	if n, ok := groupNames[g]; ok {
		return n
	}
	return "unknown"
}

// searchPaths are the GET calls Backlog counts in the Search group.
// ponytail: only the issue list and count are used; add the wiki paths when a unit calls them.
var searchPaths = map[string]bool{"/api/v2/issues": true, "/api/v2/issues/count": true}

// group returns the rate-limit group of a call.
func group(method, path string) Group {
	switch {
	case method != http.MethodGet:
		return GroupUpdate
	case searchPaths[path]:
		return GroupSearch
	default:
		return GroupRead
	}
}
