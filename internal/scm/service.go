package scm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// Connection is the Backlog connection the service reads: the selected
// projects and the on/off switch. connection.Service implements it. The
// service never subscribes to ConnectionChanged, so its items survive a
// space change, a disconnect and a project deselection (FR6.2).
type Connection interface {
	Current(ctx context.Context, workspaceID string) (connection.Snapshot, error)
	RequireEnabled(ctx context.Context, workspaceID string) error
}

// Secrets is Kandev's secret store, where the tokens live (NFR1).
type Secrets interface {
	GetSecret(ctx context.Context, key string) (string, bool, error)
	SetSecret(ctx context.Context, key, value string) error
	DeleteSecret(ctx context.Context, key string) error
}

// Tasks creates the Kandev task of a watch. internal/plugin implements it.
type Tasks interface {
	CreateTask(ctx context.Context, in NewTask) (taskID string, err error)
}

// NewTask is a Kandev task a watch creates.
type NewTask struct {
	WorkspaceID    string
	WorkflowID     string
	WorkflowStepID string
	Title          string
	Description    string
	Metadata       map[string]any
}

// MetadataKey is the task metadata key a watch task carries: its PR key.
const MetadataKey = "nulab_backlog_scm_pr"

// Provider states shown in Settings (FR2.1).
const (
	StateNotConfigured = "not_configured"
	StateConnected     = "connected"
	StateError         = "error"
)

// Credential methods shown in Settings (FR5.1).
const (
	MethodToken = "token" // a token typed in Settings, kept in the secret store
	MethodCLI   = "cli"   // read from gh / glab on the Kandev server
)

// Last test errors, shown as plain text in Settings (FR2.4).
const (
	ErrorInvalidToken = "invalid_token"
	ErrorMissingScope = "missing_scope"
	ErrorRateLimited  = "rate_limited"
	ErrorUnreachable  = "unreachable"
	// ErrorCLIUnavailable: gh / glab is missing or not logged in (FR4.2).
	ErrorCLIUnavailable = "cli_unavailable"
)

const (
	maxToken           = 1024
	maxUsername        = 200
	maxReposPerProject = 20
	prPageSize         = 20  // a PR list page, like Backlog Git (FR4.1)
	watchPRs           = 100 // pull requests one watch run reads
)

var (
	errToken    = errors.New("must be 1-1024 characters without spaces")
	errUsername = errors.New("the Bitbucket user name is required")
	errProject  = errors.New("must be a selected Backlog project")
	errRepos    = errors.New("must be at most 20 readable repositories")
)

// Service is the source control component.
type Service struct {
	clients map[Provider]Client
	conn    Connection
	secrets Secrets
	tasks   Tasks
	store   *Store
	Now     func() time.Time // clock for watch runs and the CLI token cache
	CLI     CLIRunner        // runs gh / glab (FR3.1)
	cli     cliCache
}

// NewService wires the component; clients has one Client per provider.
func NewService(clients map[Provider]Client, conn Connection, secrets Secrets, tasks Tasks, store *Store) *Service {
	return &Service{clients: clients, conn: conn, secrets: secrets, tasks: tasks, store: store, Now: time.Now,
		CLI: runCLI, cli: cliCache{tokens: map[Provider]cachedToken{}}}
}

// SecretKey is where a provider's token is stored: backlog.scm.<provider>.<workspace>.
func SecretKey(p Provider, ws string) string { return "backlog.scm." + string(p) + "." + ws }

// ProviderView is one provider in Settings. It never holds the token (NFR1).
type ProviderView struct {
	Provider  Provider  `json:"provider"`
	State     string    `json:"state"`
	Method    string    `json:"method,omitempty"` // MethodToken or MethodCLI; empty when not configured
	Account   string    `json:"account,omitempty"`
	LastError string    `json:"lastError,omitempty"`
	Mappings  []Mapping `json:"mappings"`
}

