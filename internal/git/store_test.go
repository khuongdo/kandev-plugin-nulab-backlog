package git

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

const ws = "ws-1"

func TestU4_GitState_DocsRoundTripWithSchemaVersion(t *testing.T) {
	state := newFakeState()
	s := NewStore(state)
	ctx := context.Background()
	link := Link{TaskID: "task-17", SpaceHost: host, ProjectKey: "PROJ", RepoName: "web-app", RepositoryID: 11, Number: 42, Status: StatusActive}
	require.NoError(t, s.UpdateLinks(ctx, ws, func(l []Link) ([]Link, error) { return append(l, link), nil }))
	require.NoError(t, s.UpdateWatches(ctx, ws, func(w []Watch) ([]Watch, error) { return append(w, Watch{ID: "w1", Name: "Reviews"}), nil }))
	require.NoError(t, s.UpdateLedger(ctx, ws, func(l []LedgerEntry) ([]LedgerEntry, error) {
		return append(l, LedgerEntry{Key: LinkKey(host, 11, 42), TaskID: "task-1"}), nil
	}))
	require.NoError(t, s.UpdateQueries(ctx, ws, func(q []Query) ([]Query, error) { return append(q, Query{ID: "q1", Name: "Open"}), nil }))

	links, err := s.Links(ctx, ws)
	require.NoError(t, err)
	require.Equal(t, []Link{link}, links)
	watches, err := s.Watches(ctx, ws)
	require.NoError(t, err)
	require.Equal(t, "Reviews", watches[0].Name)
	ledger, err := s.Ledger(ctx, ws)
	require.NoError(t, err)
	require.Equal(t, "task-1", ledger[0].TaskID)
	queries, err := s.Queries(ctx, ws)
	require.NoError(t, err)
	require.Equal(t, "Open", queries[0].Name)

	for _, key := range []string{"git.links", "git.watches", "git.ledger", "git.queries"} {
		raw, ok := state.get("workspace/" + ws + "/" + key)
		require.True(t, ok, key)
		require.EqualValues(t, 1, raw["schemaVersion"], key)
	}
	empty, err := s.Links(ctx, "ws-none")
	require.NoError(t, err)
	require.Empty(t, empty)
}

func TestU4_GitState_UnknownSchemaIsNeverOverwritten(t *testing.T) {
	state := newFakeState()
	future := map[string]any{"schemaVersion": float64(2), "items": []any{}}
	state.set("workspace/"+ws+"/git.links", future)
	s := NewStore(state)
	_, err := s.Links(context.Background(), ws)
	require.ErrorIs(t, err, connection.ErrStore)
	err = s.UpdateLinks(context.Background(), ws, func(l []Link) ([]Link, error) { return append(l, Link{TaskID: "t"}), nil })
	require.ErrorIs(t, err, connection.ErrStore)
	got, _ := state.get("workspace/" + ws + "/git.links")
	require.Equal(t, future, got)
	require.Zero(t, state.setCalls)
}

func TestU4_GitState_IndexListsWorkspacesWithItems(t *testing.T) {
	state := newFakeState()
	s := NewStore(state)
	ctx := context.Background()
	for range 2 {
		require.NoError(t, s.UpdateWatches(ctx, ws, func(w []Watch) ([]Watch, error) { return append(w, Watch{ID: "w"}), nil }))
	}
	require.NoError(t, s.UpdateLinks(ctx, "ws-2", func(l []Link) ([]Link, error) { return append(l, Link{TaskID: "t"}), nil }))
	require.NoError(t, s.UpdateQueries(ctx, "ws-3", func(q []Query) ([]Query, error) { return append(q, Query{ID: "q"}), nil }))
	idx, err := s.Index(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{ws, "ws-2"}, idx, "watches and links index a workspace once; queries do not")
	raw, ok := state.get("instance//git.watch_index")
	require.True(t, ok)
	require.EqualValues(t, 1, raw["schemaVersion"])
}

func TestU4_GitState_CapsAreValidationErrors(t *testing.T) {
	s := NewStore(newFakeState())
	ctx := context.Background()
	fill := func(n int) []Link { return make([]Link, n) }
	require.NoError(t, s.UpdateLinks(ctx, ws, func([]Link) ([]Link, error) { return fill(500), nil }))
	err := s.UpdateLinks(ctx, ws, func(l []Link) ([]Link, error) { return append(l, Link{}), nil })
	require.Equal(t, "limit", fieldOf(t, err))
	err = s.UpdateWatches(ctx, ws, func([]Watch) ([]Watch, error) { return make([]Watch, 51), nil })
	require.Equal(t, "limit", fieldOf(t, err))
	err = s.UpdateQueries(ctx, ws, func([]Query) ([]Query, error) { return make([]Query, 51), nil })
	require.Equal(t, "limit", fieldOf(t, err))
	require.NoError(t, s.UpdateQueries(ctx, ws, func([]Query) ([]Query, error) { return make([]Query, 50), nil }))
}

func TestU4_GitState_CallsStopAfterOneSecond(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		state := newFakeState()
		state.block = true
		s := NewStore(state)
		start := time.Now()
		_, err := s.Watches(context.Background(), ws)
		require.ErrorIs(t, err, connection.ErrStore)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, time.Second, time.Since(start))
	})
}

func TestU4_GitState_CancelIsReturnedUnchanged(t *testing.T) {
	state := newFakeState()
	state.block = true
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewStore(state).Links(ctx, ws)
	require.Equal(t, context.Canceled, err)
}

func TestU4_GitState_WorkspacesAndWritersAreIndependent(t *testing.T) {
	state := newFakeState()
	s := NewStore(state)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.UpdateLinks(ctx, ws, func(l []Link) ([]Link, error) {
				return append(l, Link{TaskID: fmt.Sprintf("task-%d", i)}), nil
			})
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	links, err := s.Links(ctx, ws)
	require.NoError(t, err)
	require.Len(t, links, 20, "parallel read-modify-writes in one workspace lose nothing")
	other, err := s.Links(ctx, "ws-2")
	require.NoError(t, err)
	require.Empty(t, other)

	state.failSet = true
	err = s.UpdateLinks(ctx, ws, func(l []Link) ([]Link, error) { return l, nil })
	require.ErrorIs(t, err, connection.ErrStore)
	errStop := errors.New("stop")
	require.ErrorIs(t, s.UpdateLinks(ctx, ws, func([]Link) ([]Link, error) { return nil, errStop }), errStop)
}
