package scm

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

const (
	cliTTL           = 5 * time.Minute  // how long a CLI token is reused (FR3.2, NFR4)
	cliTimeout       = 10 * time.Second // limit per CLI run (NFR2)
	maxCLIOutput     = 4096             // stdout bytes read from a token read (NFR2)
	maxCLIListOutput = 32 << 10         // stdout bytes read from `gh auth status --json` (AC1.1.8)
	maxLogin         = 39               // GitHub's login length limit (AC1.1.5)
)

var (
	errNoCLI = errors.New("the CLI login works for GitHub and GitLab only")
	// errCLIOutputCut: the CLI printed more than the runner reads.
	errCLIOutputCut = errors.New("CLI output over the size limit")
	loginRule       = regexp.MustCompile(`^[A-Za-z0-9](?:-?[A-Za-z0-9])*$`)
	// tokenEnv are the variables that make gh ignore its stored logins (AC1.1.6).
	tokenEnv = []string{"GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN"}
)

// Fixed gh commands (NFR2, FR3.3): gh never switches its active account.
// `--user` needs gh ≥ 2.40; `auth status --json` needs gh ≥ 2.81.0.
var (
	ghTokenArgs  = []string{"auth", "token", "--hostname", "github.com"}
	ghStatusArgs = []string{"auth", "status", "--json", "hosts", "--hostname", "github.com"}
	glabArgs     = []string{"config", "get", "token", "--host", "gitlab.com"}
)

// CLIRunner runs name with args and returns at most limit bytes of its
// stdout. Tests replace it so no real gh or glab runs (NFR3).
type CLIRunner func(ctx context.Context, limit int, name string, args ...string) ([]byte, error)

// CLIAccount is one github.com login stored by gh (FR1.1). It holds no token.
type CLIAccount struct {
	Login  string `json:"login"`
	Active bool   `json:"active"`
}

// validLogin is GitHub's login rule (AC1.1.5).
func validLogin(s string) bool { return len(s) <= maxLogin && loginRule.MatchString(s) }

// cliEnv is env without the token variables (AC1.1.6, AC2.1.8).
func cliEnv(env []string) []string {
	return slices.DeleteFunc(slices.Clone(env), func(kv string) bool {
		k, _, _ := strings.Cut(kv, "=")
		return slices.ContainsFunc(tokenEnv, func(t string) bool { return strings.EqualFold(k, t) })
	})
}

// runCLI is the production CLIRunner: no shell, a 10-second limit, at most
// limit bytes of stdout (errCLIOutputCut beyond), stderr discarded, no token
// variables (NFR1, NFR2).
func runCLI(ctx context.Context, limit int, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, cliTimeout)
	defer cancel()
	out := &cappedBuffer{limit: limit}
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: fixed arguments; the login is validated
	cmd.Env = cliEnv(os.Environ())
	cmd.Stdout = out
	err := cmd.Run()
	if err == nil && out.cut {
		err = errCLIOutputCut
	}
	return out.b, err
}

// cappedBuffer keeps the first limit bytes and notes that more came.
type cappedBuffer struct {
	b     []byte
	limit int
	cut   bool
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	room := c.limit - len(c.b)
	c.b = append(c.b, p[:max(min(len(p), room), 0)]...)
	c.cut = c.cut || len(p) > room
	return len(p), nil
}

type cliKey struct {
	p     Provider
	login string // GitHub only
}

type cachedToken struct {
	token string
	at    time.Time
}

// cliCache holds the last CLI token per provider and login. The CLI logins
// belong to the Kandev server; a workspace picks one by its AccountID.
// ponytail: one lock for every provider and the CLI runs (and the old-gh
// /user check) under it, so at most one CLI process runs at a time; split
// per key if that ever waits.
type cliCache struct {
	mu     sync.Mutex
	tokens map[cliKey]cachedToken
}

// cliToken returns p's token from the CLI, for GitHub the token of login,
// reused for cliTTL (FR3.1, FR3.2). A failure is ErrCLIUnavailable or, for
// a login gh does not have, ErrCLIAccountMissing; never the CLI's output
// (FR5.1, NFR1).
func (s *Service) cliToken(ctx context.Context, p Provider, login string) (string, error) {
	switch p {
	case GitHub:
		if !validLogin(login) {
			return "", ErrCLIAccountMissing
		}
	case GitLab:
		login = ""
	default:
		return "", &connection.FieldError{Field: FieldProvider, Err: errNoCLI}
	}
	s.cli.mu.Lock()
	defer s.cli.mu.Unlock()
	key := cliKey{p, login}
	if c, ok := s.cli.tokens[key]; ok && s.Now().Sub(c.at) < cliTTL {
		return c.token, nil
	}
	var tok string
	var err error
	if p == GitHub {
		tok, err = s.ghToken(ctx, login)
	} else {
		tok, err = s.readToken(ctx, "glab", glabArgs...)
	}
	if err != nil {
		return "", err
	}
	s.cli.tokens[key] = cachedToken{token: tok, at: s.Now()}
	return tok, nil
}