func view(s Settings) ProviderView {
	v := ProviderView{Provider: s.Provider, State: StateNotConfigured, Account: s.Account, LastError: s.LastError,
		Mappings: s.Mappings}
	switch {
	case !s.HasToken:
	case s.Source == MethodCLI:
		v.Method = MethodCLI
	default:
		v.Method = MethodToken
	}
	switch {
	case !s.HasToken:
	case s.LastError != "":
		v.State = StateError
	default:
		v.State = StateConnected
	}
	if v.Mappings == nil {
		v.Mappings = []Mapping{}
	}
	return v
}

func settingsOf(list []Settings, p Provider) Settings {
	if i := slices.IndexFunc(list, func(s Settings) bool { return s.Provider == p }); i >= 0 {
		return list[i]
	}
	return Settings{Provider: p}
}

func (s *Service) settings(ctx context.Context, ws string, p Provider) (Settings, error) {
	list, err := s.store.Settings(ctx, ws)
	return settingsOf(list, p), err
}

// Providers lists the three providers with their state (FR2.1).
func (s *Service) Providers(ctx context.Context, ws string) ([]ProviderView, error) {
	list, err := s.store.Settings(ctx, ws)
	if err != nil {
		return nil, err
	}
	out := make([]ProviderView, 0, len(Providers))
	for _, p := range Providers {
		out = append(out, view(settingsOf(list, p)))
	}
	return out, nil
}

// updateProvider runs fn on p's settings in one write and returns the view.
func (s *Service) updateProvider(ctx context.Context, ws string, p Provider, fn func(*Settings) error) (ProviderView, error) {
	var out Settings
	err := s.store.UpdateSettings(ctx, ws, func(list []Settings) ([]Settings, error) {
		i := slices.IndexFunc(list, func(x Settings) bool { return x.Provider == p })
		if i < 0 {
			list, i = append(list, Settings{Provider: p}), len(list)
		}
		if err := fn(&list[i]); err != nil {
			return nil, err
		}
		out = list[i]
		return list, nil
	})
	return view(out), err
}

// TokenInput is the scm.providers.set_token body.
type TokenInput struct {
	Provider Provider `json:"provider"`
	Token    string   `json:"token"` //nolint:gosec // G117: request field; stored only as a secret
	Username string   `json:"username,omitempty"`
}

func (in *TokenInput) validate() error {
	if _, err := ParseProvider(string(in.Provider)); err != nil {
		return err
	}
	if !validToken(in.Token) {
		return &connection.FieldError{Field: FieldToken, Err: errToken}
	}
	in.Username = strings.TrimSpace(in.Username)
	if in.Provider != Bitbucket {
		in.Username = ""
		return nil
	}
	if in.Username == "" || len(in.Username) > maxUsername || strings.IndexFunc(in.Username, unicode.IsControl) >= 0 {
		return &connection.FieldError{Field: FieldUsername, Err: errUsername}
	}
	return nil
}

// validToken is 1-1024 characters without spaces or control characters.
func validToken(tok string) bool {
	return tok != "" && len(tok) <= maxToken && strings.IndexFunc(tok, unicode.IsSpace) < 0 &&
		strings.IndexFunc(tok, unicode.IsControl) < 0
}

// SetToken checks the token with the provider's current user call and only
// then stores it as a secret, with the account name (FR2.2, FR2.4).
func (s *Service) SetToken(ctx context.Context, ws string, in TokenInput) (ProviderView, error) {
	ctx = redact.WithSecrets(ctx, in.Token)
	if err := in.validate(); err != nil {
		return ProviderView{}, err
	}
	cred := Credential{Token: in.Token, Username: in.Username}
	user, err := s.clients[in.Provider].CurrentUser(ctx, cred)
	if err != nil {
		return ProviderView{}, err
	}
	b, err := json.Marshal(cred)
	if err == nil {
		err = s.secrets.SetSecret(ctx, SecretKey(in.Provider, ws), string(b))
	}
	if err != nil {
		return ProviderView{}, storeErr(ctx, "save token", err)
	}
	s.forgetCLI(in.Provider)
	redact.Logger(ctx).InfoContext(ctx, "scm token saved", "event", "scm_token_saved", "provider", string(in.Provider))
	return s.updateProvider(ctx, ws, in.Provider, func(st *Settings) error {
		st.Source, st.HasToken, st.Account, st.AccountID, st.LastError = "", true, user.Name, user.ID, ""
		return nil
	})
}

