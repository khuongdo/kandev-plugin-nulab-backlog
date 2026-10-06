package backlog

import "encoding/json"

// Project is one entry of GET /api/v2/projects.
type Project struct {
	ID       int64
	Key      string
	Name     string
	Archived bool
}

// parseProjects decodes the project list. An entry without a numeric id or
// a key makes the whole body Unreachable.
func parseProjects(body []byte) ([]Project, error) {
	var raw []struct {
		ID         *int64 `json:"id"`
		ProjectKey string `json:"projectKey"`
		Name       string `json:"name"`
		Archived   bool   `json:"archived"`
	}
	bad := &Error{Kind: KindUnreachable, Status: 200, Class: "body"}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, bad
	}
	out := make([]Project, 0, len(raw))
	for _, p := range raw {
		if p.ID == nil || p.ProjectKey == "" {
			return nil, bad
		}
		out = append(out, Project{ID: *p.ID, Key: p.ProjectKey, Name: p.Name, Archived: p.Archived})
	}
	return out, nil
}
