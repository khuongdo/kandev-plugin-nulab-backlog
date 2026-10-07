import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { gitNotice, noticeText, watchStatusKey } from "../git/git-state";
import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { createIssueWatchDialog, type FilterOption, type IssueWatch } from "./issue-watch-dialog";
import { createSectionParts } from "./section-parts";
import type { Notice } from "./state";
import { useActionList } from "./use-list";

const LAST_ERRORS: Record<string, MessageKey> = {
  unauthorized: "lastErrorUnauthorized",
  rate_limited: "lastErrorRateLimited",
  unavailable: "lastErrorUnavailable",
  workflow_missing: "lastErrorWorkflowMissing",
  ledger_full: "lastErrorLedgerFull",
};

interface Filters {
  projects?: FilterOption[];
  statuses?: FilterOption[];
}

/**
 * Issue watches in Settings > Integrations > Backlog (FR1.3, FR3): the same
 * table, header action and dialog as PR watches, plus the interval and the
 * last run error (BR3.11-BR3.14).
 */
export function createIssueWatchesSection(
  host: PluginHostApi,
  messages: Messages = en,
): Component<{ workspaceId: string }> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const ui = hostUi(host);
  const { Button, Badge, Card, CardContent, SettingsSection } = ui;
  const { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } = ui;
  const { ListEmpty, ListError, RowMenu } = createSectionParts(host, messages);
  const ConfirmDialog = createConfirmDialog(host, messages);
  const IssueWatchDialog = createIssueWatchDialog(host, messages);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;

  return function IssueWatchesSection({ workspaceId }: { workspaceId: string }) {
    const list = useActionList<IssueWatch>(host, workspaceId, "issues.watches.list", "watches");
    const [filters, setFilters] = useState<Filters>({});
    const [editing, setEditing] = useState<IssueWatch | "new" | undefined>(undefined);
    const [deleting, setDeleting] = useState<IssueWatch | undefined>(undefined);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);

    useEffect(() => {
      host.api
        .invokeAction<Filters>("issues.filters", { workspaceId })
        .then((f) => setFilters(f ?? {}))
        .catch(() => setFilters({})); // the table still shows keys and ids
    }, [workspaceId]);

    const projectName = (key: string) => filters.projects?.find((p) => p.key === key)?.name ?? key;
    const statusNames = (ids: number[]) =>
      ids.map((id) => filters.statuses?.find((s) => s.id === id)?.name ?? String(id)).join(", ");
    const replace = (w: IssueWatch) =>
      list.setItems((items) =>
        items.some((x) => x.id === w.id) ? items.map((x) => (x.id === w.id ? w : x)) : [...items, w],
      );
    const act = async (key: string, id: string, done: (reply: unknown) => void) => {
      setNotice(undefined);
      try {
        done(await host.api.invokeAction(key, { workspaceId, body: { id } }));
      } catch (e) {
        setNotice(gitNotice(e));
      }
    };

    const add = (testId: string) => (
      <Button
        type="button"
        size="sm"
        className={BUTTON}
        data-testid={testId}
        onClick={() => setEditing("new")}
      >
        {icon(host, "plus", "mr-1 h-4 w-4")}
        {messages.addWatch}
      </Button>
    );

    const rows = list.items.map((w) => (
      <TableRow key={w.id} data-testid={`backlog-issue-watch-row-${w.id}`}>
        <TableCell>{w.name}</TableCell>
        <TableCell>{projectName(w.projectKey)}</TableCell>
        <TableCell>{statusNames(w.statusIds ?? [])}</TableCell>
        <TableCell>{format(messages.intervalValue, { minutes: w.intervalMinutes })}</TableCell>
        <TableCell>
          <Badge variant="outline">{messages[watchStatusKey(w.state)]}</Badge>
          {w.lastError && LAST_ERRORS[w.lastError] ? (
            <Badge variant="destructive" data-testid={`backlog-issue-watch-error-${w.id}`}>
              {messages[LAST_ERRORS[w.lastError]!]}
            </Badge>
          ) : null}
        </TableCell>
        <TableCell>
          {format(messages.progressValue, { created: w.createdCount, pending: w.pendingCount })}
        </TableCell>
        <TableCell>{w.lastRunAt ? relative(w.lastRunAt) : messages.neverRun}</TableCell>
        <TableCell>
          <RowMenu
            testId={`backlog-issue-watch-menu-${w.id}`}
            label={format(messages.rowActions, { name: w.name })}
            items={[
              {
                testId: `backlog-issue-watch-edit-${w.id}`,
                label: messages.edit,
                onSelect: () => setEditing(w),
              },
              ...(w.state === "active"
                ? [
                    {
                      testId: `backlog-issue-watch-run-${w.id}`,
                      label: messages.runNow,
                      onSelect: () =>
                        void act("issues.watches.run", w.id, () => setNotice({ key: "runQueued" })),
                    },
                    {
                      testId: `backlog-issue-watch-pause-${w.id}`,
                      label: messages.pause,
                      onSelect: () => void act("issues.watches.pause", w.id, (r) => replace(r as IssueWatch)),
                    },
                  ]
                : []),
              ...(w.state === "paused"
                ? [
                    {
                      testId: `backlog-issue-watch-resume-${w.id}`,
                      label: messages.resume,
                      onSelect: () =>
                        void act("issues.watches.resume", w.id, (r) => replace(r as IssueWatch)),
                    },
                  ]
                : []),
              {
                testId: `backlog-issue-watch-delete-${w.id}`,
                label: messages.delete,
                onSelect: () => setDeleting(w),
              },
            ]}
          />
        </TableCell>
      </TableRow>
    ));

    const body = list.error ? (
      <ListError testId="backlog-issue-watches" notice={list.error} onRetry={list.reload} />
    ) : !list.loading && list.items.length === 0 ? (
      <ListEmpty testId="backlog-issue-watches" title={messages.watchesEmpty}>
        {add("backlog-issue-watches-empty-add")}
      </ListEmpty>
    ) : (
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{messages.colName}</TableHead>
                <TableHead>{messages.colProject}</TableHead>
                <TableHead>{messages.colStatuses}</TableHead>
                <TableHead>{messages.colInterval}</TableHead>
                <TableHead>{messages.watchStatusLegend}</TableHead>
                <TableHead>{messages.colTasks}</TableHead>
                <TableHead>{messages.colLastRun}</TableHead>
                <TableHead>{messages.colActions}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>{rows}</TableBody>
          </Table>
        </CardContent>
      </Card>
    );

    return (
      <div data-testid="backlog-section-issue-watches">
        <SettingsSection
          title={messages.sectionIssueWatches}
          description={messages.issueWatchesDescription}
          action={add("backlog-issue-watches-add")}
        >
          {body}
          <p role="status" data-testid="backlog-issue-watches-notice">
            {notice ? noticeText(notice, messages) : ""}
          </p>
        </SettingsSection>
        {editing ? (
          <IssueWatchDialog
            workspaceId={workspaceId}
            initial={editing === "new" ? undefined : editing}
            projects={filters.projects ?? []}
            statuses={filters.statuses ?? []}
            onSaved={(w: IssueWatch) => {
              replace(w);
              setEditing(undefined);
            }}
            onClose={() => setEditing(undefined)}
          />
        ) : null}
        {deleting ? (
          <ConfirmDialog
            testId="backlog-issue-watch-delete-dialog"
            title={messages.deleteWatchTitle}
            body={messages.deleteWatchBody}
            confirmLabel={messages.delete}
            destructive
            onConfirm={async () => {
              await host.api.invokeAction("issues.watches.delete", {
                workspaceId,
                body: { id: deleting.id },
              });
              list.setItems((items) => items.filter((x) => x.id !== deleting.id));
            }}
            onClose={() => setDeleting(undefined)}
          />
        ) : null}
      </div>
    );
  };
}