// UseCLI connects p with the gh / glab login of the Kandev server (FR1.2,
// FR2.1): the CLI token is read, checked with the current user call, and
// only the method and account are stored. A typed token is deleted (FR1.3).
// A failure changes nothing (FR4.1).
func (s *Service) UseCLI(ctx context.Context, ws string, p Provider) (ProviderView, error) {
	if _, err := ParseProvider(string(p)); err != nil {
		return ProviderView{}, err
	}
	s.forgetCLI(p)
	tok, err := s.cliToken(ctx, p)
	if err != nil {
		return ProviderView{}, err
	}
	ctx = redact.WithSecrets(ctx, tok)
	user, err := s.clients[p].CurrentUser(ctx, Credential{Token: tok})
	if err != nil {
		s.forgetCLI(p)
		return ProviderView{}, err
	}
	if err := s.secrets.DeleteSecret(ctx, SecretKey(p, ws)); err != nil {
		return ProviderView{}, storeErr(ctx, "delete token", err)
	}
	redact.Logger(ctx).InfoContext(ctx, "scm cli login used", "event", "scm_cli_used", "provider", string(p))
	return s.updateProvider(ctx, ws, p, func(st *Settings) error {
		st.Source, st.HasToken, st.Account, st.AccountID, st.LastError = MethodCLI, true, user.Name, user.ID, ""
		return nil
	})
}

// credential reads p's token, from the CLI when that is p's method (FR3.1),
// and returns a context that redacts it.
func (s *Service) credential(ctx context.Context, ws string, p Provider) (context.Context, Credential, error) {
	st, err := s.settings(ctx, ws, p)
	if err != nil {
		return ctx, Credential{}, err
	}
	if st.Source == MethodCLI && st.HasToken {
		tok, err := s.cliToken(ctx, p)
		if err != nil {
			return ctx, Credential{}, err
		}
		return redact.WithSecrets(ctx, tok), Credential{Token: tok}, nil
	}
	raw, ok, err := s.secrets.GetSecret(ctx, SecretKey(p, ws))
	if err != nil {
		return ctx, Credential{}, storeErr(ctx, "read token", err)
	}
	if !ok {
		return ctx, Credential{}, ErrNoToken
	}
	var c Credential
	if err := json.Unmarshal([]byte(raw), &c); err != nil || c.Token == "" {
		return ctx, Credential{}, fmt.Errorf("read token: unsupported value: %w", connection.ErrStore)
	}
	return redact.WithSecrets(ctx, c.Token), c, nil
}

// Test calls the provider's current user endpoint (FR2.4). A refusal is a
// result: the view's state is error and lastError says why.
func (s *Service) Test(ctx context.Context, ws string, p Provider) (ProviderView, error) {
	if _, err := ParseProvider(string(p)); err != nil {
		return ProviderView{}, err
	}
	s.forgetCLI(p) // a test always asks the CLI again (FR4.3)
	ctx, cred, err := s.credential(ctx, ws, p)
	var user User
	switch {
	case errors.Is(err, ErrCLIUnavailable): // recorded below (FR4.2)
	case err != nil:
		return ProviderView{}, err
	default:
		user, err = s.clients[p].CurrentUser(ctx, cred)
		if err != nil && ctx.Err() != nil {
			return ProviderView{}, ctx.Err()
		}
		if err != nil {
			s.forgetCLI(p)
		}
	}
	return s.updateProvider(ctx, ws, p, func(st *Settings) error {
		if err != nil {
			st.LastError = errorCode(err)
			return nil
		}
		st.Account, st.AccountID, st.LastError = user.Name, user.ID, ""
		return nil
	})
}

// errorCode is the plain reason of a failed provider call.
func errorCode(err error) string {
	switch {
	case errors.Is(err, ErrCLIUnavailable):
		return ErrorCLIUnavailable
	case IsStatus(err, 401):
		return ErrorInvalidToken
	case IsStatus(err, 403):
		return ErrorMissingScope
	case IsStatus(err, 429):
		return ErrorRateLimited
	default:
		return ErrorUnreachable
	}
}

