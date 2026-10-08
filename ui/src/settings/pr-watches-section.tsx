import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { gitNotice, noticeText, providerName, scmNotice, watchStatusKey } from "../git/git-state";
import type { ScmWatch } from "../git/scm-watch-form";
import { createWatchForm, type Watch } from "../git/watch-form";
import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { createSectionParts } from "./section-parts";
import type { Notice } from "./state";
import { useActionList } from "./use-list";

type AnyWatch = Watch | ScmWatch;

/** A GitHub, GitLab or Bitbucket watch (they carry a provider). */
const isScm = (w: AnyWatch): w is ScmWatch => "provider" in w && Boolean(w.provider);
/** The action family of a watch. */
const family = (w: AnyWatch) => (isScm(w) ? "scm.watches" : "git.watches");

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
): Component<{ workspaceId: string; selectedProjects?: string[] }> {
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

  return function PrWatchesSection({
    workspaceId,
    selectedProjects = [],
  }: {
    workspaceId: string;
    selectedProjects?: string[];
  }) {
    const list = useActionList<Watch>(host, workspaceId, "git.watches.list", "watches");
    // Intent 261007-source-control-agnostic (FR4.3): GitHub, GitLab and Bitbucket watches in the same table.
    const scmList = useActionList<ScmWatch>(host, workspaceId, "scm.watches.list", "watches");
    const [editing, setEditing] = useState<Partial<Watch> | Partial<ScmWatch> | undefined>(undefined);
    const [deleting, setDeleting] = useState<AnyWatch | undefined>(undefined);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);
    const all: AnyWatch[] = [...list.items, ...scmList.items];

    const replace = (w: AnyWatch) => {
      const put = <T extends { id: string }>(items: T[]) =>
        items.some((x) => x.id === w.id)
          ? items.map((x) => (x.id === w.id ? (w as unknown as T) : x))
          : [...items, w as unknown as T];
      if (isScm(w)) scmList.setItems(put);
      else list.setItems(put);
    };
    const act = async (w: AnyWatch, verb: string, done: (reply: unknown) => void) => {
      setNotice(undefined);
      try {
        done(await host.api.invokeAction(`${family(w)}.${verb}`, { workspaceId, body: { id: w.id } }));
      } catch (e) {
        setNotice(isScm(w) ? scmNotice(e) : gitNotice(e));
      }
    };

    const add = (testId: string) => (
      <Button type="button" size="sm" className={BUTTON} data-testid={testId} onClick={() => setEditing({})}>
        {icon(host, "plus", "mr-1 h-4 w-4")}
        {messages.addWatch}
      </Button>
    );

    const rows = all.map((w) => (
      <TableRow key={w.id} data-testid={`backlog-pr-watch-row-${w.id}`}>
        <TableCell>{w.name}</TableCell>
        <TableCell>
          {isScm(w)
            ? format(messages.scmWatchRepo, {
                provider: providerName(w.provider, messages),
                project: w.projectKey,
                repo: w.repo,
              })
            : format(messages.watchRepo, { project: w.projectKey, repo: w.repoName })}
        </TableCell>
        <TableCell>
          <Badge variant="outline">{messages[watchStatusKey(w.state)]}</Badge>
          {isScm(w) && w.unmapped ? (
            <Badge variant="secondary" data-testid={`backlog-pr-watch-unmapped-${w.id}`}>
              {messages.scmUnmapped}
            </Badge>
          ) : null}
        </TableCell>
        <TableCell>
          {format(messages.progressValue, {
            created: w.createdCount,
            pending: isScm(w) ? 0 : w.pendingCount,
          })}
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
                        void act(w, "run", (r) =>
                          setNotice(
                            isScm(w)
                              ? {
                                  key: "scmRunDone",
                                  params: { count: (r as { created?: number })?.created ?? 0 },
                                }
                              : { key: "runQueued" },
                          ),
                        ),
                    },
                    {
                      testId: `backlog-pr-watch-pause-${w.id}`,
                      label: messages.pause,
                      onSelect: () => void act(w, "pause", (r) => replace(r as AnyWatch)),
                    },
                  ]
                : []),
              ...(w.state === "paused"
                ? [
                    {
                      testId: `backlog-pr-watch-resume-${w.id}`,
                      label: messages.resume,
                      onSelect: () => void act(w, "resume", (r) => replace(r as AnyWatch)),
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
    ) : !list.loading && !scmList.loading && all.length === 0 ? (
      // FR4.3: the section header holds the only Add watch button.
      <ListEmpty testId="backlog-pr-watches" title={messages.watchesEmpty} />
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
          {scmList.error ? (
            <ListError testId="backlog-scm-watches" notice={scmList.error} onRetry={scmList.reload} />
          ) : null}
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
                selectedProjects={selectedProjects}
                onSaved={(w: AnyWatch) => {
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
              await host.api.invokeAction(`${family(deleting)}.delete`, {
                workspaceId,
                body: { id: deleting.id },
              });
              if (isScm(deleting)) scmList.setItems((items) => items.filter((x) => x.id !== deleting.id));
              else list.setItems((items) => items.filter((x) => x.id !== deleting.id));
            }}
            onClose={() => setDeleting(undefined)}
          />
        ) : null}
      </div>
    );
  };
}