// readToken runs a token command; anything but one plain token is
// ErrCLIUnavailable.
func (s *Service) readToken(ctx context.Context, name string, args ...string) (string, error) {
	out, err := s.CLI(ctx, maxCLIOutput, name, args...)
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	tok := strings.TrimSpace(string(out))
	if err != nil || !validToken(tok) {
		return "", ErrCLIUnavailable
	}
	return tok, nil
}

// ghToken reads login's token with `--user`. When that fails, the account
// list decides (AC3.1.5): no list is unavailable, an unlisted login is
// missing, and a listed one (gh before 2.40 has no --user) gets the active
// account's token only when GitHub says it is login (FR3.4).
func (s *Service) ghToken(ctx context.Context, login string) (string, error) {
	tok, err := s.readToken(ctx, "gh", append(slices.Clone(ghTokenArgs), "--user", login)...)
	if err == nil || ctx.Err() != nil {
		return tok, err
	}
	accounts, err := s.ghAccounts(ctx)
	if err != nil {
		return "", err
	}
	if !slices.ContainsFunc(accounts, func(a CLIAccount) bool { return strings.EqualFold(a.Login, login) }) {
		return "", ErrCLIAccountMissing
	}
	tok, active, err := s.ghActive(ctx)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(active, login) {
		return "", ErrCLIAccountMissing
	}
	return tok, nil
}

// ListCLIAccounts lists gh's usable github.com logins (FR1.1).
func (s *Service) ListCLIAccounts(ctx context.Context) ([]CLIAccount, error) {
	s.cli.mu.Lock()
	defer s.cli.mu.Unlock()
	return s.ghAccounts(ctx)
}

// ghAccounts reads `gh auth status --json`, using stdout even after a
// non-zero exit; only accounts in the success state with a valid login are
// kept (AC1.1.8). A gh without --json offers its active account (AC1.1.9).
func (s *Service) ghAccounts(ctx context.Context) ([]CLIAccount, error) {
	out, err := s.CLI(ctx, maxCLIListOutput, "gh", ghStatusArgs...)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if errors.Is(err, errCLIOutputCut) {
		return nil, ErrCLIUnavailable
	}
	var st struct {
		Hosts map[string][]struct {
			Login  string `json:"login"`
			State  string `json:"state"`
			Active bool   `json:"active"`
		} `json:"hosts"`
	}
	if json.Unmarshal(out, &st) != nil {
		_, login, err := s.ghActive(ctx)
		if err != nil {
			return nil, err
		}
		return []CLIAccount{{Login: login, Active: true}}, nil
	}
	var accounts []CLIAccount
	for _, a := range st.Hosts["github.com"] {
		if a.State == "success" && validLogin(a.Login) {
			accounts = append(accounts, CLIAccount{Login: a.Login, Active: a.Active})
		}
	}
	if len(accounts) == 0 {
		return nil, ErrCLIUnavailable
	}
	return accounts, nil
}

// ghActive reads gh's active token and asks GitHub whose it is.
func (s *Service) ghActive(ctx context.Context) (tok, login string, err error) {
	tok, err = s.readToken(ctx, "gh", ghTokenArgs...)
	if err != nil {
		return "", "", err
	}
	ctx = redact.WithSecrets(ctx, tok)
	user, err := s.clients[GitHub].CurrentUser(ctx, Credential{Token: tok})
	if ctx.Err() != nil {
		return "", "", ctx.Err()
	}
	if err != nil || !validLogin(user.ID) {
		return "", "", ErrCLIUnavailable
	}
	return tok, user.ID, nil
}

// forgetCLI drops every cached CLI token of p, so the next call asks the CLI.
func (s *Service) forgetCLI(p Provider) {
	s.cli.mu.Lock()
	defer s.cli.mu.Unlock()
	maps.DeleteFunc(s.cli.tokens, func(k cliKey, _ cachedToken) bool { return k.p == p })
}
