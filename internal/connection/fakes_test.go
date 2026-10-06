package connection

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"
)

var errInjected = errors.New("injected store failure")

// fakeSecrets is an in-memory SecretStore that can inject failures.
type fakeSecrets struct {
	mu        sync.Mutex
	data      map[string]string
	failGet   bool
	failSet   int // fail the Nth SetSecret call (1-based); 0 = never
	failAll   bool
	failDel   bool
	blockSet  bool // block SetSecret until the context ends
	setCalls  int
	getCalls  int
	ctxErrSet []error // ctx.Err() seen by each SetSecret/DeleteSecret call
	// U4: the order of successful writes ("set:<key>", "del:<key>") and a
	// key whose DeleteSecret fails.
	ops        []string
	failDelKey string
}

func newFakeSecrets() *fakeSecrets { return &fakeSecrets{data: map[string]string{}} }

func (f *fakeSecrets) GetSecret(ctx context.Context, key string) (string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getCalls++
	if f.failGet {
		return "", false, errInjected
	}
	v, ok := f.data[key]
	return v, ok, nil
}

func (f *fakeSecrets) SetSecret(ctx context.Context, key, value string) error {
	f.mu.Lock()
	f.setCalls++
	n := f.setCalls
	f.ctxErrSet = append(f.ctxErrSet, ctx.Err())
	block := f.blockSet
	f.mu.Unlock()
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failAll || n == f.failSet {
		return errInjected
	}
	f.data[key] = value
	f.ops = append(f.ops, "set:"+key)
	return nil
}

func (f *fakeSecrets) DeleteSecret(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ctxErrSet = append(f.ctxErrSet, ctx.Err())
	if f.failDel || f.failAll || (f.failDelKey != "" && key == f.failDelKey) {
		return errInjected
	}
	delete(f.data, key)
	f.ops = append(f.ops, "del:"+key)
	return nil
}

func (f *fakeSecrets) snapshot() map[string]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := map[string]string{}
	for k, v := range f.data {
		out[k] = v
	}
	return out
}

// fakeState is an in-memory StateStore that can inject failures.
type fakeState struct {
	mu        sync.Mutex
	data      map[string]map[string]any
	failGet   bool
	failSet   bool
	failDel   bool
	cancelSet context.CancelFunc // cancels the action context during SetState
	getDelay  time.Duration      // each GetState takes this long (virtual time under synctest)
	blockGet  bool               // block GetState until the context ends
	setCalls  int
}

func newFakeState() *fakeState { return &fakeState{data: map[string]map[string]any{}} }

func stateKey(scope, scopeID, key string) string { return scope + "/" + scopeID + "/" + key }

func (f *fakeState) GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error) {
	f.mu.Lock()
	delay, block := f.getDelay, f.blockGet
	f.mu.Unlock()
	if block {
		<-ctx.Done()
		return nil, false, ctx.Err()
	}
	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, false, ctx.Err()
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failGet {
		return nil, false, errInjected
	}
	v, ok := f.data[stateKey(scope, scopeID, key)]
	return v, ok, nil
}

func (f *fakeState) SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.setCalls++
	if f.cancelSet != nil {
		f.cancelSet()
		return context.Canceled
	}
	if f.failSet {
		return errInjected
	}
	f.data[stateKey(scope, scopeID, key)] = value
	return nil
}

func (f *fakeState) DeleteState(ctx context.Context, scope, scopeID, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failDel {
		return errInjected
	}
	delete(f.data, stateKey(scope, scopeID, key))
	return nil
}

func (f *fakeState) snapshot() map[string]map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := map[string]map[string]any{}
	for k, v := range f.data {
		out[k] = v
	}
	return out
}

// syncBuffer is a goroutine-safe log sink.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func testLogger(t *testing.T) (*slog.Logger, *syncBuffer) {
	t.Helper()
	buf := &syncBuffer{}
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})), buf
}

var fixedNow = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func newTestStore(secrets *fakeSecrets, state *fakeState) *Store {
	s := NewStore(secrets, state)
	s.Now = func() time.Time { return fixedNow }
	return s
}
