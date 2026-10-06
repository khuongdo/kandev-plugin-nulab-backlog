package redact

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestStringMasksRegisteredSecrets(t *testing.T) {
	key := testutil.APIKey(t)
	ctx := WithSecrets(context.Background(), key)
	got := String(ctx, "the key is "+key+" ok")
	require.Equal(t, "the key is "+Marker+" ok", got)
}

func TestStringIgnoresEmptyAndShortSecrets(t *testing.T) {
	ctx := WithSecrets(context.Background(), "", "a")
	require.Equal(t, "a b c", String(ctx, "a b c"))
}

func TestStringReplacesURLQueries(t *testing.T) {
	in := `Get "https://example-space.backlog.com/api/v2/users/myself?apiKey=abc123&x=1": dial tcp`
	got := String(context.Background(), in)
	require.Equal(t, `Get "https://example-space.backlog.com/api/v2/users/myself?REDACTED": dial tcp`, got)
}

func TestSecretsAccumulateAcrossContexts(t *testing.T) {
	a, b := testutil.APIKey(t), testutil.APIKey(t)
	ctx := WithSecrets(WithSecrets(context.Background(), a), b)
	got := String(ctx, a+" "+b)
	require.Equal(t, Marker+" "+Marker, got)
}

func TestHandlerMasksAttributesMessagesAndErrorChains(t *testing.T) {
	key := testutil.APIKey(t)
	var buf bytes.Buffer
	log := slog.New(NewHandler(slog.NewJSONHandler(&buf, nil)))
	ctx := WithSecrets(context.Background(), key)

	inner := errors.New("upstream said " + key)
	wrapped := fmt.Errorf("connect: %w", fmt.Errorf("myself: %w", inner))
	log.With("url", "https://a.backlog.com/x?apiKey="+key).
		WithGroup("g").
		InfoContext(ctx, "msg "+key, "err", wrapped, "plain", key, "n", 42, slog.Group("nested", "k", key))

	out := buf.String()
	testutil.AssertNoLeak(t, out, key, 8)
	require.Contains(t, out, "?REDACTED")
	require.Contains(t, out, `"n":42`)
	require.Contains(t, out, Marker)
}

func TestHandlerHonoursLevel(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(NewHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	log.Debug("hidden")
	require.Empty(t, buf.String())
	require.False(t, log.Enabled(context.Background(), slog.LevelDebug))
}

func TestLoggerTravelsOnTheContext(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, nil)).With("requestId", "r1")
	ctx := WithLogger(context.Background(), l)
	Logger(ctx).InfoContext(ctx, "hello")
	require.True(t, strings.Contains(buf.String(), `"requestId":"r1"`))
	require.NotNil(t, Logger(context.Background()), "a context without a logger yields a discard logger")
}
