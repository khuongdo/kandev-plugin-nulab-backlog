package plugin

import (
	"context"
	"testing"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"
)

func TestOAuthConfigAdapterPassesGetConfigThrough(t *testing.T) {
	h := newFakeHost()
	h.config = map[string]any{"oauth_client_id": "client-id-1"}
	got, err := hostStores{host: func() pluginsdk.Host { return h }}.GetConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, h.config, got)
}

func TestOAuthConfigAdapterWithoutAHostIsAnError(t *testing.T) {
	_, err := hostStores{host: func() pluginsdk.Host { return nil }}.GetConfig(context.Background())
	require.ErrorIs(t, err, errNoHost)
}
