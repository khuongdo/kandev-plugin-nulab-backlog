package scm

import (
	"errors"
	"fmt"
	"time"
)

// Errors the plugin maps to action error codes.
var (
	// ErrHostRefused is a request to anything but the provider's https host (NFR2).
	ErrHostRefused = errors.New("scm: refusing a request outside the provider host")
	// ErrNotFound is a query, watch, link or pull request that does not exist (not_found).
	ErrNotFound = errors.New("not found")
	// ErrConflict is an action the item's state does not allow (conflict).
	ErrConflict = errors.New("not allowed in the current state")
	// ErrNoToken means the provider has no token in this workspace; its text
	// says where to add one.
	ErrNoToken = errors.New("no access token is stored for this provider; add one under Source control in the Backlog settings")
)

// HTTPError is a failed provider call (NFR5), shared by the three clients.
// Status 0 means the provider could not be reached or its reply was not
// usable. Its text holds only the provider and status: never a response
// body, a URL or a secret.
type HTTPError struct {
	Provider   Provider
	Status     int           // 429 also for GitHub's out-of-quota 403 (NFR3)
	RetryAfter time.Duration // set for 429 only
}

func (e *HTTPError) Error() string {
	if e.Status == 0 {
		return fmt.Sprintf("%s: unreachable", e.Provider)
	}
	return fmt.Sprintf("%s: HTTP %d", e.Provider, e.Status)
}
