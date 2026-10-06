package backlog

import (
	"encoding/json"
	"net/url"
)

// Repository is one entry of GET /api/v2/projects/:key/git/repositories.
// Backlog gives no default branch.
type Repository struct {
	ID        int64
	ProjectID int64
	Name      string
	HTTPURL   string // https clone URL on the space host, no credentials
}

// parseRepositories decodes the repository list. An entry without a numeric
// id or a name, or whose httpUrl is not a credential-free https URL on host,
// makes the whole body Unreachable.
func parseRepositories(body []byte, host string) ([]Repository, error) {
	var raw []struct {
		ID        *int64 `json:"id"`
		ProjectID int64  `json:"projectId"`
		Name      string `json:"name"`
		HTTPURL   string `json:"httpUrl"`
	}
	bad := &Error{Kind: KindUnreachable, Status: 200, Class: "body"}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, bad
	}
	out := make([]Repository, 0, len(raw))
	for _, r := range raw {
		if r.ID == nil || r.Name == "" || !sameOriginHTTPS(r.HTTPURL, host) {
			return nil, bad
		}
		out = append(out, Repository{ID: *r.ID, ProjectID: r.ProjectID, Name: r.Name, HTTPURL: r.HTTPURL})
	}
	return out, nil
}

// sameOriginHTTPS reports whether raw is https://host/... with no user info and no port.
func sameOriginHTTPS(raw, host string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.User == nil && u.Host == host
}
