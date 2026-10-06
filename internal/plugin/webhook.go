package plugin

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// webhookOAuthCallback is the public webhook Backlog redirects back to.
const webhookOAuthCallback = "oauth-callback"

// failedLocation is used when the state names no workspace.
const failedLocation = "/settings/integrations?oauth=failed"

// HandleWebhook serves the OAuth callback. It answers with a relative
// redirect to the settings page carrying only the outcome; the code, tokens,
// state and any Backlog text never reach the Location. It never returns a
// Go error.
func (r *Runtime) HandleWebhook(ctx context.Context, req *pluginsdk.WebhookRequest) (resp *pluginsdk.WebhookResponse, err error) {
	log := r.log.With("requestId", newRequestID())
	ctx = redact.WithLogger(ctx, log)
	defer func() {
		if p := recover(); p != nil {
			log.ErrorContext(ctx, "webhook panicked", "event", "webhook_panic", "webhookKey", req.WebhookKey)
			resp, err = redirect(failedLocation), nil
		}
	}()
	if req.WebhookKey != webhookOAuthCallback {
		return &pluginsdk.WebhookResponse{Status: http.StatusNotFound}, nil
	}
	if !strings.EqualFold(req.Method, http.MethodGet) {
		return &pluginsdk.WebhookResponse{Status: http.StatusMethodNotAllowed, Headers: map[string]string{"Allow": http.MethodGet}}, nil
	}
	res := connection.OAuthResult{Outcome: connection.OutcomeFailed}
	if q, perr := url.ParseQuery(req.Query); perr == nil {
		// The callback gets the same budget as an action (12 s).
		cctx, cancel := context.WithTimeout(ctx, r.service.Deadline)
		defer cancel()
		res = r.service.CompleteOAuth(cctx, q)
	}
	log.InfoContext(ctx, "oauth callback", "event", "oauth_callback", "outcome", res.Outcome)
	return redirect(callbackLocation(res)), nil
}

// callbackLocation is the settings page of the state's workspace.
func callbackLocation(res connection.OAuthResult) string {
	if res.WorkspaceID == "" {
		return failedLocation
	}
	loc := "/settings/workspaces/" + url.PathEscape(res.WorkspaceID) + "/integrations/nulab-backlog?oauth=" + res.Outcome
	if res.Restored {
		loc += "&restored=1"
	}
	return loc
}

func redirect(location string) *pluginsdk.WebhookResponse {
	return &pluginsdk.WebhookResponse{Status: http.StatusFound, Headers: map[string]string{
		"Location": location, "Cache-Control": "no-store",
	}}
}
