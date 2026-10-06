package redact

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// Marker replaces every masked secret.
const Marker = "[REDACTED]"

// minSecretLen keeps one-character values from masking ordinary text.
const minSecretLen = 4

// urlQuery matches the query string of any http(s) URL in free text.
var urlQuery = regexp.MustCompile(`(https?://[^\s?#"'<>]*)\?[^\s#"'<>]*`)

type secretsKey struct{}
type loggerKey struct{}

// WithSecrets returns a context whose redaction set also holds secrets.
func WithSecrets(ctx context.Context, secrets ...string) context.Context {
	prev, _ := ctx.Value(secretsKey{}).([]string)
	next := append([]string(nil), prev...)
	for _, s := range secrets {
		if len(s) >= minSecretLen {
			next = append(next, s)
		}
	}
	return context.WithValue(ctx, secretsKey{}, next)
}

// String masks the context's secrets and every URL query string in s.
func String(ctx context.Context, s string) string {
	if ctx != nil {
		secrets, _ := ctx.Value(secretsKey{}).([]string)
		for _, secret := range secrets {
			s = strings.ReplaceAll(s, secret, Marker)
		}
	}
	return urlQuery.ReplaceAllString(s, "$1?REDACTED")
}

// WithLogger attaches the action's logger (already carrying workspaceId and
// requestId) to the context.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// Logger returns the context's logger, or a logger that discards everything.
func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.New(slog.DiscardHandler)
}

// handler masks the message and every attribute before the inner handler
// sees them (NFR3.3).
type handler struct {
	inner slog.Handler
}

// NewHandler wraps inner so nothing unredacted reaches it.
func NewHandler(inner slog.Handler) slog.Handler {
	return handler{inner: inner}
}

func (h handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h handler) Handle(ctx context.Context, r slog.Record) error {
	out := slog.NewRecord(r.Time, r.Level, String(ctx, r.Message), r.PC)
	r.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(maskAttr(ctx, a))
		return true
	})
	return h.inner.Handle(ctx, out)
}

func (h handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	masked := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		masked[i] = maskAttr(context.Background(), a)
	}
	return handler{inner: h.inner.WithAttrs(masked)}
}

func (h handler) WithGroup(name string) slog.Handler {
	return handler{inner: h.inner.WithGroup(name)}
}

func maskAttr(ctx context.Context, a slog.Attr) slog.Attr {
	v := a.Value.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return slog.String(a.Key, String(ctx, v.String()))
	case slog.KindGroup:
		group := v.Group()
		masked := make([]any, len(group))
		for i, g := range group {
			masked[i] = maskAttr(ctx, g)
		}
		return slog.Group(a.Key, masked...)
	case slog.KindAny:
		if err, ok := v.Any().(error); ok {
			return slog.String(a.Key, String(ctx, err.Error()))
		}
		return slog.String(a.Key, String(ctx, fmt.Sprintf("%+v", v.Any())))
	default:
		return slog.Attr{Key: a.Key, Value: v}
	}
}
