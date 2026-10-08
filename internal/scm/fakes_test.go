package scm

import (
	"context"
	"errors"
	"sync"
)

var errInjected = errors.New("injected failure")

// fakeState is an in-memory StateStore, like internal/git's.
type fakeState struct {
	mu       sync.Mutex
	data     map[string]map[string]any
	failGet  bool
	failSet  bool
	setCalls int
}

func newFakeState() *fakeState { return &fakeState{data: map[string]map[string]any{}} }

func (f *fakeState) GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if f.failGet {
		return nil, false, errInjected
	}
	v, ok := f.data[scope+"/"+scopeID+"/"+key]
	return v, ok, nil
}

func (f *fakeState) SetState(_ context.Context, scope, scopeID, key string, value map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.setCalls++
	if f.failSet {
		return errInjected
	}
	f.data[scope+"/"+scopeID+"/"+key] = value
	return nil
}

func (f *fakeState) get(key string) (map[string]any, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[key]
	return v, ok
}

func (f *fakeState) set(key string, v map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = v
}
