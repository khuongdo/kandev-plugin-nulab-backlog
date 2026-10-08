package scm

import (
	"context"
	"slices"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// ListQueries returns the active service's saved queries; one whose
// repository is no longer mapped is marked unmapped, never deleted (FR3.4).
// Other providers' queries are kept but hidden (FR2.4).
func (s *Service) ListQueries(ctx context.Context, ws string) ([]Query, error) {
	qs, active, err := s.allQueries(ctx, ws)
	return slices.DeleteFunc(qs, func(q Query) bool { return allowed(active, q.Provider) != nil }), err
}

// allQueries returns every saved query, marked unmapped, and the active service.
func (s *Service) allQueries(ctx context.Context, ws string) ([]Query, Provider, error) {
	qs, err := s.store.Queries(ctx, ws)
	if err != nil {
		return nil, "", err
	}
	settings, active, err := s.settingsDoc(ctx, ws)
	if err != nil {
		return nil, "", err
	}
	out := []Query{}
	for _, q := range qs {
		q.Unmapped = !isMapped(settingsOf(settings, q.Provider), q.ProjectKey, q.Repo)
		out = append(out, q)
	}
	return out, active, nil
}

// find returns the index of item id in list, refusing an item of a
// non-active provider (FR1.4).
func find[T any](list []T, id string, active Provider, input func(T) QueryInput) (int, error) {
	i := slices.IndexFunc(list, func(x T) bool { return input(x).ID == id })
	if i < 0 {
		return i, ErrNotFound
	}
	return i, allowed(active, input(list[i]).Provider)
}

func queryInput(q Query) QueryInput   { return q.QueryInput }
func watchInputOf(w Watch) QueryInput { return w.QueryInput }

// requireMapped refuses a repository that is not mapped to the project (FR3.3).
func (s *Service) requireMapped(ctx context.Context, ws string, q QueryInput) error {
	st, err := s.settings(ctx, ws, q.Provider)
	if err != nil {
		return err
	}
	if !isMapped(st, q.ProjectKey, q.Repo) {
		return &connection.FieldError{Field: FieldRepository, Err: errUnmapped}
	}
	return nil
}

// SaveQuery creates (no id) or replaces a saved query; a replacement keeps
// its default star (FR4.2).
func (s *Service) SaveQuery(ctx context.Context, ws string, in QueryInput) (Query, error) {
	if err := in.Validate(); err != nil {
		return Query{}, err
	}
	if err := s.requireMapped(ctx, ws, in); err != nil {
		return Query{}, err
	}
	out := Query{QueryInput: in}
	err := s.store.UpdateQueries(ctx, ws, func(list []Query) ([]Query, error) {
		if out.ID == "" {
			out.ID = newID()
			return append(list, out), nil
		}
		i := slices.IndexFunc(list, func(q Query) bool { return q.ID == out.ID })
		if i < 0 {
			return nil, ErrNotFound
		}
		out.IsDefault = list[i].IsDefault
		list[i] = out
		return list, nil
	})
	return out, err
}

// SetQueryDefault stars or un-stars a query; starring clears the other
// stars of the same provider, so each provider has at most one (FR4.2).
func (s *Service) SetQueryDefault(ctx context.Context, ws, id string, isDefault bool) ([]Query, error) {
	active, err := s.Active(ctx, ws)
	if err != nil {
		return nil, err
	}
	err = s.store.UpdateQueries(ctx, ws, func(list []Query) ([]Query, error) {
		i, err := find(list, id, active, queryInput)
		if err != nil {
			return nil, err
		}
		for j := range list {
			if isDefault && list[j].Provider == list[i].Provider {
				list[j].IsDefault = false
			}
		}
		list[i].IsDefault = isDefault
		return list, nil
	})
	if err != nil {
		return nil, err
	}
	return s.ListQueries(ctx, ws)
}

// DeleteQuery removes a saved query.
func (s *Service) DeleteQuery(ctx context.Context, ws, id string) error {
	active, err := s.Active(ctx, ws)
	if err != nil {
		return err
	}
	return s.store.UpdateQueries(ctx, ws, func(list []Query) ([]Query, error) {
		i, err := find(list, id, active, queryInput)
		if err != nil {
			return nil, err
		}
		return slices.Delete(list, i, i+1), nil
	})
}

// RunQuery returns the first page of a saved query (FR4.2). An unmapped
// query cannot run (conflict).
func (s *Service) RunQuery(ctx context.Context, ws, id string) (PRListPage, error) {
	qs, active, err := s.allQueries(ctx, ws)
	if err != nil {
		return PRListPage{}, err
	}
	i, err := find(qs, id, active, queryInput)
	switch {
	case err != nil:
		return PRListPage{}, err
	case qs[i].Unmapped:
		return PRListPage{}, ErrConflict
	}
	st, err := s.settings(ctx, ws, qs[i].Provider)
	if err != nil {
		return PRListPage{}, err
	}
	return s.listPRs(ctx, ws, st, qs[i].QueryInput, 1)
}

// ListWatches returns the active service's watches, marked unmapped like
// queries (FR3.4); other providers' watches are kept but hidden (FR2.4).
func (s *Service) ListWatches(ctx context.Context, ws string) ([]Watch, error) {
	list, active, err := s.allWatches(ctx, ws)
	return slices.DeleteFunc(list, func(w Watch) bool { return allowed(active, w.Provider) != nil }), err
}

// allWatches returns every watch, marked unmapped, and the active service.
func (s *Service) allWatches(ctx context.Context, ws string) ([]Watch, Provider, error) {
	list, err := s.store.Watches(ctx, ws)
	if err != nil {
		return nil, "", err
	}
	settings, active, err := s.settingsDoc(ctx, ws)
	if err != nil {
		return nil, "", err
	}
	out := []Watch{}
	for _, w := range list {
		w.Unmapped = !isMapped(settingsOf(settings, w.Provider), w.ProjectKey, w.Repo)
		out = append(out, w)
	}
	return out, active, nil
}

// SaveWatch creates (no id) or edits a watch; an edit keeps its state,
// count and last run (FR4.3).
func (s *Service) SaveWatch(ctx context.Context, ws string, in WatchInput) (Watch, error) {
	if err := in.Validate(); err != nil {
		return Watch{}, err
	}
	if err := s.requireMapped(ctx, ws, in.QueryInput); err != nil {
		return Watch{}, err
	}
	out := Watch{WatchInput: in, State: WatchActive}
	err := s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		if out.ID == "" {
			out.ID = newID()
			return append(list, out), nil
		}
		i := slices.IndexFunc(list, func(w Watch) bool { return w.ID == out.ID })
		if i < 0 {
			return nil, ErrNotFound
		}
		out.State, out.CreatedCount, out.LastRunAt = list[i].State, list[i].CreatedCount, list[i].LastRunAt
		list[i] = out
		return list, nil
	})
	return out, err
}

// DeleteWatch removes a watch; its ledger and created tasks stay.
func (s *Service) DeleteWatch(ctx context.Context, ws, id string) error {
	active, err := s.Active(ctx, ws)
	if err != nil {
		return err
	}
	return s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		i, err := find(list, id, active, watchInputOf)
		if err != nil {
			return nil, err
		}
		return slices.Delete(list, i, i+1), nil
	})
}

// PauseWatch stops a watch from running.
func (s *Service) PauseWatch(ctx context.Context, ws, id string) (Watch, error) {
	return s.setWatchState(ctx, ws, id, WatchPaused)
}

// ResumeWatch lets a paused watch run again.
func (s *Service) ResumeWatch(ctx context.Context, ws, id string) (Watch, error) {
	return s.setWatchState(ctx, ws, id, WatchActive)
}

func (s *Service) setWatchState(ctx context.Context, ws, id, state string) (Watch, error) {
	active, err := s.Active(ctx, ws)
	if err != nil {
		return Watch{}, err
	}
	var out Watch
	err = s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		i, err := find(list, id, active, watchInputOf)
		if err != nil {
			return nil, err
		}
		list[i].State = state
		out = list[i]
		return list, nil
	})
	return out, err
}
