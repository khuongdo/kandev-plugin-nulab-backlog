package connection

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestU2_ReadsAU1RecordAndSecret(t *testing.T) {
	secrets, state := newFakeSecrets(), newFakeState()
	state.data[stateKey("workspace", ws, "connection")] = map[string]any{
		"schemaVersion": 1, "spaceHost": "a.backlog.com", "authMethod": "api_key",
		"connectedUserName": "Test User", "connectedUserId": 1234, "connectionEpoch": 3,
		"connectedAt": fixedNow.Format(time.RFC3339),
	}
	raw, err := json.Marshal(map[string]any{"apiKey": testutil.APIKey(t), "spaceHost": "a.backlog.com", "connectionEpoch": 3})
	require.NoError(t, err)
	secrets.data["backlog.connection."+ws] = string(raw)

	view, err := newTestStore(secrets, state).Load(context.Background(), ws)
	require.NoError(t, err)
	require.Equal(t, StateConnected, view.State)
	require.Equal(t, "api_key", view.AuthMethod)
	require.True(t, view.HasAPIKey)
	require.False(t, view.HasOAuthToken)
	require.Empty(t, view.SelectedProjects)
}

func TestU2_SecretWithoutAuthMethodIsAnAPIKeySecret(t *testing.T) {
	var sec secret
	require.NoError(t, json.Unmarshal([]byte(`{"apiKey":"k-1234","spaceHost":"a.backlog.com","connectionEpoch":1}`), &sec))
	require.Equal(t, authAPIKey, sec.method())
}
