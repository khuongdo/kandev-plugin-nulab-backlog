package connection

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// refreshBefore is how long before expiry an OAuth token is refreshed.
const refreshBefore = 5 * time.Minute

// Snapshot is the non-secret part of a connection (contract C3).
type Snapshot struct {
	SpaceHost        string
	AuthMethod       string
	ConnectionEpoch  int
	SelectedProjects []string
}

// Current returns the workspace's connection, or ErrNotConnected.
func (s *Service) Current(ctx context.Context, workspaceID string) (Snapshot, error) {
	p, err := s.usable(ctx, workspaceID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{SpaceHost: p.rec.SpaceHost, AuthMethod: p.rec.AuthMethod,
		ConnectionEpoch: p.rec.ConnectionEpoch, SelectedProjects: p.rec.SelectedProjects}, nil
}

// Credentials returns the credentials to call Backlog with and their epoch,
// after BR2.11. An OAuth token with less than 5 minutes left is refreshed
// first, once per workspace however many callers wait (AC1.4.1, AC1.4.2).
func (s *Service) Credentials(ctx context.Context, workspaceID string) (backlog.Credentials, int, error) {
	p, err := s.usable(ctx, workspaceID)
	if err != nil {
		return backlog.Credentials{}, 0, err
	}
	if p.sec.method() == authOAuth && s.needsRefresh(p.sec) {
		return s.refresh(ctx, workspaceID)
	}
	return credsOf(p), p.rec.ConnectionEpoch, nil
}

// usable reads the pair and refuses one that is not connected, or whose
// OAuth sign-in was refused (0 Backlog calls until a new sign-in, AC1.4.3).
func (s *Service) usable(ctx context.Context, workspaceID string) (pair, error) {
	p, err := s.store.readPair(ctx, workspaceID)
	switch {
	case err != nil:
		return pair{}, err
	case !p.connected():
		return pair{}, ErrNotConnected
	case p.rec.SignInAgain:
		return pair{}, ErrReconnectRequired
	}
	return p, nil
}

func credsOf(p pair) backlog.Credentials {
	if p.sec.method() == authOAuth {
		return backlog.Credentials{SpaceHost: p.rec.SpaceHost, AccessToken: p.sec.AccessToken}
	}
	return backlog.Credentials{SpaceHost: p.rec.SpaceHost, APIKey: p.sec.APIKey}
}

func (s *Service) needsRefresh(sec secret) bool {
	exp, err := time.Parse(time.RFC3339, sec.ExpiresAt)
	return err != nil || exp.Sub(s.store.Now()) < refreshBefore
}

// refresh runs under the workspace lock and re-reads the secret after taking
// it, so the callers that waited reuse the first caller's new token.
func (s *Service) refresh(ctx context.Context, workspaceID string) (backlog.Credentials, int, error) {
	unlock, err := s.lockWS(ctx, workspaceID)
	if err != nil {
		return backlog.Credentials{}, 0, err
	}
	defer unlock()
	p, err := s.usable(ctx, workspaceID)
	if err != nil {
		return backlog.Credentials{}, 0, err
	}
	epoch := p.rec.ConnectionEpoch
	if !s.needsRefresh(p.sec) {
		return credsOf(p), epoch, nil
	}
	cfg, err := s.oauthConfig(ctx)
	if err != nil {
		return backlog.Credentials{}, 0, err
	}
	ctx = redact.WithSecrets(ctx, p.sec.AccessToken, p.sec.RefreshToken, cfg.ClientSecret)
	tokens, err := s.gateway.RefreshToken(ctx, p.rec.SpaceHost, cfg.Client(), p.sec.RefreshToken)
	if err != nil {
		return backlog.Credentials{}, 0, s.refreshFailed(ctx, workspaceID, epoch, err)
	}
	ctx = redact.WithSecrets(ctx, tokens.AccessToken, tokens.RefreshToken)
	if err := s.store.UpdateTokens(ctx, workspaceID, epoch, tokens); err != nil {
		if errors.Is(err, ErrStale) {
			return backlog.Credentials{}, 0, ErrReconnectRequired // the connection changed meanwhile (R-08)
		}
		return backlog.Credentials{}, 0, err
	}
	redact.Logger(ctx).InfoContext(ctx, "token refreshed", "event", "token_refreshed", "connectionEpoch", epoch)
	return backlog.Credentials{SpaceHost: p.rec.SpaceHost, AccessToken: tokens.AccessToken}, epoch, nil
}

// refreshFailed asks for a new sign-in when Backlog refused the refresh
// token (400/401). Other failures are temporary and change nothing.
func (s *Service) refreshFailed(ctx context.Context, workspaceID string, epoch int, err error) error {
	var be *backlog.Error
	refused := errors.As(err, &be) && (be.Kind == backlog.KindInvalid || be.Kind == backlog.KindUnauthorized)
	redact.Logger(ctx).WarnContext(ctx, "token refresh failed", "event", "token_refresh_failed",
		"errorCode", Classify(err).Code, "signInAgain", refused, "connectionEpoch", epoch)
	if !refused {
		return err
	}
	if markErr := s.store.MarkSignInAgain(ctx, workspaceID, epoch); markErr != nil && !errors.Is(markErr, ErrStale) {
		return markErr
	}
	return fmt.Errorf("%w: %w", ErrReconnectRequired, err)
}

// oauthConfig reads the OAuth app on every use: Kandev may change it.
func (s *Service) oauthConfig(ctx context.Context) (OAuthConfig, error) {
	if s.Config == nil {
		return OAuthConfig{}, ErrOAuthNotConfigured
	}
	m, err := s.Config.GetConfig(ctx)
	if err != nil {
		return OAuthConfig{}, s.store.storeErr(ctx, "read plugin config", err)
	}
	return ParseOAuthConfig(m)
}

// reconnectError turns Backlog refusing the stored credentials into
// ErrReconnectRequired (R-06: 401 and 403 on U2 calls).
func reconnectError(err error) error {
	var be *backlog.Error
	if errors.As(err, &be) && (be.Kind == backlog.KindUnauthorized || be.Kind == backlog.KindForbidden) {
		return fmt.Errorf("%w: %w", ErrReconnectRequired, err)
	}
	return err
}

// Test re-checks the stored credentials with Myself and returns the view
// with the fresh user name. It stores nothing (AC1.5.1, AC1.5.2).
func (s *Service) Test(ctx context.Context, workspaceID string) (View, error) {
	start := time.Now()
	user, creds, err := s.test(ctx, workspaceID)
	attrs := []any{"event", "connection_tested", "durationMs", time.Since(start).Milliseconds()}
	if err != nil {
		attrs = append(attrs, "errorCode", Classify(err).Code)
	}
	redact.Logger(ctx).InfoContext(ctx, "connection tested", attrs...)
	if err != nil {
		return View{}, err
	}
	view, err := s.store.Load(ctx, workspaceID)
	view.ConnectedUserName = user.Name
	if err == nil && view.HasGitCredential {
		view.GitCheck = s.gitCheck(ctx, workspaceID, creds, view.SelectedProjects)
	}
	return view, err
}

func (s *Service) test(ctx context.Context, workspaceID string) (backlog.User, backlog.Credentials, error) {
	creds, _, err := s.Credentials(ctx, workspaceID)
	if err != nil {
		return backlog.User{}, creds, err
	}
	ctx = redact.WithSecrets(ctx, creds.APIKey, creds.AccessToken)
	user, err := s.gateway.Myself(ctx, creds)
	return user, creds, reconnectError(err)
}

// Disconnect deletes the credentials and any pending sign-in, keeps a
// disconnect record and sends one disconnected event (AC1.5.4).
func (s *Service) Disconnect(ctx context.Context, workspaceID string) (View, error) {
	unlock, err := s.lockWS(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	defer unlock()
	if err := s.store.DeletePending(ctx, workspaceID); err != nil {
		return View{}, err
	}
	view, changed, err := s.store.Disconnect(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	if changed {
		s.publish(ctx, ConnectionChanged{WorkspaceID: workspaceID, Reason: ReasonDisconnected, ConnectionEpoch: view.ConnectionEpoch})
	}
	return s.store.Load(ctx, workspaceID)
}

// ProjectItem is one row of connection.list_projects.
type ProjectItem struct {
	ProjectKey  string `json:"projectKey"`
	ProjectID   int64  `json:"projectId"`
	ProjectName string `json:"projectName"`
	Selected    bool   `json:"selected"`
}

// ListProjects returns the space's projects with the current selection.
func (s *Service) ListProjects(ctx context.Context, workspaceID string) ([]ProjectItem, error) {
	snap, err := s.Current(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	projects, err := s.projects(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	items := make([]ProjectItem, 0, len(projects))
	for _, p := range projects {
		items = append(items, ProjectItem{ProjectKey: p.Key, ProjectID: p.ID, ProjectName: p.Name,
			Selected: slices.Contains(snap.SelectedProjects, p.Key)})
	}
	return items, nil
}

func (s *Service) projects(ctx context.Context, workspaceID string) ([]backlog.Project, error) {
	creds, _, err := s.Credentials(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	ctx = redact.WithSecrets(ctx, creds.APIKey, creds.AccessToken)
	projects, err := s.gateway.Projects(ctx, creds)
	return projects, reconnectError(err)
}

// SetProjects stores a project selection made of keys on the space and
// sends projects_changed; an unchanged selection writes and sends nothing.
func (s *Service) SetProjects(ctx context.Context, workspaceID string, raw any) (View, error) {
	keys, err := ValidateProjectKeys(raw)
	if err != nil {
		return View{}, err
	}
	projects, err := s.projects(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	for _, k := range keys {
		if !slices.ContainsFunc(projects, func(p backlog.Project) bool { return p.Key == k }) {
			return View{}, fmt.Errorf("project %s is not on the space: %w", k, ErrInvalidProjects)
		}
	}
	unlock, err := s.lockWS(ctx, workspaceID)
	if err != nil {
		return View{}, err
	}
	defer unlock()
	prev, next, view, err := s.store.saveProjects(ctx, workspaceID, keys)
	if err != nil {
		return View{}, err
	}
	if next.ConnectionEpoch != prev.ConnectionEpoch {
		s.emit(ctx, workspaceID, prev, next)
	}
	view.Enabled = true // a guarded action only runs while Backlog is on
	return view, nil
}
