package connection

import (
	"context"
	"log/slog"
	"slices"
	"sync"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Reason says why a ConnectionChanged event was sent (contract C3).
type Reason string

// Change reasons.
const (
	ReasonConnected           Reason = "connected"
	ReasonCredentialsReplaced Reason = "credentials_replaced"
	ReasonSpaceChanged        Reason = "space_changed"
	ReasonDisconnected        Reason = "disconnected"
	ReasonProjectsChanged     Reason = "projects_changed"
)

// ConnectionChanged is the in-process event U3 and U4 subscribe to (C3). It
// carries no secret.
type ConnectionChanged struct {
	WorkspaceID      string
	Reason           Reason
	ConnectionEpoch  int
	SpaceHost        string
	SelectedProjects []string
	// Restore is true when items from an earlier connection to this space,
	// or of a reselected project, should come back (M12).
	Restore bool
}

// changeFor compares the record before and after a write.
func changeFor(prev, next record) (Reason, bool) {
	switch {
	case next.Disconnected:
		return ReasonDisconnected, false
	case prev.SpaceHost == "" || prev.Disconnected:
		return ReasonConnected, prev.PreviousSpaceHost != "" && next.SpaceHost == prev.PreviousSpaceHost
	case next.SpaceHost != prev.SpaceHost:
		return ReasonSpaceChanged, next.SpaceHost == prev.PreviousSpaceHost
	case !slices.Equal(prev.SelectedProjects, next.SelectedProjects):
		return ReasonProjectsChanged, added(prev.SelectedProjects, next.SelectedProjects)
	default:
		return ReasonCredentialsReplaced, false
	}
}

// added reports whether next holds a key that prev does not.
func added(prev, next []string) bool {
	for _, k := range next {
		if !slices.Contains(prev, k) {
			return true
		}
	}
	return false
}

// emit sends the ConnectionChanged event for a successful write and logs it.
// The caller holds the workspace write lock, so events keep epoch order. It
// returns the event's Restore flag.
func (s *Service) emit(ctx context.Context, workspaceID string, prev, next record) bool {
	reason, restore := changeFor(prev, next)
	s.publish(ctx, ConnectionChanged{
		WorkspaceID: workspaceID, Reason: reason, ConnectionEpoch: next.ConnectionEpoch,
		SpaceHost: next.SpaceHost, SelectedProjects: next.SelectedProjects, Restore: restore,
	})
	return restore
}

func (s *Service) publish(ctx context.Context, e ConnectionChanged) {
	log := redact.Logger(ctx)
	log.InfoContext(ctx, "connection changed", "event", "connection_changed", "reason", string(e.Reason),
		"connectionEpoch", e.ConnectionEpoch, "restore", e.Restore, "projectCount", len(e.SelectedProjects))
	s.hub.publish(delivery{e: e, log: log})
}

// Subscribe registers fn for every ConnectionChanged event. Each subscriber
// has its own queue and goroutine, so a slow one never blocks an action and
// a panicking one is recovered. Call the returned func to stop.
func (s *Service) Subscribe(fn func(ConnectionChanged)) (unsubscribe func()) {
	sub := &subscriber{fn: fn, wake: make(chan struct{}, 1), done: make(chan struct{})}
	id := s.hub.add(sub)
	go sub.run()
	var once sync.Once
	return func() {
		once.Do(func() {
			s.hub.remove(id)
			close(sub.done)
		})
	}
}

type delivery struct {
	e   ConnectionChanged
	log *slog.Logger
}

// hub holds the subscribers.
type hub struct {
	mu   sync.Mutex
	subs map[int]*subscriber
	next int
}

func (h *hub) add(sub *subscriber) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs == nil {
		h.subs = map[int]*subscriber{}
	}
	h.next++
	h.subs[h.next] = sub
	return h.next
}

func (h *hub) remove(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs, id)
}

func (h *hub) publish(d delivery) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, sub := range h.subs {
		sub.push(d)
	}
}

// subscriber is one unbounded FIFO queue drained by its own goroutine.
// ponytail: unbounded queue; events are rare (one per user action), cap it if a subscriber can stall for long.
type subscriber struct {
	fn    func(ConnectionChanged)
	mu    sync.Mutex
	queue []delivery
	wake  chan struct{}
	done  chan struct{}
}

func (s *subscriber) push(d delivery) {
	s.mu.Lock()
	s.queue = append(s.queue, d)
	s.mu.Unlock()
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *subscriber) pop() (delivery, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.queue) == 0 {
		return delivery{}, false
	}
	d := s.queue[0]
	s.queue = s.queue[1:]
	return d, true
}

func (s *subscriber) run() {
	for {
		select {
		case <-s.done:
			return
		case <-s.wake:
		}
		for d, ok := s.pop(); ok; d, ok = s.pop() {
			s.deliver(d)
		}
	}
}

func (s *subscriber) deliver(d delivery) {
	defer func() {
		if p := recover(); p != nil {
			d.log.Error("connection subscriber panicked", "event", "connection_subscriber_panic", "reason", string(d.e.Reason))
		}
	}()
	s.fn(d.e)
}