// RemoveToken deletes the secret. Mappings, links, queries and watches stay,
// disabled until a token exists again (FR2.2).
func (s *Service) RemoveToken(ctx context.Context, ws string, p Provider) (ProviderView, error) {
	if _, err := ParseProvider(string(p)); err != nil {
		return ProviderView{}, err
	}
	if err := s.secrets.DeleteSecret(ctx, SecretKey(p, ws)); err != nil {
		return ProviderView{}, storeErr(ctx, "delete token", err)
	}
	s.forgetCLI(p)
	return s.updateProvider(ctx, ws, p, func(st *Settings) error {
		st.Source, st.HasToken, st.Account, st.AccountID, st.LastError = "", false, "", "", ""
		return nil
	})
}

// SearchRepos lists the repositories p's token can read (FR3.2).
func (s *Service) SearchRepos(ctx context.Context, ws string, p Provider, query string) ([]Repo, error) {
	if _, err := ParseProvider(string(p)); err != nil {
		return nil, err
	}
	ctx, cred, err := s.credential(ctx, ws, p)
	if err != nil {
		return nil, err
	}
	return s.clients[p].SearchRepos(ctx, cred, strings.TrimSpace(query))
}

// MappingInput is the scm.mappings.set body: the whole repository list of
// one Backlog project. An empty list removes the mapping.
type MappingInput struct {
	Provider   Provider `json:"provider"`
	ProjectKey string   `json:"projectKey"`
	Repos      []string `json:"repos"`
}

// SetMapping maps a selected Backlog project to repositories, each checked
// with the provider and saved in its spelling (FR3.1, FR3.2). Queries and
// watches of a removed repository are disabled, not deleted (FR3.4).
func (s *Service) SetMapping(ctx context.Context, ws string, in MappingInput) (ProviderView, error) {
	if _, err := ParseProvider(string(in.Provider)); err != nil {
		return ProviderView{}, err
	}
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return ProviderView{}, err
	}
	if !slices.Contains(snap.SelectedProjects, in.ProjectKey) {
		return ProviderView{}, &connection.FieldError{Field: FieldProjectKey, Err: errProject}
	}
	repos, err := s.checkRepos(ctx, ws, in)
	if err != nil {
		return ProviderView{}, err
	}
	return s.updateProvider(ctx, ws, in.Provider, func(st *Settings) error {
		st.Mappings = slices.DeleteFunc(st.Mappings, func(m Mapping) bool { return m.ProjectKey == in.ProjectKey })
		if len(repos) > 0 {
			st.Mappings = append(st.Mappings, Mapping{ProjectKey: in.ProjectKey, Repos: repos})
		}
		return nil
	})
}

// checkRepos validates the names, then reads each repository once.
func (s *Service) checkRepos(ctx context.Context, ws string, in MappingInput) ([]string, error) {
	bad := &connection.FieldError{Field: FieldRepos, Err: errRepos}
	if len(in.Repos) > maxReposPerProject {
		return nil, bad
	}
	for _, r := range in.Repos {
		if !ValidRepo(in.Provider, strings.TrimSpace(r)) {
			return nil, bad
		}
	}
	if len(in.Repos) == 0 {
		return nil, nil
	}
	ctx, cred, err := s.credential(ctx, ws, in.Provider)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, r := range in.Repos {
		repo, err := s.clients[in.Provider].GetRepo(ctx, cred, strings.TrimSpace(r))
		if IsStatus(err, 404) || (err == nil && !ValidRepo(in.Provider, repo.FullName)) {
			return nil, bad
		}
		if err != nil {
			return nil, err
		}
		if !slices.Contains(out, repo.FullName) {
			out = append(out, repo.FullName)
		}
	}
	return out, nil
}

// mappedProjects lists the projects repo is mapped to.
func mappedProjects(st Settings, repo string) []string {
	var out []string
	for _, m := range st.Mappings {
		if slices.ContainsFunc(m.Repos, func(r string) bool { return strings.EqualFold(r, repo) }) {
			out = append(out, m.ProjectKey)
		}
	}
	return out
}

func isMapped(st Settings, project, repo string) bool {
	return slices.Contains(mappedProjects(st, repo), project)
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return hex.EncodeToString(b)
}
