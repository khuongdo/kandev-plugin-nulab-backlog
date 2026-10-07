package issues

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/connection"
)

// Issue watch input fields reported with the validation code (BR3.1).
const (
	FieldName       = "name"
	FieldProjectKey = "projectKey"
	FieldAssignee   = "assignee"
	FieldCreator    = "creator"
	FieldInterval   = "intervalMinutes"
	FieldID         = "id"
)

// Who values of the assignee and creator filters.
const (
	WhoAnyone = "anyone"
	WhoMe     = "me"
)

// Issue watch states (SM1); StateActive and StateNotConnected are shared with links.
const StatePaused = "paused"

// Issue watch lastError values (BR3.11, BR3.13, BR3.14).
const (
	ErrorUnauthorized    = "unauthorized"
	ErrorRateLimited     = "rate_limited"
	ErrorUnavailable     = "unavailable"
	ErrorWorkflowMissing = "workflow_missing"
	ErrorLedgerFull      = "ledger_full"
)

// Ledger outcomes (SM2).
const (
	OutcomeReserved      = "reserved"
	OutcomeCreated       = "created"
	OutcomeSkippedLinked = "skipped_linked"
)

const (
	maxWatchName       = 100
	maxWatchStatuses   = 20
	defaultWatchMinute = 5
	maxWatchMinutes    = 1440
)

// IssueWatchInput is the issues.watches.save body.
type IssueWatchInput struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name"`
	ProjectKey      string   `json:"projectKey"`
	StatusIDs       []int64  `json:"statusIds"`
	Assignee        string   `json:"assignee"`
	Creator         string   `json:"creator"`
	WorkflowID      string   `json:"workflowId"`
	WorkflowStepID  string   `json:"workflowStepId,omitempty"`
	IntervalMinutes *float64 `json:"intervalMinutes,omitempty"` // absent = 5 (BR3.2)
}

