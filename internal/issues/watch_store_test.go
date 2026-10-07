package issues

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

func TestWatchStore_RoundTripsWatchesAndIndexesTheWorkspace(t *testing.T) {
	state := newMemState()
	s := NewStore(state)
	ctx := context.Background()
	w := IssueWatch{ID: "w1", Name: "Open bugs", SpaceHost: spaceHost, ProjectKey: "PROJ", StatusIDs: []int64{1},
		Assignee: WhoAnyone, Creator: WhoMe, CreatedUserID: 7, WorkflowID: "wf-1", IntervalMinutes: 5, State: StateActive,
		Cursor: &IssueCursor{Created: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC), IssueID: 5118, DayOffset: 3}}
	require.NoError(t, s.UpdateWatches(ctx, "ws-1", func(l []IssueWatch) ([]IssueWatch, error) { return append(l, w), nil }))
	got, err := s.Watches(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []IssueWatch{w}, got)
	require.EqualValues(t, 1, state.raw("workspace", "ws-1", "issues.watches")["schemaVersion"])
	idx, err := s.WatchIndex(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"ws-1"}, idx)
	links, err := s.Index(ctx)
	require.NoError(t, err)
	require.Empty(t, links, "the link sync index is separate: a watch alone needs no status sync")
}

func TestWatchStore_CapsFiftyWatches(t *testing.T) {
	s := NewStore(newMemState())
	ctx := context.Background()
	require.NoError(t, s.UpdateWatches(ctx, "ws-1", func([]IssueWatch) ([]IssueWatch, error) {
		return make([]IssueWatch, maxIssueWatches), nil
	}))
	err := s.UpdateWatches(ctx, "ws-1", func(l []IssueWatch) ([]IssueWatch, error) { return append(l, IssueWatch{}), nil })
	require.Equal(t, FieldLimit, fieldOf(t, err))
	require.Equal(t, 50, maxIssueWatches)
}

// BR3.10, BR3.13: one ledger per watch; an issue key is unique in it; 5000
// entries at most; deleting the watch's ledger empties it.
func TestWatchStore_LedgerPerWatch(t *testing.T) {
	s := NewStore(newMemState())
	ctx := context.Background()
	entry := func(key string) IssueWatchLedgerEntry {
		return IssueWatchLedgerEntry{Key: key, Outcome: OutcomeReserved, At: "2026-10-07T00:00:00Z"}
	}
	add := func(watch, key string) error {
		return s.UpdateLedger(ctx, "ws-1", watch, func(l []IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
			return append(l, entry(key)), nil
		})
	}
	require.NoError(t, add("w1", spaceHost+"|5118"))
	require.NoError(t, add("w2", spaceHost+"|5118"), "another watch has its own ledger")
	require.ErrorIs(t, add("w1", spaceHost+"|5118"), errDuplicateEntry, "unique (watchId, key)")
	got, err := s.Ledger(ctx, "ws-1", "w1")
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "w1", got[0].WatchID, "the store stamps the watch id")

	require.NoError(t, s.DeleteLedger(ctx, "ws-1", "w1"))
	got, err = s.Ledger(ctx, "ws-1", "w1")
	require.NoError(t, err)
	require.Empty(t, got)
	other, err := s.Ledger(ctx, "ws-1", "w2")
	require.NoError(t, err)
	require.Len(t, other, 1, "other watches keep theirs")

	full := make([]IssueWatchLedgerEntry, 0, maxLedgerEntries)
	for i := range maxLedgerEntries {
		full = append(full, entry(fmt.Sprintf("%s|%d", spaceHost, i)))
	}
	require.NoError(t, s.UpdateLedger(ctx, "ws-1", "w3", func([]IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) {
		return full, nil
	}))
	require.ErrorIs(t, add("w3", spaceHost+"|x"), ErrLedgerFull)
	require.Equal(t, 5000, maxLedgerEntries)
}

func TestWatchStore_NeverOverwritesUnknownSchema(t *testing.T) {
	for _, key := range []string{"issues.watches", "issues.watch_ledger.w1"} {
		t.Run(key, func(t *testing.T) {
			state := newMemState()
			future := map[string]any{"schemaVersion": 2, "items": []any{}}
			state.put("workspace", "ws-1", key, future)
			s := NewStore(state)
			ctx := context.Background()
			var err error
			if key == "issues.watches" {
				_, err = s.Watches(ctx, "ws-1")
				require.ErrorIs(t, err, connection.ErrStore)
				err = s.UpdateWatches(ctx, "ws-1", func(l []IssueWatch) ([]IssueWatch, error) { return l, nil })
			} else {
				_, err = s.Ledger(ctx, "ws-1", "w1")
				require.ErrorIs(t, err, connection.ErrStore)
				err = s.UpdateLedger(ctx, "ws-1", "w1", func(l []IssueWatchLedgerEntry) ([]IssueWatchLedgerEntry, error) { return l, nil })
			}
			require.True(t, errors.Is(err, connection.ErrStore))
			require.Equal(t, roundTrip(future), state.raw("workspace", "ws-1", key), "left as it was")
		})
	}
}
