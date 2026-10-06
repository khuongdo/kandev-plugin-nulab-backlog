package connection

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// ErrInvalidSpaceAddress is returned for any address that is not a bare https
// Backlog space host (BR1.1, BR1.2). Its text states the expected format.
var ErrInvalidSpaceAddress = errors.New("space address must look like <space>.backlog.com, <space>.backlog.jp or <space>.backlogtool.com")

// spaceHostPattern is BR1.1: one ASCII label of 1-63 characters followed by an
// allowed suffix. It runs on the parsed, lower-cased host only.
var spaceHostPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.(backlog\.com|backlog\.jp|backlogtool\.com)$`)

// SpaceAddress is a validated, normalised Backlog space host.
type SpaceAddress struct {
	Host string
}

// BaseURL is always https://<host>.
func (a SpaceAddress) BaseURL() string {
	return "https://" + a.Host
}

// ParseSpaceAddress normalises and validates raw user input (BR1.2, then BR1.1).
func ParseSpaceAddress(raw string) (SpaceAddress, error) {
	s := strings.TrimSpace(raw)
	if s == "" || !isASCII(s) {
		return SpaceAddress{}, ErrInvalidSpaceAddress
	}
	if scheme, _, found := strings.Cut(s, "://"); found {
		if !strings.EqualFold(scheme, "https") {
			return SpaceAddress{}, ErrInvalidSpaceAddress
		}
	} else {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Opaque != "" || u.User != nil || u.Port() != "" ||
		u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(s, "#") ||
		(u.Path != "" && u.Path != "/") {
		return SpaceAddress{}, ErrInvalidSpaceAddress
	}
	host := strings.ToLower(u.Hostname())
	if strings.HasSuffix(host, ".") || net.ParseIP(host) != nil || !spaceHostPattern.MatchString(host) {
		return SpaceAddress{}, ErrInvalidSpaceAddress
	}
	return SpaceAddress{Host: host}, nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
