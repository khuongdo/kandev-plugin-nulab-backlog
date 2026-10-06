package git

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/backlog"
	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/testutil"
)

func TestU4_Leak_NoSecretInRepliesErrorsOrTasks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newRig(t)
		bait := testutil.Token(t)
		list := prs(3)
		for i := range list {
			list[i].Description = "body " + bait
		}
		r.gw.setPRs("PROJ/web-app", list)
		var texts []string
		record := func(v any, err error) {
			b, _ := json.Marshal(v)
			texts = append(texts, string(b), fmt.Sprintf("%+v", v))
			if err != nil {
				texts = append(texts, err.Error())
			}
		}
		record(r.svc.ListRepositories(r.ctx, ws, "", ""))
		record(r.svc.Link(r.ctx, ws, "task-17", "3", "repo-k1"))
		record(r.svc.Status(r.ctx, ws, "task-17"))
		record(r.svc.CreatePR(r.ctx, ws, createInput()))
		record(r.svc.ResolveCredential(r.ctx, Scope{ProviderID: "github"}))
		q := saveQuery(t, r, nil)
		record(r.svc.RunQuery(r.ctx, ws, q.ID))
		r.gw.setErr("PullRequest", &backlog.Error{Kind: backlog.KindUnauthorized, Status: 401})
		record(r.svc.Link(r.ctx, ws, "task-17", "2", "repo-k1"))
		r.saveWatch(t, nil)
		watcher := r.startWatcher(t)
		cycle(watcher)
		for _, task := range r.host.tasks {
			texts = append(texts, fmt.Sprintf("%+v", task))
		}
		all := strings.Join(texts, "\n") + r.logs.String()
		for _, secret := range []string{r.conn.creds.APIKey, r.conn.git.Password, bait} {
			testutil.AssertNoLeak(t, all, secret, 8)
		}
	})
}

func TestU4_Leak_LeaseErrorsHoldNoPassword(t *testing.T) {
	r := newRig(t)
	for _, sc := range []Scope{{}, {ProviderID: ProviderID}, {ProviderID: ProviderID, TaskID: "t", SessionID: "s", RepositoryID: "r", Host: "evil.example.com"}} {
		_, err := r.svc.ResolveCredential(r.ctx, sc)
		require.Error(t, err)
		testutil.AssertNoLeak(t, err.Error()+fmt.Sprintf("%+v", err), r.conn.git.Password, 8)
	}
}
