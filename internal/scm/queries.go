package scm

import (
	"context"
	"slices"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// ListQueries returns the saved queries; one whose repository is no longer
// mapped is marked unmapped, never deleted (FR3.4).
func (s *Service) ListQueries(ctx context.Context, ws string) ([]Query, error) {
	qs, err := s.store.Queries(ctx, ws)
	if err != nil {
		return nil, err
	}
	settings, err := s.store.Settings(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := []Query{}
	for _, q := range qs {
		q.Unmapped = !isMapped(settingsOf(settings, q.Provider), q.ProjectKey, q.Repo)
		out = append(out, q)
	}
	return out, nil
}

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
	err := s.store.UpdateQueries(ctx, ws, func(list []Query) ([]Query, error) {
		i := slices.IndexFunc(list, func(q Query) bool { return q.ID == id })
		if i < 0 {
			return nil, ErrNotFound
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
	return s.store.UpdateQueries(ctx, ws, func(list []Query) ([]Query, error) {
		n := len(list)
		list = slices.DeleteFunc(list, func(q Query) bool { return q.ID == id })
		if len(list) == n {
			return nil, ErrNotFound
		}
		return list, nil
	})
}

// RunQuery returns the first page of a saved query (FR4.2). An unmapped
// query cannot run (conflict).
func (s *Service) RunQuery(ctx context.Context, ws, id string) (PRListPage, error) {
	qs, err := s.ListQueries(ctx, ws)
	if err != nil {
		return PRListPage{}, err
	}
	i := slices.IndexFunc(qs, func(q Query) bool { return q.ID == id })
	switch {
	case i < 0:
		return PRListPage{}, ErrNotFound
	case qs[i].Unmapped:
		return PRListPage{}, ErrConflict
	}
	st, err := s.settings(ctx, ws, qs[i].Provider)
	if err != nil {
		return PRListPage{}, err
	}
	return s.listPRs(ctx, ws, st, qs[i].QueryInput, 1)
}

// ListWatches returns the watches, marked unmapped like queries (FR3.4).
func (s *Service) ListWatches(ctx context.Context, ws string) ([]Watch, error) {
	list, err := s.store.Watches(ctx, ws)
	if err != nil {
		return nil, err
	}
	settings, err := s.store.Settings(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := []Watch{}
	for _, w := range list {
		w.Unmapped = !isMapped(settingsOf(settings, w.Provider), w.ProjectKey, w.Repo)
		out = append(out, w)
	}
	return out, nil
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
	return s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		n := len(list)
		list = slices.DeleteFunc(list, func(w Watch) bool { return w.ID == id })
		if len(list) == n {
			return nil, ErrNotFound
		}
		return list, nil
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
	var out Watch
	err := s.store.UpdateWatches(ctx, ws, func(list []Watch) ([]Watch, error) {
		i := slices.IndexFunc(list, func(w Watch) bool { return w.ID == id })
		if i < 0 {
			return nil, ErrNotFound
		}
		list[i].State = state
		out = list[i]
		return list, nil
	})
	return out, err
}
