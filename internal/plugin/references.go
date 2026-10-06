package plugin

import (
	"context"

	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/redact"
)

// The `#` issue source declared under reference_sources in manifest.yaml (M10).
const (
	referenceSource   = "nulab-backlog-issues"
	referenceProvider = "nulab-backlog"
	referenceKind     = "issue"
	// referenceDenied is the one reason of every denial: never Backlog content.
	referenceDenied = "The Backlog issue is not available."
)

var _ pluginsdk.EntityReferenceHandler = (*Runtime)(nil)

// SearchEntityReferences answers Kandev's `#` issue search. Results are
// display-only; anything that goes wrong is an empty list (AC3.4.3).
func (r *Runtime) SearchEntityReferences(ctx context.Context, req *pluginsdk.SearchEntityReferencesRequest) (*pluginsdk.SearchEntityReferencesResponse, error) {
	out := &pluginsdk.SearchEntityReferencesResponse{Candidates: []pluginsdk.EntityReferenceCandidate{}}
	if req.Source != referenceSource {
		return out, nil
	}
	ctx = redact.WithLogger(ctx, r.log.With("workspaceId", req.WorkspaceID))
	found, _ := r.issues.Suggest(ctx, req.WorkspaceID, req.Query, int(req.Limit))
	for _, c := range found {
		out.Candidates = append(out.Candidates, pluginsdk.EntityReferenceCandidate{ProviderLocalID: c.Key, Title: c.Title, URL: c.URL})
	}
	return out, nil
}

// AuthorizeEntityReference allows a reference only for this source,
// provider and kind and an issue Backlog returns live; every other case is
// a denial with a fixed reason (fail closed).
func (r *Runtime) AuthorizeEntityReference(ctx context.Context, req *pluginsdk.AuthorizeEntityReferenceRequest) (*pluginsdk.AuthorizeEntityReferenceResponse, error) {
	deny := &pluginsdk.AuthorizeEntityReferenceResponse{Reason: referenceDenied}
	ref := req.Reference
	key, _ := ref["key"].(string)
	if req.Source != referenceSource || ref["provider"] != referenceProvider || ref["kind"] != referenceKind || key == "" {
		return deny, nil
	}
	ctx = redact.WithLogger(ctx, r.log.With("workspaceId", req.WorkspaceID))
	if !r.issues.Authorize(ctx, req.WorkspaceID, key) {
		return deny, nil
	}
	return &pluginsdk.AuthorizeEntityReferenceResponse{Allowed: true}, nil
}
