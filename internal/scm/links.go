package scm

import (
	"context"
	"slices"
	"strings"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Link resolves a pasted pull request URL with one provider call and links
// it to the task (FR5.1). The repository must be mapped. Linking twice is a
// no-op.
func (s *Service) Link(ctx context.Context, ws, taskID, rawURL string) (Link, error) {
	ref, err := ParsePRURL(rawURL)
	if err != nil {
		return Link{}, err
	}
	if taskID == "" {
		return Link{}, invalid(FieldURL)
	}
	st, err := s.settings(ctx, ws, ref.Provider)
	if err != nil {
		return Link{}, err
	}
	projects := mappedProjects(st, ref.Repo)
	if len(projects) == 0 {
		return Link{}, &connection.FieldError{Field: FieldURL, Err: errUnmapped}
	}
	ctx, cred, err := s.credential(ctx, ws, ref.Provider)
	if err != nil {
		return Link{}, err
	}
	pr, err := s.clients[ref.Provider].GetPR(ctx, cred, ref.Repo, ref.Number)
	if IsStatus(err, 404) {
		return Link{}, ErrNotFound
	}
	if err != nil {
		return Link{}, err
	}
	l := Link{PRRef: ref, TaskID: taskID, ProjectKey: projects[0], Title: pr.Title, State: pr.State, URL: PRURL(ref)}
	return l, s.putLink(ctx, ws, l)
}

// putLink adds l, or replaces the same task's link to the same pull request.
func (s *Service) putLink(ctx context.Context, ws string, l Link) error {
	return s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		i := slices.IndexFunc(links, func(x Link) bool { return !x.Auto && x.TaskID == l.TaskID && x.Key() == l.Key() })
		if i >= 0 {
			links[i] = l
			return links, nil
		}
		return append(links, l), nil
	})
}

// Unlink removes the task's manual link to key (no issueKey), or the
// auto-link of key to issueKey; never both. A removed auto-link is dismissed and never created again (FR5.3).
func (s *Service) Unlink(ctx context.Context, ws, taskID, key, issueKey string) error {
	if p, _, _ := strings.Cut(key, "|"); slices.Contains(Providers, Provider(p)) {
		if err := s.Allow(ctx, ws, Provider(p)); err != nil {
			return err
		}
	}
	removedAuto := false
	err := s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		n := len(links)
		links = slices.DeleteFunc(links, func(l Link) bool {
			match := l.Key() == key && !l.Auto && issueKey == "" && taskID != "" && l.TaskID == taskID
			if issueKey != "" {
				match = l.Key() == key && l.Auto && l.IssueKey == issueKey
			}
			removedAuto = removedAuto || (match && l.Auto)
			return match
		})
		if len(links) == n {
			return nil, ErrNotFound
		}
		return links, nil
	})
	if err != nil || !removedAuto {
		return err
	}
	return s.store.UpdateDismissed(ctx, ws, func(d []string) ([]string, error) {
		return append(d, key+"|"+issueKey), nil
	})
}

// Links returns the task's links and the issue's auto-links (the issue
// panel), of the active service only (FR1.5).
func (s *Service) Links(ctx context.Context, ws, taskID, issueKey string) ([]Link, error) {
	active, err := s.Active(ctx, ws)
	if err != nil {
		return nil, err
	}
	links, err := s.store.Links(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := []Link{}
	for _, l := range links {
		if allowed(active, l.Provider) != nil {
			continue
		}
		if (taskID != "" && l.TaskID == taskID) || (issueKey != "" && l.IssueKey == issueKey) {
			out = append(out, l)
		}
	}
	return out, nil
}

// autoLink links each issue key of a selected project mapped to repo that
// appears in a pull request's source branch or title (FR5.2, A4), unless
// that link was dismissed (FR5.3). It is best effort: a failure is logged
// and never fails the list or watch run that saw the pull requests.
func (s *Service) autoLink(ctx context.Context, ws string, st Settings, repo string, prs []PullRequest) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return // no Backlog connection: no selected projects to match
	}
	projects := slices.DeleteFunc(mappedProjects(st, repo), func(p string) bool {
		return !slices.Contains(snap.SelectedProjects, p)
	})
	var found []Link
	for _, pr := range prs {
		for _, key := range IssueKeys([]string{pr.SourceBranch, pr.Title}, projects) {
			project, _, _ := strings.Cut(key, "-")
			found = append(found, Link{PRRef: pr.PRRef, IssueKey: key, Auto: true, ProjectKey: project,
				Title: pr.Title, State: pr.State, URL: PRURL(pr.PRRef)})
		}
	}
	if len(found) == 0 {
		return
	}
	dismissed, err := s.store.Dismissed(ctx, ws)
	if err == nil {
		err = s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
			n := len(links)
			for _, f := range found {
				known := slices.ContainsFunc(links, func(l Link) bool { return l.Auto && l.Key() == f.Key() && l.IssueKey == f.IssueKey })
				if !known && !slices.Contains(dismissed, f.Key()+"|"+f.IssueKey) {
					links = append(links, f)
				}
			}
			if len(links) == n {
				return nil, errUnchanged
			}
			return links, nil
		})
	}
	if err != nil {
		redact.Logger(ctx).WarnContext(ctx, "auto-link failed", "event", "scm_autolink_failed", "provider", string(st.Provider))
	}
}
