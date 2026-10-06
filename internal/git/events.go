package git

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

// target is the connection items must match to stay active.
type target struct {
	host     string
	selected []string
}

// OnConnectionChanged turns items off when the connection no longer covers
// them, and brings them back on a restore: links active, watches Paused
// (M12). An event older than the last one handled is ignored.
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
		redact.Logger(ctx).ErrorContext(ctx, "git items not updated", "event", "git_items_update_failed",
			"workspaceId", e.WorkspaceID, "reason", string(e.Reason))
	}
}

// ReconcileAll applies the current connection to every workspace with
// links or watches (startup).
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

// apply turns off items t does not cover (all of them when t is nil) and,
// on restore, brings covered not-connected items back.
func (s *Service) apply(ctx context.Context, ws string, t *target, restore bool) error {
	covered := func(host, project string) bool {
		return t != nil && host == t.host && slices.Contains(t.selected, project)
	}
	next := func(state, covered bool, back string) string {
		switch {
		case !covered:
			return StatusNotConnected
		case restore && state:
			return back
		}
		return ""
	}
	err := s.store.UpdateLinks(ctx, ws, func(links []Link) ([]Link, error) {
		changed := false
		for i, l := range links {
			if st := next(l.Status == StatusNotConnected, covered(l.SpaceHost, l.ProjectKey), StatusActive); st != "" && st != l.Status {
				links[i].Status, changed = st, true
			}
		}
		if !changed {
			return nil, errUnchanged
		}
		return links, nil
	})
	if err != nil {
		return err
	}
	return s.store.UpdateWatches(ctx, ws, func(watches []Watch) ([]Watch, error) {
		changed := false
		for i, w := range watches {
			if st := next(w.State == StatusNotConnected, covered(w.SpaceHost, w.ProjectKey), StatusPaused); st != "" && st != w.State {
				watches[i].State, changed = st, true
			}
		}
		if !changed {
			return nil, errUnchanged
		}
		return watches, nil
	})
}
