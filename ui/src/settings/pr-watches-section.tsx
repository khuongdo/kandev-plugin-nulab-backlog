import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { gitNotice, noticeText, watchStatusKey } from "../git/git-state";
import { createWatchForm, type Watch } from "../git/watch-form";
import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { createSectionParts } from "./section-parts";
import type { Notice } from "./state";
import { useActionList } from "./use-list";

/** The anchor the restore notice scrolls to (BR1.5). */
export const PR_WATCHES_ANCHOR = "backlog-pr-watches";

/**
 * PR watches in Settings > Integrations > Backlog (FR1.2): a table with a
 * row menu (Edit, Run now, Pause/Resume, Delete), Add watch in the header,
 * add and edit in a dialog. Open to every member (FR1.4).
 */
export function createPrWatchesSection(
  host: PluginHostApi,
  messages: Messages = en,
): Component<{ workspaceId: string }> {
  const h = host.jsx;
  const { useState } = host.React;
  const ui = hostUi(host);
  const {
    Button,
    Badge,
    Card,
    CardContent,
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    SettingsSection,
  } = ui;
  const { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } = ui;
  const { ListEmpty, ListError, RowMenu } = createSectionParts(host, messages);
  const ConfirmDialog = createConfirmDialog(host, messages);
  const WatchForm = createWatchForm(host, messages);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;

  return function PrWatchesSection({ workspaceId }: { workspaceId: string }) {
    const list = useActionList<Watch>(host, workspaceId, "git.watches.list", "watches");
    const [editing, setEditing] = useState<Partial<Watch> | undefined>(undefined);
    const [deleting, setDeleting] = useState<Watch | undefined>(undefined);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);

    const replace = (w: Watch) =>
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
      <Button type="button" size="sm" className={BUTTON} data-testid={testId} onClick={() => setEditing({})}>
        {icon(host, "plus", "mr-1 h-4 w-4")}
        {messages.addWatch}
      </Button>
    );

    const rows = list.items.map((w) => (
      <TableRow key={w.id} data-testid={`backlog-pr-watch-row-${w.id}`}>
        <TableCell>{w.name}</TableCell>
        <TableCell>{format(messages.watchRepo, { project: w.projectKey, repo: w.repoName })}</TableCell>
        <TableCell>
          <Badge variant="outline">{messages[watchStatusKey(w.state)]}</Badge>
        </TableCell>
        <TableCell>
          {format(messages.progressValue, { created: w.createdCount, pending: w.pendingCount })}
        </TableCell>
        <TableCell>{w.lastRunAt ? relative(w.lastRunAt) : messages.neverRun}</TableCell>
        <TableCell>
          <RowMenu
            testId={`backlog-pr-watch-menu-${w.id}`}
            label={format(messages.rowActions, { name: w.name })}
            items={[
              {
                testId: `backlog-pr-watch-edit-${w.id}`,
                label: messages.edit,
                onSelect: () => setEditing(w),
              },
              ...(w.state === "active"
                ? [
                    {
                      testId: `backlog-pr-watch-run-${w.id}`,
                      label: messages.runNow,
                      onSelect: () =>
                        void act("git.watches.run", w.id, () => setNotice({ key: "runQueued" })),
                    },
                    {
                      testId: `backlog-pr-watch-pause-${w.id}`,
                      label: messages.pause,
                      onSelect: () => void act("git.watches.pause", w.id, (r) => replace(r as Watch)),
                    },
                  ]
                : []),
              ...(w.state === "paused"
                ? [
                    {
                      testId: `backlog-pr-watch-resume-${w.id}`,
                      label: messages.resume,
                      onSelect: () => void act("git.watches.resume", w.id, (r) => replace(r as Watch)),
                    },
                  ]
                : []),
              {
                testId: `backlog-pr-watch-delete-${w.id}`,
                label: messages.delete,
                onSelect: () => setDeleting(w),
              },
            ]}
          />
        </TableCell>
      </TableRow>
    ));

    const body = list.error ? (
      <ListError testId="backlog-pr-watches" notice={list.error} onRetry={list.reload} />
    ) : !list.loading && list.items.length === 0 ? (
      <ListEmpty testId="backlog-pr-watches" title={messages.watchesEmpty}>
        {add("backlog-pr-watches-empty-add")}
      </ListEmpty>
    ) : (
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{messages.colName}</TableHead>
                <TableHead>{messages.colRepository}</TableHead>
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
      <div id={PR_WATCHES_ANCHOR} data-testid="backlog-section-pr-watches">
        <SettingsSection
          title={messages.watchesTitle}
          description={messages.prWatchesDescription}
          action={add("backlog-pr-watches-add")}
        >
          {body}
          <p role="status" data-testid="backlog-pr-watches-notice">
            {notice ? noticeText(notice, messages) : ""}
          </p>
        </SettingsSection>
        <Dialog open={Boolean(editing)} onOpenChange={(open: boolean) => !open && setEditing(undefined)}>
          <DialogContent
            data-testid="backlog-pr-watch-dialog"
            aria-labelledby="backlog-pr-watch-dialog-title"
          >
            <DialogHeader>
              <DialogTitle id="backlog-pr-watch-dialog-title">
                {editing?.id ? messages.editPrWatch : messages.addPrWatch}
              </DialogTitle>
            </DialogHeader>
            {editing ? (
              <WatchForm
                workspaceId={workspaceId}
                watch={editing}
                onSaved={(w: Watch) => {
                  replace(w);
                  setEditing(undefined);
                }}
                onCancel={() => setEditing(undefined)}
              />
            ) : null}
          </DialogContent>
        </Dialog>
        {deleting ? (
          <ConfirmDialog
            testId="backlog-pr-watch-delete-dialog"
            title={messages.deleteWatchTitle}
            body={messages.deleteWatchBody}
            confirmLabel={messages.delete}
            destructive
            onConfirm={async () => {
              await host.api.invokeAction("git.watches.delete", { workspaceId, body: { id: deleting.id } });
              list.setItems((items) => items.filter((x) => x.id !== deleting.id));
            }}
            onClose={() => setDeleting(undefined)}
          />
        ) : null}
      </div>
    );
  };
}
