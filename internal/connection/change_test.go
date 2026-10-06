package connection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionChangedReasonAndRestore(t *testing.T) {
	a := record{SchemaVersion: 1, SpaceHost: "a.backlog.com", ConnectionEpoch: 1, SelectedProjects: []string{"PROJ"}}
	cases := []struct {
		name    string
		prev    record
		next    record
		reason  Reason
		restore bool
	}{
		{"first connect", record{}, a, ReasonConnected, false},
		{"reconnect to the same host after a disconnect", record{Disconnected: true, PreviousSpaceHost: "a.backlog.com", ConnectionEpoch: 2}, a, ReasonConnected, true},
		{"connect to another host after a disconnect", record{Disconnected: true, PreviousSpaceHost: "b.backlog.com", ConnectionEpoch: 2}, a, ReasonConnected, false},
		{"same host while connected", a, a, ReasonCredentialsReplaced, false},
		{"a different host", a, record{SpaceHost: "b.backlog.com"}, ReasonSpaceChanged, false},
		{"back to the remembered host", record{SpaceHost: "b.backlog.com", PreviousSpaceHost: "a.backlog.com"}, a, ReasonSpaceChanged, true},
		{"disconnect", a, record{Disconnected: true, PreviousSpaceHost: "a.backlog.com"}, ReasonDisconnected, false},
		{"a project removed", a, record{SpaceHost: "a.backlog.com"}, ReasonProjectsChanged, false},
		{"a project added", a, record{SpaceHost: "a.backlog.com", SelectedProjects: []string{"PROJ", "DEMO"}}, ReasonProjectsChanged, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reason, restore := changeFor(tc.prev, tc.next)
			require.Equal(t, tc.reason, reason)
			require.Equal(t, tc.restore, restore)
		})
	}
}
