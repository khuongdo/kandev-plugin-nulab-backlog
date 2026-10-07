package issues

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validWatchInput() IssueWatchInput {
	return IssueWatchInput{Name: "Open bugs", ProjectKey: "PROJ", StatusIDs: []int64{1, 2}, Assignee: "anyone",
		Creator: "me", WorkflowID: "wf-1", WorkflowStepID: "step-1"}
}

func minutes(f float64) *float64 { return &f }

// BR3.1, BR3.2: every field is checked; the first failing field is reported.
func TestIssueWatch_ValidateRejectsEachBadField(t *testing.T) {
	selected := []string{"PROJ", "DEMO"}
	twenty := make([]int64, 20)
	for i := range twenty {
		twenty[i] = int64(i + 1)
	}
	cases := []struct {
		name  string
		mut   func(*IssueWatchInput)
		field string
	}{
		{"name empty", func(in *IssueWatchInput) { in.Name = "  " }, FieldName},
		{"name 101 chars", func(in *IssueWatchInput) { in.Name = strings.Repeat("あ", 101) }, FieldName},
		{"project not selected", func(in *IssueWatchInput) { in.ProjectKey = "OTHER" }, FieldProjectKey},
		{"no status", func(in *IssueWatchInput) { in.StatusIDs = nil }, FieldStatusIDs},
		{"21 statuses", func(in *IssueWatchInput) { in.StatusIDs = append(twenty, 21) }, FieldStatusIDs},
		{"status id 0", func(in *IssueWatchInput) { in.StatusIDs = []int64{0} }, FieldStatusIDs},
		{"assignee other", func(in *IssueWatchInput) { in.Assignee = "lan" }, FieldAssignee},
		{"creator other", func(in *IssueWatchInput) { in.Creator = "lan" }, FieldCreator},
		{"no workflow", func(in *IssueWatchInput) { in.WorkflowID = " " }, FieldWorkflow},
		{"interval 0", func(in *IssueWatchInput) { in.IntervalMinutes = minutes(0) }, FieldInterval},
		{"interval 1441", func(in *IssueWatchInput) { in.IntervalMinutes = minutes(1441) }, FieldInterval},
		{"interval not whole", func(in *IssueWatchInput) { in.IntervalMinutes = minutes(2.5) }, FieldInterval},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validWatchInput()
			tc.mut(&in)
			require.Equal(t, tc.field, fieldOf(t, in.Validate(selected)))
		})
	}
}

func TestIssueWatch_ValidateAcceptsTheBoundsAndFillsDefaults(t *testing.T) {
	selected := []string{"PROJ"}
	twenty := make([]int64, 20)
	for i := range twenty {
		twenty[i] = int64(i + 1)
	}
	cases := []struct {
		name string
		mut  func(*IssueWatchInput)
		want func(*testing.T, IssueWatchInput)
	}{
		{"name of 1 char", func(in *IssueWatchInput) { in.Name = " x " }, func(t *testing.T, in IssueWatchInput) {
			require.Equal(t, "x", in.Name, "trimmed")
		}},
		{"name of 100 chars", func(in *IssueWatchInput) { in.Name = strings.Repeat("あ", 100) }, nil},
		{"one status", func(in *IssueWatchInput) { in.StatusIDs = []int64{3} }, nil},
		{"twenty statuses", func(in *IssueWatchInput) { in.StatusIDs = twenty }, nil},
		{"duplicate statuses are removed", func(in *IssueWatchInput) { in.StatusIDs = []int64{2, 1, 2, 1} },
			func(t *testing.T, in IssueWatchInput) { require.Equal(t, []int64{2, 1}, in.StatusIDs) }},
		{"interval 1", func(in *IssueWatchInput) { in.IntervalMinutes = minutes(1) }, nil},
		{"interval 1440", func(in *IssueWatchInput) { in.IntervalMinutes = minutes(1440) }, nil},
		{"BR3.2 interval absent is 5", func(in *IssueWatchInput) { in.IntervalMinutes = nil },
			func(t *testing.T, in IssueWatchInput) { require.Equal(t, 5.0, *in.IntervalMinutes) }},
		{"assignee and creator absent are anyone", func(in *IssueWatchInput) { in.Assignee, in.Creator = "", "" },
			func(t *testing.T, in IssueWatchInput) {
				require.Equal(t, "anyone", in.Assignee)
				require.Equal(t, "anyone", in.Creator)
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validWatchInput()
			tc.mut(&in)
			require.NoError(t, in.Validate(selected))
			if tc.want != nil {
				tc.want(t, in)
			}
		})
	}
}

// BR3.5: issues are ordered by (created, id); only those after the cursor count.
func TestIssueCursor_AfterOrdersByCreatedThenID(t *testing.T) {
	at := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		require.NoError(t, err)
		return v
	}
	c := &IssueCursor{Created: at("2026-10-01T09:00:00Z"), IssueID: 50}
	cases := []struct {
		name    string
		created string
		id      int64
		after   bool
	}{
		{"earlier time", "2026-10-01T08:59:59Z", 99, false},
		{"same time, lower id", "2026-10-01T09:00:00Z", 49, false},
		{"the cursor itself", "2026-10-01T09:00:00Z", 50, false},
		{"same time, higher id", "2026-10-01T09:00:00Z", 51, true},
		{"later time, lower id", "2026-10-01T09:00:01Z", 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.after, c.After(at(tc.created), tc.id))
		})
	}
	var none *IssueCursor
	require.True(t, none.After(at("2000-01-01T00:00:00Z"), 1), "without a cursor every issue is new")
}

// R-10: createdSince is the cursor's UTC date minus one day, so a cursor just
// after midnight UTC still sees later issues of the same UTC day.
func TestIssueCursor_CreatedSinceIsTheUTCDayBefore(t *testing.T) {
	c := &IssueCursor{Created: time.Date(2026, 10, 7, 0, 30, 0, 0, time.UTC)}
	require.Equal(t, "2026-10-06", c.CreatedSince())
	tokyo := time.FixedZone("JST", 9*3600)
	c = &IssueCursor{Created: time.Date(2026, 10, 7, 8, 0, 0, 0, tokyo)} // 2026-10-06T23:00Z
	require.Equal(t, "2026-10-05", c.CreatedSince())
	var none *IssueCursor
	require.Empty(t, none.CreatedSince(), "no cursor: no createdSince")
}
