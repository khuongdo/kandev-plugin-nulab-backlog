package connection

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Git credential input fields reported with CodeValidation (C5).
const (
	FieldGitUsername   = "gitUsername"
	FieldGitPassword   = "gitPassword"
	FieldGitCredential = "gitCredential" //nolint:gosec // G101: an input field name, not a credential
)

// ErrNoGitCredential is returned when no Git credential is stored for the
// connected space. Its text tells the user where to fix it.
var ErrNoGitCredential = errors.New("no Git access is stored for this Backlog space; update Git access in the Backlog settings")

var (
	errGitUsername = errors.New("git user name must be 1-100 characters without control characters")
	errGitPassword = errors.New("git password must be 1-256 characters")
)

const (
	maxGitUsername = 100
	maxGitPassword = 256
)

// GitCredentialInput is the connection.set_git_credential body.
type GitCredentialInput struct {
	Username string `json:"gitUsername"`
	Password string `json:"gitPassword"` //nolint:gosec // G117: request field; never logged or returned
}

// ValidateGitCredential trims the user name and checks both fields. The
// error never contains the input.
func ValidateGitCredential(in GitCredentialInput) (GitCredentialInput, error) {
	user := strings.TrimSpace(in.Username)
	if user == "" || utf8.RuneCountInString(user) > maxGitUsername || strings.IndexFunc(user, unicode.IsControl) >= 0 {
		return GitCredentialInput{}, &FieldError{Field: FieldGitUsername, Err: errGitUsername}
	}
	if in.Password == "" || len(in.Password) > maxGitPassword {
		return GitCredentialInput{}, &FieldError{Field: FieldGitPassword, Err: errGitPassword}
	}
	return GitCredentialInput{Username: user, Password: in.Password}, nil
}

// GitCredential is the Git user name and password for the connected space (C3).
type GitCredential struct {
	Username  string
	Password  string
	SpaceHost string
}

// String hides the password.
func (c GitCredential) String() string {
	return fmt.Sprintf("connection.GitCredential{Username: %s, SpaceHost: %s}", c.Username, c.SpaceHost)
}

// GoString hides the password for %#v.
func (c GitCredential) GoString() string { return c.String() }

// Format hides the password for every verb, including %+v.
func (c GitCredential) Format(f fmt.State, _ rune) { _, _ = fmt.Fprint(f, c.String()) }

// gitSecret is the secret backlog.git.<workspaceId>. It is bound to the
// space host it was saved for; Revision grows on every save.
type gitSecret struct {
	Username  string `json:"username"`
	Password  string `json:"password"` //nolint:gosec // G117: goes only to Kandev's encrypted secret store
	SpaceHost string `json:"spaceHost"`
	Revision  int    `json:"revision"`
}

// Git check results on the connection.test reply (AC5.5.2).
const (
	GitCheckOK       = "ok"
	GitCheckInvalid  = "invalid"
	GitCheckUntested = "untested"
)

// SetGitCredential validates and stores the Git user name and password for
// the connected space under the workspace write lock (AC5.5.1). The view
// says hasGitCredential and never holds the password.
func (s *Service) SetGitCredential(ctx context.Context, workspaceID string, in GitCredentialInput) (View, error) {
	ctx = redact.WithSecrets(ctx, in.Password)
	in, err := ValidateGitCredential(in)
	if err != nil {
		return View{}, err
	}
	unlock, err := s.lockWS(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	defer unlock()
	if err := s.store.SaveGit(ctx, workspaceID, in.Username, in.Password); err != nil {
		return View{}, err
	}
	redact.Logger(ctx).InfoContext(ctx, "git credential saved", "event", "git_credential_saved")
	return s.store.Load(ctx, workspaceID)
}

// GitCredential returns the Git credential of the connected space and its
// binding "<connectionEpoch>.<revision>", which changes on every save,
// replacement, project change, space change and disconnect (C3).
func (s *Service) GitCredential(ctx context.Context, workspaceID string) (GitCredential, string, error) {
	p, err := s.usable(ctx, workspaceID)
	if err != nil {
		return GitCredential{}, "", err
	}
	g, err := s.store.readGit(ctx, workspaceID)
	if err != nil {
		return GitCredential{}, "", err
	}
	if !g.matches(p.rec.SpaceHost) {
		return GitCredential{}, "", ErrNoGitCredential
	}
	return GitCredential{Username: g.Username, Password: g.Password, SpaceHost: g.SpaceHost},
		fmt.Sprintf("%d.%d", p.rec.ConnectionEpoch, g.Revision), nil
}

// gitCheck probes Git access on the first repository of the first selected
// project. Only a refusal is invalid; anything that stops the probe is untested.
func (s *Service) gitCheck(ctx context.Context, workspaceID string, creds backlog.Credentials, selected []string) string {
	g, _, err := s.GitCredential(ctx, workspaceID)
	if err != nil || len(selected) == 0 {
		return GitCheckUntested
	}
	repos, err := s.gateway.Repositories(ctx, creds, selected[0])
	if err != nil || len(repos) == 0 {
		return GitCheckUntested
	}
	ctx = redact.WithSecrets(ctx, g.Password)
	err = s.gateway.CheckGitAccess(ctx, g.SpaceHost, g.Username, g.Password, selected[0], repos[0].Name)
	var be *backlog.Error
	switch {
	case err == nil:
		return GitCheckOK
	case errors.As(err, &be) && be.Kind == backlog.KindUnauthorized:
		return GitCheckInvalid
	default:
		return GitCheckUntested
	}
}