// Validate checks every field against the selected projects (BR3.1), trims
// the name, removes duplicate statuses and fills the defaults: interval 5,
// assignee and creator anyone (BR3.2). Workflow existence is checked at run
// time (BR3.14): the plugin cannot read workflows.
func (in *IssueWatchInput) Validate(selected []string) error {
	in.Name = strings.TrimSpace(in.Name)
	if n := utf8.RuneCountInString(in.Name); n == 0 || n > maxWatchName {
		return invalid(FieldName)
	}
	if !slices.Contains(selected, in.ProjectKey) {
		return invalid(FieldProjectKey)
	}
	var ids []int64
	for _, id := range in.StatusIDs {
		if id <= 0 {
			return invalid(FieldStatusIDs)
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 || len(ids) > maxWatchStatuses {
		return invalid(FieldStatusIDs)
	}
	in.StatusIDs = ids
	// ponytail: anyone|me is also checked in internal/git; issues may not import git, so it stays local.
	for _, f := range []struct {
		field string
		who   *string
	}{{FieldAssignee, &in.Assignee}, {FieldCreator, &in.Creator}} {
		if *f.who == "" {
			*f.who = WhoAnyone
		}
		if *f.who != WhoAnyone && *f.who != WhoMe {
			return invalid(f.field)
		}
	}
	if strings.TrimSpace(in.WorkflowID) == "" {
		return invalid(FieldWorkflow)
	}
	if in.IntervalMinutes == nil {
		d := float64(defaultWatchMinute)
		in.IntervalMinutes = &d
	}
	if m := *in.IntervalMinutes; m != float64(int(m)) || m < 1 || m > maxWatchMinutes {
		return invalid(FieldInterval)
	}
	return nil
}

// IssueCursor is the last issue a watch passed, in (created, id) order,
// with its position in the list for its createdSince, so a run resumes near
// it (BR3.5, BR3.12).
type IssueCursor struct {
	Created   time.Time `json:"created"`
	IssueID   int64     `json:"issueId"`
	DayOffset int       `json:"dayOffset"`
}

// After reports whether an issue created at created with id comes after the
// cursor. Every issue is after a nil cursor.
func (c *IssueCursor) After(created time.Time, id int64) bool {
	if c == nil {
		return true
	}
	if !created.Equal(c.Created) {
		return created.After(c.Created)
	}
	return id > c.IssueID
}

// CreatedSince is the createdSince date of a run: the cursor's UTC date
// minus one day, so a timezone difference never hides a later issue of the
// same day (R-10); the (created, id) filter drops the extra day. Empty
// without a cursor.
func (c *IssueCursor) CreatedSince() string {
	if c == nil {
		return ""
	}
	return c.Created.UTC().AddDate(0, 0, -1).Format(time.DateOnly)
}

// IssueWatch is a saved issue watch (FR3).
type IssueWatch struct {
	ID                    string       `json:"id"`
	Name                  string       `json:"name"`
	SpaceHost             string       `json:"spaceHost"`
	ProjectKey            string       `json:"projectKey"`
	StatusIDs             []int64      `json:"statusIds"`
	Assignee              string       `json:"assignee"`
	Creator               string       `json:"creator"`
	AssigneeID            int64        `json:"assigneeId,omitempty"`
	CreatedUserID         int64        `json:"createdUserId,omitempty"`
	WorkflowID            string       `json:"workflowId"`
	WorkflowStepID        string       `json:"workflowStepId,omitempty"`
	IntervalMinutes       int          `json:"intervalMinutes"`
	State                 string       `json:"state"`
	StateBeforeDisconnect string       `json:"stateBeforeDisconnect,omitempty"`
	Cursor                *IssueCursor `json:"cursor,omitempty"`
	CreatedCount          int          `json:"createdCount"`
	PendingCount          int          `json:"pendingCount"`
	LastRunAt             string       `json:"lastRunAt,omitempty"`
	LastError             string       `json:"lastError,omitempty"`
}

// IssueWatchLedgerEntry records that a watch reserved, created or skipped
// the task of one issue, so the issue is handled at most once (FR3.4).
type IssueWatchLedgerEntry struct {
	Key     string `json:"key"` // spaceHost|issueId
	WatchID string `json:"watchId"`
	Outcome string `json:"outcome"`
	TaskID  string `json:"taskId,omitempty"`
	At      string `json:"at"`
}

var errDuplicateName = errors.New("another watch has this name")

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return hex.EncodeToString(b)
}

// ListWatches returns the workspace's issue watches, never nil.
func (s *Service) ListWatches(ctx context.Context, ws string) ([]IssueWatch, error) {
	list, err := s.store.Watches(ctx, ws)
	if list == nil && err == nil {
		list = []IssueWatch{}
	}
	return list, err
}

// SaveWatch creates (no id) or edits an issue watch (WF2). "me" becomes the
// connected user's id with one Myself call. An edit keeps the state, counts
// and ledger; a changed project, status, assignee or creator clears the
// cursor (BR3.8), and any edit clears the last error.
func (s *Service) SaveWatch(ctx context.Context, ws string, in IssueWatchInput) (IssueWatch, error) {
	snap, err := s.conn.Current(ctx, ws)
	if err != nil {
		return IssueWatch{}, err
	}
	if err := in.Validate(snap.SelectedProjects); err != nil {
		return IssueWatch{}, err
	}
	w := IssueWatch{ID: in.ID, Name: in.Name, SpaceHost: snap.SpaceHost, ProjectKey: in.ProjectKey, StatusIDs: in.StatusIDs,
		Assignee: in.Assignee, Creator: in.Creator, WorkflowID: strings.TrimSpace(in.WorkflowID), WorkflowStepID: in.WorkflowStepID,
		IntervalMinutes: int(*in.IntervalMinutes), State: StateActive}
	if in.Assignee == WhoMe || in.Creator == WhoMe {
		cctx, creds, err := s.credentials(ctx, ws)
		if err != nil {
			return IssueWatch{}, err
		}
		me, err := s.gateway.Myself(cctx, creds)
		if err != nil {
			return IssueWatch{}, err
		}
		if in.Assignee == WhoMe {
			w.AssigneeID = me.ID
		}
		if in.Creator == WhoMe {
			w.CreatedUserID = me.ID
		}
	}
	err = s.store.UpdateWatches(ctx, ws, func(list []IssueWatch) ([]IssueWatch, error) {
		if slices.ContainsFunc(list, func(x IssueWatch) bool { return x.ID != w.ID && strings.EqualFold(x.Name, w.Name) }) {
			return nil, &connection.FieldError{Field: FieldName, Err: errDuplicateName}
		}
		if w.ID == "" {
			w.ID = newID()
			return append(list, w), nil
		}
		i := slices.IndexFunc(list, func(x IssueWatch) bool { return x.ID == w.ID })
		if i < 0 {
			return nil, ErrNotFound
		}
		old := list[i]
		w.State, w.StateBeforeDisconnect = old.State, old.StateBeforeDisconnect
		w.CreatedCount, w.PendingCount, w.LastRunAt = old.CreatedCount, old.PendingCount, old.LastRunAt
		if sameFilters(old, w) {
			w.Cursor = old.Cursor
		}
		list[i] = w
		return list, nil
	})
	return w, err
}

func sameFilters(a, b IssueWatch) bool {
	return a.ProjectKey == b.ProjectKey && slices.Equal(a.StatusIDs, b.StatusIDs) && a.Assignee == b.Assignee &&
		a.Creator == b.Creator && a.AssigneeID == b.AssigneeID && a.CreatedUserID == b.CreatedUserID
}

// DeleteWatch removes a watch and its ledger; its tasks and links stay (BR3.10).
func (s *Service) DeleteWatch(ctx context.Context, ws, id string) error {
	err := s.store.UpdateWatches(ctx, ws, func(list []IssueWatch) ([]IssueWatch, error) {
		n := len(list)
		list = slices.DeleteFunc(list, func(w IssueWatch) bool { return w.ID == id })
		if len(list) == n {
			return nil, ErrNotFound
		}
		return list, nil
	})
	if err != nil {
		return err
	}
	return s.store.DeleteLedger(ctx, ws, id)
}

// PauseWatch stops a watch from running (BR3.9).
func (s *Service) PauseWatch(ctx context.Context, ws, id string) (IssueWatch, error) {
	return s.setWatchState(ctx, ws, id, StatePaused)
}

// ResumeWatch lets a paused watch run again; a not-connected one is a conflict.
func (s *Service) ResumeWatch(ctx context.Context, ws, id string) (IssueWatch, error) {
	return s.setWatchState(ctx, ws, id, StateActive)
}

func (s *Service) setWatchState(ctx context.Context, ws, id, state string) (IssueWatch, error) {
	var out IssueWatch
	err := s.store.UpdateWatches(ctx, ws, func(list []IssueWatch) ([]IssueWatch, error) {
		i := slices.IndexFunc(list, func(w IssueWatch) bool { return w.ID == id })
		switch {
		case i < 0:
			return nil, ErrNotFound
		case list[i].State == StateNotConnected:
			return nil, ErrConflict
		}
		list[i].State = state
		out = list[i]
		return list, nil
	})
	return out, err
}

func (s *Service) findWatch(ctx context.Context, ws, id string) (IssueWatch, error) {
	list, err := s.store.Watches(ctx, ws)
	if err != nil {
		return IssueWatch{}, err
	}
	i := slices.IndexFunc(list, func(w IssueWatch) bool { return w.ID == id })
	if i < 0 {
		return IssueWatch{}, ErrNotFound
	}
	return list[i], nil
}

// applyWatches turns watches t does not cover (all when t is nil) to
// not_connected, remembering their state, and on restore brings covered
// ones back to the state they had (SM1).
func (s *Service) applyWatches(ctx context.Context, ws string, t *target, restore bool) error {
	return s.store.UpdateWatches(ctx, ws, func(list []IssueWatch) ([]IssueWatch, error) {
		changed := false
		for i, w := range list {
			covered := t != nil && w.SpaceHost == t.host && slices.Contains(t.selected, w.ProjectKey)
			switch {
			case !covered && w.State != StateNotConnected:
				list[i].StateBeforeDisconnect, list[i].State = w.State, StateNotConnected
			case covered && restore && w.State == StateNotConnected:
				list[i].State = cmp.Or(w.StateBeforeDisconnect, StateActive)
				list[i].StateBeforeDisconnect = ""
			default:
				continue
			}
			changed = true
		}
		if !changed {
			return nil, errUnchanged
		}
		return list, nil
	})
}
