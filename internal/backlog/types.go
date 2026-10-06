package backlog

import (
	"encoding/json"
	"fmt"
	"time"
)

// Credentials are passed on every call; the client never stores them.
type Credentials struct {
	SpaceHost   string // a validated SpaceAddress host only
	APIKey      string // empty when OAuth is used (U2)
	AccessToken string // empty in U1
}

// String hides the secrets so Credentials can never leak through %v.
func (c Credentials) String() string {
	return fmt.Sprintf("backlog.Credentials{SpaceHost: %s}", c.SpaceHost)
}

// GoString hides the secrets for %#v as well.
func (c Credentials) GoString() string { return c.String() }

// User is the result of Myself (GET /api/v2/users/myself).
type User struct {
	ID     int64
	UserID string
	Name   string
}

// CallClass tells the client how long a caller may wait on a rate limit.
type CallClass int

const (
	// Interactive calls come from a user action and never wait on a rate limit in U1.
	Interactive CallClass = iota + 1
	// Background calls come from sync loops (U3 and later).
	Background
)

// Kind classifies every error the client returns.
type Kind int

// Error kinds (entities.md, C1 Error).
const (
	KindUnauthorized Kind = iota + 1
	KindForbidden
	KindRateLimited
	KindUnreachable
	KindNotFound
	KindInvalid
	KindConflict
)

var kindNames = map[Kind]string{
	KindUnauthorized: "unauthorized",
	KindForbidden:    "forbidden",
	KindRateLimited:  "rate_limited",
	KindUnreachable:  "unreachable",
	KindNotFound:     "not_found",
	KindInvalid:      "invalid",
	KindConflict:     "conflict",
}

func (k Kind) String() string {
	if n, ok := kindNames[k]; ok {
		return n
	}
	return "unknown"
}

// Error is the only error type the client returns. Its text holds only Kind
// and Status: never a body, a URL or a secret (BR3.3).
type Error struct {
	Kind       Kind
	Status     int           // HTTP status; 0 for network errors and timeouts
	RetryAfter time.Duration // set only for KindRateLimited
	Class      string        // http, timeout, dns, tls, connection, redirect, body, canceled
}

func (e *Error) Error() string {
	return fmt.Sprintf("backlog: %s (status %d)", e.Kind, e.Status)
}

// kindForStatus maps an HTTP status to a Kind. isErr is false for 2xx.
func kindForStatus(status int) (kind Kind, isErr bool) {
	switch {
	case status >= 200 && status < 300:
		return 0, false
	case status == 401:
		return KindUnauthorized, true
	case status == 403:
		return KindForbidden, true
	case status == 404:
		return KindNotFound, true
	case status == 409:
		return KindConflict, true
	case status == 400 || status == 422:
		return KindInvalid, true
	case status == 429:
		return KindRateLimited, true
	default:
		return KindUnreachable, true
	}
}

// parseUser decodes a Myself body. Anything that does not map to a User with a
// numeric id and a non-empty name is Unreachable (BR2.4).
func parseUser(body []byte) (User, error) {
	var raw struct {
		ID     *int64 `json:"id"`
		UserID string `json:"userId"`
		Name   string `json:"name"`
	}
	if err := json.Unmarshal(body, &raw); err != nil || raw.ID == nil || raw.Name == "" {
		return User{}, &Error{Kind: KindUnreachable, Status: 200, Class: "body"}
	}
	return User{ID: *raw.ID, UserID: raw.UserID, Name: raw.Name}, nil
}
