package plugin

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/kandev/kandev/pkg/pluginsdk"
	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
)

var _ pluginsdk.EntityReferenceHandler = (*Runtime)(nil)

func search(t *testing.T, r *u3rig, source, query string) []pluginsdk.EntityReferenceCandidate {
	t.Helper()
	resp, err := r.rt.SearchEntityReferences(context.Background(), &pluginsdk.SearchEntityReferencesRequest{
		Source: source, WorkspaceID: "ws-1", Query: query, Limit: 5})
	require.NoError(t, err, "a search never fails the composer")
	return resp.Candidates
}

func TestU3_References_SearchReturnsCandidates(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		got := search(t, r, referenceSource, "PROJ-118")
		require.Equal(t, pluginsdk.EntityReferenceCandidate{ProviderLocalID: "PROJ-118", Title: "Fix login timeout",
			URL: "https://" + spaceHost + "/view/PROJ-118"}, got[0], "AC3.4.1")
	})
}

func TestU3_References_SearchEmptyOtherwise(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		require.Empty(t, search(t, r, "other-source", "PROJ-118"), "an unknown source")
		r.call(t, actionSetEnabled, map[string]any{"enabled": false})
		require.Empty(t, search(t, r, referenceSource, "PROJ-118"), "switch off")
		r.call(t, actionSetEnabled, map[string]any{"enabled": true})
		r.call(t, keyDisconnect, nil)
		require.Empty(t, search(t, r, referenceSource, "PROJ-118"), "not connected")
		require.Zero(t, r.gw3.count("issue")+r.gw3.count("issues"))
	})
}

func authorize(t *testing.T, r *u3rig, source string, ref map[string]any) *pluginsdk.AuthorizeEntityReferenceResponse {
	t.Helper()
	resp, err := r.rt.AuthorizeEntityReference(context.Background(), &pluginsdk.AuthorizeEntityReferenceRequest{
		Source: source, WorkspaceID: "ws-1", Purpose: "submission", Reference: ref})
	require.NoError(t, err)
	return resp
}

func issueRef(key string) map[string]any {
	return map[string]any{"version": 1, "ref": "", "provider": referenceProvider, "kind": referenceKind,
		"id": key, "key": key, "title": "t", "url": "", "scope": "ws-1"}
}

func TestU3_References_AuthorizeLiveIssueOnly(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newU3Rig(t)
		require.True(t, authorize(t, r, referenceSource, issueRef("PROJ-118")).Allowed)
		r.gw3.mu.Lock()
		require.Equal(t, []backlog.CallClass{backlog.Interactive}, r.gw3.classes)
		r.gw3.mu.Unlock()
	})
}

func TestU3_References_AuthorizeFailsClosed(t *testing.T) {
	other := issueRef("PROJ-118")
	other["provider"] = "github"
	notIssue := issueRef("PROJ-118")
	notIssue["kind"] = "pull_request"
	noKey := issueRef("PROJ-118")
	delete(noKey, "key")
	cases := map[string]struct {
		source string
		ref    map[string]any
		setup  func(r *u3rig)
	}{
		"unknown issue":  {referenceSource, issueRef("PROJ-999"), func(*u3rig) {}},
		"other source":   {"x", issueRef("PROJ-118"), func(*u3rig) {}},
		"other provider": {referenceSource, other, func(*u3rig) {}},
		"other kind":     {referenceSource, notIssue, func(*u3rig) {}},
		"no key":         {referenceSource, noKey, func(*u3rig) {}},
		"rate limited": {referenceSource, issueRef("PROJ-118"), func(r *u3rig) {
			r.gw3.getErr = &backlog.Error{Kind: backlog.KindRateLimited, Status: 429, RetryAfter: time.Second}
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r := newU3Rig(t)
				tc.setup(r)
				resp := authorize(t, r, tc.source, tc.ref)
				require.False(t, resp.Allowed)
				require.Equal(t, referenceDenied, resp.Reason, "a fixed reason, never Backlog content")
			})
		})
	}
}
