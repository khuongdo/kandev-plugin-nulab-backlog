package issues

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU3_Settings_DefaultIsFiveMinutes(t *testing.T) {
	r := newRig(t)
	got, err := r.svc.Settings(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, Settings{PollMinutes: 5}, got)
}

func TestU3_Settings_SetPollInterval(t *testing.T) {
	r := newRig(t)
	got, err := r.svc.SetPollInterval(r.ctx, "ws-1", 15.0)
	require.NoError(t, err)
	require.Equal(t, 15, got.PollMinutes)
	again, err := r.svc.Settings(r.ctx, "ws-1")
	require.NoError(t, err)
	require.Equal(t, got, again)
}

func TestU3_Settings_RefusesBadMinutes(t *testing.T) {
	r := newRig(t)
	for _, raw := range []any{0.5, 0.0, "x", nil} {
		_, err := r.svc.SetPollInterval(r.ctx, "ws-1", raw)
		require.Equal(t, FieldMinutes, fieldOf(t, err), "AC4.2.2: %v", raw)
	}
	require.Zero(t, r.state.sets, "nothing stored")
}
