package issues

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// memState is an in-memory Kandev state store. Values round-trip through
// JSON, like the real gRPC Struct.
type memState struct {
	mu        sync.Mutex
	values    map[string]map[string]any
	delay     time.Duration // each call takes this long (virtual time under synctest)
	failGet   bool
	failSet   map[string]bool // keys whose writes fail
	sets      int
	setsByKey map[string]int
}

func newMemState() *memState {
	return &memState{values: map[string]map[string]any{}, failSet: map[string]bool{}, setsByKey: map[string]int{}}
}

var errStateDown = errors.New("state unavailable")

func (m *memState) wait(ctx context.Context) error {
	if m.delay == 0 {
		return nil
	}
	select {
	case <-time.After(m.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *memState) GetState(ctx context.Context, scope, scopeID, key string) (map[string]any, bool, error) {
	if err := m.wait(ctx); err != nil {
		return nil, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failGet {
		return nil, false, errStateDown
	}
	v, ok := m.values[scope+"/"+scopeID+"/"+key]
	if !ok {
		return nil, false, nil
	}
	return roundTrip(v), true, nil
}

func (m *memState) SetState(ctx context.Context, scope, scopeID, key string, value map[string]any) error {
	if err := m.wait(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failSet[key] {
		return errStateDown
	}
	m.sets++
	m.setsByKey[key]++
	m.values[scope+"/"+scopeID+"/"+key] = roundTrip(value)
	return nil
}

func (m *memState) raw(scope, scopeID, key string) map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.values[scope+"/"+scopeID+"/"+key]
}

func (m *memState) put(scope, scopeID, key string, v map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[scope+"/"+scopeID+"/"+key] = roundTrip(v)
}

func roundTrip(v map[string]any) map[string]any {
	b, _ := json.Marshal(v)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}
