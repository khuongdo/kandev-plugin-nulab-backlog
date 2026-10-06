package issues

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

func TestU3_Store_RoundTripsTheDocuments(t *testing.T) {
	state := newMemState()
	s := NewStore(state)
	ctx := context.Background()
	l := Link{IssueKey: "PROJ-120", IssueID: 5120, ProjectKey: "PROJ", SpaceHost: "example-space.backlog.com",
		TaskID: "task-17", TaskKey: "T-17", State: StateActive, LastKnownStatus: "Open", ConnectionEpoch: 2, CreatedAt: "2026-10-06T00:00:00Z"}
	require.NoError(t, s.UpdateLinks(ctx, "ws-1", func(links []Link) ([]Link, error) { return append(links, l), nil }))
	got, err := s.Links(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, []Link{l}, got)
	require.EqualValues(t, 1, state.raw("workspace", "ws-1", "issues.links")["schemaVersion"])

	require.NoError(t, s.UpdateSettings(ctx, "ws-1", func(st Settings) (Settings, error) {
		require.Equal(t, Settings{PollMinutes: DefaultPollMinutes}, st, "defaults before the first write")
		st.PollMinutes, st.LastCycleAt = 2, "2026-10-06T00:05:00Z"
		return st, nil
	}))
	st, err := s.Settings(ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, Settings{PollMinutes: 2, LastCycleAt: "2026-10-06T00:05:00Z"}, st)
	raw := state.raw("workspace", "ws-1", "issues.settings")
	require.EqualValues(t, 1, raw["schemaVersion"])
	require.EqualValues(t, 2, raw["pollMinutes"])

	empty, err := s.Links(ctx, "ws-2")
	require.NoError(t, err)
	require.Empty(t, empty, "workspaces are independent")
}

func TestU3_Store_NeverOverwritesUnknownSchema(t *testing.T) {
	for _, key := range []string{"issues.links", "issues.settings"} {
		t.Run(key, func(t *testing.T) {
			state := newMemState()
			future := map[string]any{"schemaVersion": 2, "items": []any{}, "pollMinutes": 9}
			state.put("workspace", "ws-1", key, future)
			s := NewStore(state)
			ctx := context.Background()
			_, err1 := s.Links(ctx, "ws-1")
			_, err2 := s.Settings(ctx, "ws-1")
			err3 := s.UpdateLinks(ctx, "ws-1", func(l []Link) ([]Link, error) { return l, nil })
			err4 := s.UpdateSettings(ctx, "ws-1", func(st Settings) (Settings, error) { return st, nil })
			errs := map[string][]error{"issues.links": {err1, err3}, "issues.settings": {err2, err4}}[key]
			for _, err := range errs {
				require.ErrorIs(t, err, connection.ErrStore)
			}
			require.Equal(t, roundTrip(future), state.raw("workspace", "ws-1", key), "left as it was")
			require.Zero(t, state.setsByKey[key])
		})
	}
}

func TestU3_Store_IndexListsLinkedWorkspaces(t *testing.T) {
	s := NewStore(newMemState())
	ctx := context.Background()
	add := func(ws string) {
		require.NoError(t, s.UpdateLinks(ctx, ws, func(l []Link) ([]Link, error) {
			return append(l, Link{IssueKey: "PROJ-1", TaskID: "t-" + ws, State: StateActive}), nil
		}))
	}
	add("ws-1")
	add("ws-2")
	add("ws-1")
	idx, err := s.Index(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"ws-1", "ws-2"}, idx)
}

func TestU3_Store_CapsTheLinks(t *testing.T) {
	s := NewStore(newMemState())
	ctx := context.Background()
	err := s.UpdateLinks(ctx, "ws-1", func([]Link) ([]Link, error) { return make([]Link, maxLinks), nil })
	require.NoError(t, err)
	err = s.UpdateLinks(ctx, "ws-1", func(l []Link) ([]Link, error) { return append(l, Link{}), nil })
	var fe *connection.FieldError
	require.True(t, errors.As(err, &fe))
	require.Equal(t, FieldLimit, fe.Field)
	got, _ := s.Links(ctx, "ws-1")
	require.Len(t, got, maxLinks, "nothing written")
}

func TestU3_Store_CallsKeepTheOneSecondLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		state := newMemState()
		state.delay = 5 * time.Second
		s := NewStore(state)
		start := time.Now()
		_, err := s.Links(context.Background(), "ws-1")
		require.ErrorIs(t, err, connection.ErrStore)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, time.Second, time.Since(start))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = s.Settings(ctx, "ws-1")
		require.Equal(t, context.Canceled, err, "the caller's cancel is returned unchanged")
	})
}

func TestU3_Store_ParallelWritesAreSerialised(t *testing.T) {
	s := NewStore(newMemState())
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			require.NoError(t, s.UpdateLinks(ctx, "ws-1", func(l []Link) ([]Link, error) {
				return append(l, Link{IssueKey: "PROJ-1", TaskID: fmt.Sprintf("task-%d", i)}), nil
			}))
		})
	}
	wg.Wait()
	got, err := s.Links(ctx, "ws-1")
	require.NoError(t, err)
	require.Len(t, got, 20, "no write was lost")
}

func TestU3_Store_UnchangedWritesNothing(t *testing.T) {
	state := newMemState()
	s := NewStore(state)
	require.NoError(t, s.UpdateLinks(context.Background(), "ws-1", func([]Link) ([]Link, error) { return nil, errUnchanged }))
	require.Zero(t, state.sets)
	state.failGet = true
	require.ErrorIs(t, s.UpdateLinks(context.Background(), "ws-1", func(l []Link) ([]Link, error) { return l, nil }), connection.ErrStore)
}
