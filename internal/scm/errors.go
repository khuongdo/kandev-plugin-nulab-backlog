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
	// ErrCLIUnavailable means gh or glab is missing, not logged in, timed out
	// or printed no usable token on the Kandev server (FR4.1). It never holds
	// the CLI's output (NFR1).
	ErrCLIUnavailable = errors.New("the gh or glab CLI is not available or not logged in on the Kandev server")
	// ErrCLIAccountMissing means gh has no login for the workspace's chosen
	// GitHub account; the plugin never falls back to another one (FR5.1).
	ErrCLIAccountMissing = errors.New("the chosen GitHub account is not logged in to gh on the Kandev server")
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

// serviceNames are the services' display names, for error texts.
var serviceNames = map[Provider]string{BacklogGit: "Backlog Git", GitHub: "GitHub", GitLab: "GitLab", Bitbucket: "Bitbucket"}

// InactiveError refuses an action of a service that is not the workspace's
// active source control service (FR1.4). Active is the active one.
type InactiveError struct {
	Active Provider
}

func (e *InactiveError) Error() string {
	return fmt.Sprintf("this workspace uses %s for source control; change the service under Source control in the Backlog settings first",
		serviceNames[e.Active])
}
