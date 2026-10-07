package connection

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

// fakeConfig is the operator config Host.GetConfig returns.
type fakeConfig struct {
	mu    sync.Mutex
	m     map[string]any
	err   error
	calls int
}

func (f *fakeConfig) GetConfig(context.Context) (map[string]any, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.m, f.err
}

// u2 is a harness with OAuth configured and a settable clock.
type u2 struct {
	*harness
	cfg          *fakeConfig
	clientSecret string
	verifier     string // the browser cookie of the last start
	mu           sync.Mutex
	now          time.Time
}

func newU2(t *testing.T) *u2 {
	t.Helper()
	h := newHarness(t)
	m, secret := oauthConfigMap(t)
	u := &u2{harness: h, cfg: &fakeConfig{m: m}, clientSecret: secret, now: fixedNow}
	h.svc.Config = u.cfg
	h.svc.store.Now = u.clock
	return u
}

func (u *u2) clock() time.Time {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.now
}

func (u *u2) advance(d time.Duration) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.now = u.now.Add(d)
}

// connectKey connects ws with a fresh API key and returns it.
func (u *u2) connectKey(t *testing.T, host string) string {
	t.Helper()
	key := testutil.APIKey(t)
	_, err := u.svc.Connect(u.ctx, ws, ConnectInput{SpaceURL: host, APIKey: key})
	require.NoError(t, err)
	return key
}

// connectOAuth stores an OAuth connection that expires after ttl.
func (u *u2) connectOAuth(t *testing.T, ttl time.Duration) backlog.TokenSet {
	t.Helper()
	tokens := backlog.TokenSet{AccessToken: testutil.Token(t), RefreshToken: testutil.Token(t), ExpiresAt: u.clock().Add(ttl)}
	_, err := u.svc.store.SaveOAuth(u.ctx, ws, "a.backlog.com", tokens, testUser)
	require.NoError(t, err)
	return tokens
}

// collector receives ConnectionChanged events.
type collector struct {
	ch    chan ConnectionChanged
	unsub func()
}

func subscribe(t *testing.T, svc *Service) *collector {
	t.Helper()
	c := &collector{ch: make(chan ConnectionChanged, 64)}
	c.unsub = svc.Subscribe(func(e ConnectionChanged) { c.ch <- e })
	t.Cleanup(c.unsub)
	return c
}

// next waits for the next event; the timeout only guards a broken test.
func (c *collector) next(t *testing.T) ConnectionChanged {
	t.Helper()
	select {
	case e := <-c.ch:
		return e
	case <-time.After(5 * time.Second):
		t.Fatal("no ConnectionChanged event")
		return ConnectionChanged{}
	}
}

// changedLogs counts connection_changed lines, which are written when an
// event is sent; 0 lines means no event.
func (u *u2) changedLogs(t *testing.T) []map[string]any {
	t.Helper()
	return eventsNamed(u.events(t), "connection_changed")
}
