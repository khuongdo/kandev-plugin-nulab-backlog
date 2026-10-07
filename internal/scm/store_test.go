package scm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// FR4.2: every document is {schemaVersion: 1, items} under its own key.
func TestStore_DocumentsRoundTripWithSchemaVersion1(t *testing.T) {
	ctx := context.Background()
	st := newFakeState()
	s := NewStore(st)

	require.NoError(t, s.UpdateSettings(ctx, "ws-1", func([]Settings) ([]Settings, error) {
		return []Settings{{Provider: GitHub, HasToken: true, Account: "Lan",
			Mappings: []Mapping{{ProjectKey: "PROJ", Repos: []string{"acme/web"}}}}}, nil
	}))
	require.NoError(t, s.UpdateLinks(ctx, "ws-1", func([]Link) ([]Link, error) {
		return []Link{{PRRef: PRRef{GitHub, "acme/web", 1}, TaskID: "t-1"}}, nil
	}))
	require.NoError(t, s.UpdateDismissed(ctx, "ws-1", func([]string) ([]string, error) { return []string{"k"}, nil }))
	require.NoError(t, s.UpdateQueries(ctx, "ws-1", func([]Query) ([]Query, error) { return []Query{{QueryInput: QueryInput{ID: "q"}}}, nil }))
	require.NoError(t, s.UpdateWatches(ctx, "ws-1", func([]Watch) ([]Watch, error) {
		return []Watch{{WatchInput: WatchInput{QueryInput: QueryInput{ID: "w"}}}}, nil
	}))
	require.NoError(t, s.UpdateLedger(ctx, "ws-1", func([]LedgerEntry) ([]LedgerEntry, error) {
		return []LedgerEntry{{Key: "k", WatchID: "w"}}, nil
	}))

	for _, key := range []string{"scm.settings", "scm.links", "scm.dismissed", "scm.queries", "scm.watches", "scm.ledger"} {
		v, ok := st.get("workspace/ws-1/" + key)
		require.True(t, ok, key)
		require.EqualValues(t, 1, v["schemaVersion"], key)
		require.Len(t, v["items"], 1, key)
	}
	settings, err := s.Settings(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, "acme/web", settings[0].Mappings[0].Repos[0])
	links, err := s.Links(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, "github|acme/web|1", links[0].Key())
	dismissed, err := s.Dismissed(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []string{"k"}, dismissed)
	idx, err := s.Index(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"ws-1"}, idx, "a workspace with links or watches is indexed for the watcher")
}

func TestStore_MissingDocumentsAreEmpty(t *testing.T) {
	s := NewStore(newFakeState())
	q, err := s.Queries(context.Background(), "ws-1")
	require.NoError(t, err)
	require.Empty(t, q)
}

// FR4.2: links 500, queries 50, watches 50 per workspace.
func TestStore_Caps(t *testing.T) {
	ctx := context.Background()
	s := NewStore(newFakeState())
	fill := func(n int) []Query { return make([]Query, n) }
	require.NoError(t, s.UpdateQueries(ctx, "ws", func([]Query) ([]Query, error) { return fill(50), nil }))
	err := s.UpdateQueries(ctx, "ws", func([]Query) ([]Query, error) { return fill(51), nil })
	require.Equal(t, FieldLimit, fieldOf(t, err))
	err = s.UpdateWatches(ctx, "ws", func([]Watch) ([]Watch, error) { return make([]Watch, 51), nil })
	require.Equal(t, FieldLimit, fieldOf(t, err))
	err = s.UpdateLinks(ctx, "ws", func([]Link) ([]Link, error) { return make([]Link, 501), nil })
	require.Equal(t, FieldLimit, fieldOf(t, err))
	require.NoError(t, s.UpdateLinks(ctx, "ws", func([]Link) ([]Link, error) { return make([]Link, 500), nil }))
}

func TestStore_UnknownSchemaIsAStoreErrorAndNeverOverwritten(t *testing.T) {
	ctx := context.Background()
	st := newFakeState()
	st.set("workspace/ws/scm.links", map[string]any{"schemaVersion": 2, "items": []any{}})
	s := NewStore(st)
	_, err := s.Links(ctx, "ws")
	require.ErrorIs(t, err, connection.ErrStore)
	err = s.UpdateLinks(ctx, "ws", func(l []Link) ([]Link, error) { return l, nil })
	require.ErrorIs(t, err, connection.ErrStore)
	v, _ := st.get("workspace/ws/scm.links")
	require.EqualValues(t, 2, v["schemaVersion"])
}

func TestStore_UnchangedWritesNothingAndFnErrorsPassThrough(t *testing.T) {
	ctx := context.Background()
	st := newFakeState()
	s := NewStore(st)
	require.NoError(t, s.UpdateLinks(ctx, "ws", func([]Link) ([]Link, error) { return nil, errUnchanged }))
	require.Zero(t, st.setCalls)
	boom := errors.New("boom")
	require.ErrorIs(t, s.UpdateLinks(ctx, "ws", func([]Link) ([]Link, error) { return nil, boom }), boom)
}

func TestStore_ReadAndWriteFailuresAreStoreErrors(t *testing.T) {
	ctx := context.Background()
	st := newFakeState()
	s := NewStore(st)
	st.failSet = true
	err := s.UpdateWatches(ctx, "ws", func([]Watch) ([]Watch, error) {
		return []Watch{{WatchInput: WatchInput{QueryInput: QueryInput{ID: "w"}}}}, nil
	})
	require.ErrorIs(t, err, connection.ErrStore)
	st.failSet, st.failGet = false, true
	_, err = s.Watches(ctx, "ws")
	require.ErrorIs(t, err, connection.ErrStore)

	cctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = s.Watches(cctx, "ws")
	require.ErrorIs(t, err, context.Canceled, "the caller's cancellation is returned unchanged")
}

func TestStore_ConcurrentUpdatesDoNotLoseWrites(t *testing.T) {
	ctx := context.Background()
	s := NewStore(newFakeState())
	done := make(chan error)
	for i := range 20 {
		go func() {
			done <- s.UpdateQueries(ctx, "ws", func(q []Query) ([]Query, error) {
				return append(q, Query{QueryInput: QueryInput{ID: fmt.Sprint(i)}}), nil
			})
		}()
	}
	for range 20 {
		require.NoError(t, <-done)
	}
	q, err := s.Queries(ctx, "ws")
	require.NoError(t, err)
	require.Len(t, q, 20)
}
