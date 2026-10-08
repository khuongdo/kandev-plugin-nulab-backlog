package scm

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

const (
	cliTTL       = 5 * time.Minute  // how long a CLI token is reused (FR3.2, NFR4)
	cliTimeout   = 10 * time.Second // limit per CLI run (NFR2)
	maxCLIOutput = 4096             // stdout bytes read from the CLI (NFR2)
)

var errNoCLI = errors.New("the CLI login works for GitHub and GitLab only")

// CLIRunner runs name with args and returns its stdout. Tests replace it so
// no real gh or glab runs (NFR3).
type CLIRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// cliCommand is the fixed command that prints p's token (NFR2): gh ≥ 2.17
// and glab's stored login for gitlab.com.
func cliCommand(p Provider) (string, []string, bool) {
	switch p {
	case GitHub:
		return "gh", []string{"auth", "token", "--hostname", "github.com"}, true
	case GitLab:
		return "glab", []string{"config", "get", "token", "--host", "gitlab.com"}, true
	}
	return "", nil, false
}

// runCLI is the production CLIRunner: no shell, a 10-second limit, at most
// 4 KiB of stdout, stderr discarded (NFR1, NFR2).
func runCLI(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, cliTimeout)
	defer cancel()
	out := &cappedBuffer{}
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: fixed arguments from cliCommand
	cmd.Stdout = out
	err := cmd.Run()
	return out.b, err
}

// cappedBuffer keeps the first maxCLIOutput bytes and drops the rest.
type cappedBuffer struct{ b []byte }

func (c *cappedBuffer) Write(p []byte) (int, error) {
	if room := maxCLIOutput - len(c.b); room > 0 {
		c.b = append(c.b, p[:min(len(p), room)]...)
	}
	return len(p), nil
}

type cachedToken struct {
	token string
	at    time.Time
}

// cliCache holds the last CLI token per provider. The CLI login belongs to
// the Kandev server, so the cache is not per workspace.
// ponytail: one lock for every provider and the CLI run under it, so at most
// one CLI process runs at a time; split per provider if that ever waits.
type cliCache struct {
	mu     sync.Mutex
	tokens map[Provider]cachedToken
}

// cliToken returns p's token from the CLI, reused for cliTTL (FR3.1, FR3.2).
// Every failure is ErrCLIUnavailable, never the CLI's output (FR4.1, NFR1).
func (s *Service) cliToken(ctx context.Context, p Provider) (string, error) {
	name, args, ok := cliCommand(p)
	if !ok {
		return "", &connection.FieldError{Field: FieldProvider, Err: errNoCLI}
	}
	s.cli.mu.Lock()
	defer s.cli.mu.Unlock()
	if c, ok := s.cli.tokens[p]; ok && s.Now().Sub(c.at) < cliTTL {
		return c.token, nil
	}
	out, err := s.CLI(ctx, name, args...)
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	tok := strings.TrimSpace(string(out))
	if err != nil || !validToken(tok) {
		return "", ErrCLIUnavailable
	}
	s.cli.tokens[p] = cachedToken{token: tok, at: s.Now()}
	return tok, nil
}

// forgetCLI drops p's cached CLI token, so the next call asks the CLI.
func (s *Service) forgetCLI(p Provider) {
	s.cli.mu.Lock()
	defer s.cli.mu.Unlock()
	delete(s.cli.tokens, p)
}
