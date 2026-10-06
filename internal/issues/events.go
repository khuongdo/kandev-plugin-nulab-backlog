package issues

import (
	"context"
	"errors"
	"slices"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Listen subscribes the service to ConnectionChanged; call the returned
// func to stop.
func (s *Service) Listen() (unsubscribe func()) {
	return s.conn.Subscribe(func(e connection.ConnectionChanged) { s.OnConnectionChanged(context.Background(), e) })
}

// target is the connection links must match to stay active.
type target struct {
	host     string
	selected []string
}

// OnConnectionChanged turns links off when the connection no longer covers
// them (disconnect AC1.5.4, another space AC1.8.2, a deselected project
// AC1.9.2), and brings covered links back on a restore (M12). An event
// older than the last one handled is ignored.
func (s *Service) OnConnectionChanged(ctx context.Context, e connection.ConnectionChanged) {
	s.mu.Lock()
	if e.ConnectionEpoch < s.lastEpoch[e.WorkspaceID] {
		s.mu.Unlock()
		return
	}
	s.lastEpoch[e.WorkspaceID] = e.ConnectionEpoch
	s.mu.Unlock()
	var t *target
	if e.Reason != connection.ReasonDisconnected {
		t = &target{host: e.SpaceHost, selected: e.SelectedProjects}
	}
	if err := s.apply(ctx, e.WorkspaceID, t, e.Restore); err != nil {
		redact.Logger(ctx).ErrorContext(ctx, "issue links not updated", "event", "issue_links_update_failed",
			"workspaceId", e.WorkspaceID, "reason", string(e.Reason))
	}
}

// ReconcileAll applies the current connection to every workspace with links
// (startup, and the start of every sync cycle).
func (s *Service) ReconcileAll(ctx context.Context) error {
	idx, err := s.store.Index(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, ws := range idx {
		errs = append(errs, s.reconcile(ctx, ws))
	}
	return errors.Join(errs...)
}

func (s *Service) reconcile(ctx context.Context, ws string) error {
	snap, err := s.conn.Current(ctx, ws)
	switch {
	case errors.Is(err, connection.ErrNotConnected), errors.Is(err, connection.ErrReconnectRequired):
		return s.apply(ctx, ws, nil, false)
	case err != nil:
		return err
	}
	return s.apply(ctx, ws, &target{host: snap.SpaceHost, selected: snap.SelectedProjects}, true)
}

// apply turns off links t does not cover (all of them when t is nil) and,
// on restore, brings covered not-connected links back.
func (s *Service) apply(ctx context.Context, ws string, t *target, restore bool) error {
	return s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		changed := false
		for i, l := range links {
			covered := t != nil && l.SpaceHost == t.host && slices.Contains(t.selected, l.ProjectKey)
			next := l.State
			switch {
			case !covered:
				next = StateNotConnected
			case restore:
				next = StateActive
			}
			if next != l.State {
				links[i].State, changed = next, true
			}
		}
		if !changed {
			return nil, errUnchanged
		}
		return links, nil
	})
}

// OnTaskDeleted removes the deleted task's links; deleting twice is fine (R-05).
func (s *Service) OnTaskDeleted(ctx context.Context, ws, taskID string) error {
	return s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		n := len(links)
		links = slices.DeleteFunc(links, func(l Link) bool { return l.TaskID == taskID })
		if len(links) == n {
			return nil, errUnchanged
		}
		return links, nil
	})
}
