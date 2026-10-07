package issues

import (
	"context"
	"errors"
	"slices"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

const (
	keyWatches       = "issues.watches"
	keyWatchIndex    = "issues.watch_index" // instance key: workspaces with issue watches
	keyLedgerPrefix  = "issues.watch_ledger."
	maxIssueWatches  = 50
	maxLedgerEntries = 5000
)

// ErrLedgerFull means the watch's ledger holds 5000 entries; the watch stops
// creating tasks (BR3.13).
var ErrLedgerFull = errors.New("the watch ledger is full")

// errDuplicateEntry is a second ledger entry for the same issue of a watch.
var errDuplicateEntry = errors.New("the issue is already in the watch ledger")

// Watches returns the workspace's issue watches.
func (s *Store) Watches(ctx context.Context, ws string) ([]IssueWatch, error) {
	var d listDoc[IssueWatch]
	err := s.load(ctx, "workspace", ws, keyWatches, &d)
	return d.Items, err
}

// WatchIndex lists the workspaces that have issue watches.
func (s *Store) WatchIndex(ctx context.Context) ([]string, error) {
	var d listDoc[string]
	err := s.load(ctx, "instance", "", keyWatchIndex, &d)
	return d.Items, err
}

// UpdateWatches replaces the watches with fn's result, at most 50.
func (s *Store) UpdateWatches(ctx context.Context, ws string, fn func([]IssueWatch) ([]IssueWatch, error)) error {
	defer s.lock("watches/" + ws)()
	list, err := s.Watches(ctx, ws)
	if err != nil {
		return err
	}
	next, err := fn(list)
	if errors.Is(err, errUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(next) > maxIssueWatches {
		return &connection.FieldError{Field: FieldLimit, Err: errLimit}
	}
	if err := s.put(ctx, "workspace", ws, keyWatches, listDoc[IssueWatch]{Items: next}); err != nil {
		return err
	}
	if len(next) == 0 {
		return nil
	}
	defer s.lock("instance/watches")()
	idx, err := s.WatchIndex(ctx)
	if err != nil || slices.Contains(idx, ws) {
		return err
	}
	return s.put(ctx, "instance", "", keyWatchIndex, listDoc[string]{Items: append(idx, ws)})
}

// Ledger returns one watch's ledger.
func (s *Store) Ledger(ctx context.Context, ws, watchID string) ([]IssueWatchLedgerEntry, error) {
	var d listDoc[IssueWatchLedgerEntry]
	err := s.load(ctx, "workspace", ws, keyLedgerPrefix+watchID, &d)
	return d.Items, err
}

// UpdateLedger replaces one watch's ledger with fn's result. Each issue key
// appears once (FR3.4) and the ledger holds at most 5000 entries
// (ErrLedgerFull, BR3.13); entries are never pruned while the watch exists.
func (s *Store) UpdateLedger(ctx context.Context, ws, watchID string, fn func([]IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error)) error {
	defer s.lock("ledger/" + ws + "/" + watchID)()
	list, err := s.Ledger(ctx, ws, watchID)
	if err != nil {
		return err
	}
	next, err := fn(list)
	if errors.Is(err, errUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(next) > maxLedgerEntries {
		return ErrLedgerFull
	}
	seen := make(map[string]bool, len(next))
	for i := range next {
		if seen[next[i].Key] {
			return errDuplicateEntry
		}
		seen[next[i].Key] = true
		next[i].WatchID = watchID
	}
	return s.put(ctx, "workspace", ws, keyLedgerPrefix+watchID, listDoc[IssueWatchLedgerEntry]{Items: next})
}

// DeleteLedger empties a deleted watch's ledger (BR3.10).
func (s *Store) DeleteLedger(ctx context.Context, ws, watchID string) error {
	return s.UpdateLedger(ctx, ws, watchID, func([]IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
		return []IssueWatchLedgerEntry{}, nil
	})
}
