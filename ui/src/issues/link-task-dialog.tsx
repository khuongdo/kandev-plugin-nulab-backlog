import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { hostUi } from "../host-ui";
import { BUTTON, FIELD, STACK } from "../layout";
import { en, format, type Messages } from "../messages/en";
import type { Notice } from "../settings/state";
import { issueNotice, type TaskLink } from "./issues-state";
import type { LinksStore } from "./links-store";

interface TaskItem {
  taskId: string;
  taskKey?: string;
  title: string;
  linkedIssueKey?: string;
}

export interface LinkTaskDialogProps {
  workspaceId: string;
  issueKey: string;
  onLinked: (task: TaskLink) => void;
  onClose: () => void;
}

/**
 * Link to task (M3, US3.3): a host Dialog that searches the workspace's
 * tasks and links the chosen one to the issue. Kandev's own link dialog
 * (`openTaskLinkDialog`) takes one text field and no task list, so the plugin
 * builds this one in the same shell as Kandev's GitHub link dialog (FR4):
 * width, description line, inline red error, Cancel + Save/Saving..., Enter
 * submits, success toast. The focus starts in the search field, Esc closes,
 * and the focus goes back to the element that opened it. A task linked to
 * another issue cannot be chosen (AC3.3.3). After a link the shared links
 * store refreshes, so every badge shows it at once (FR4.5).
 */
export function createLinkTaskDialog(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): Component<LinkTaskDialogProps> {
  const h = host.jsx;
  const { useEffect, useRef, useState } = host.React;
  const ui = hostUi(host);
  const { Button, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } = ui;
  const { Input, Label } = ui;

  return function LinkTaskDialog({ workspaceId, issueKey, onLinked, onClose }: LinkTaskDialogProps) {
    const [query, setQuery] = useState("");
    const [tasks, setTasks] = useState<TaskItem[] | undefined>(undefined);
    const [chosen, setChosen] = useState<string | undefined>(undefined);
    const [linking, setLinking] = useState(false);
    const [error, setError] = useState<Notice | undefined>(undefined);
    const latest = useRef(0);

    useEffect(() => {
      const opener = document.activeElement as HTMLElement | null;
      document.querySelector<HTMLElement>('[data-testid="backlog-link-task-search"]')?.focus();
      return () => opener?.focus();
    }, []);

    useEffect(() => {
      const seq = ++latest.current;
      host.api
        .invokeAction<{ tasks?: TaskItem[] }>("issues.tasks.search", { workspaceId, body: { query } })
        .then((r) => seq === latest.current && setTasks(r?.tasks ?? []))
        .catch((e) => seq === latest.current && setError(issueNotice(e)));
    }, [query]);

    const link = async () => {
      if (!chosen || linking) return;
      setLinking(true);
      setError(undefined);
      try {
        const l = await host.api.invokeAction<TaskLink>("issues.link", {
          workspaceId,
          taskId: chosen,
          body: { issueKey },
        });
        // FR4.5: the GitHub dialog's toast, and fresh links for every badge.
        host.toast.success(messages.linkIssueSuccess);
        void store?.refresh(workspaceId).catch(() => undefined);
        onLinked({ taskId: chosen, taskKey: l?.taskKey });
        onClose();
      } catch (e) {
        setError(issueNotice(e));
        setLinking(false);
      }
    };

    return (
      <Dialog open onOpenChange={(open: boolean) => !open && !linking && onClose()}>
        <DialogContent
          className="w-[calc(100vw-2rem)] sm:max-w-lg"
          data-testid="backlog-link-task-dialog"
          aria-labelledby="backlog-link-task-title"
          aria-describedby="backlog-link-task-description"
        >
          <DialogHeader>
            <DialogTitle id="backlog-link-task-title">
              {format(messages.linkTaskTitle, { key: issueKey })}
            </DialogTitle>
            <DialogDescription id="backlog-link-task-description">
              {messages.linkTaskDescription}
            </DialogDescription>
          </DialogHeader>
          <form
            className="contents"
            onSubmit={(e: { preventDefault(): void }) => {
              e.preventDefault();
              void link();
            }}
          >
            <div className={FIELD}>
              <Label htmlFor="backlog-link-task-search">{messages.searchTasks}</Label>
              <Input
                id="backlog-link-task-search"
                data-testid="backlog-link-task-search"
                type="search"
                value={query}
                onChange={(e: { target: { value: string } }) => setQuery(e.target.value)}
              />
            </div>
            {tasks && tasks.length === 0 ? (
              <p data-testid="backlog-link-task-none">{messages.noTasksFound}</p>
            ) : null}
            {tasks && tasks.length > 0 ? (
              <ul className={STACK}>
                {tasks.map((t) => {
                  const other = Boolean(t.linkedIssueKey) && t.linkedIssueKey !== issueKey;
                  return (
                    <li key={t.taskId}>
                      <Button
                        type="button"
                        variant={chosen === t.taskId ? "secondary" : "ghost"}
                        className={`${BUTTON} h-auto w-full justify-start gap-2`}
                        data-testid={`backlog-link-task-option-${t.taskId}`}
                        aria-pressed={chosen === t.taskId}
                        aria-disabled={other}
                        onClick={() => !other && setChosen(t.taskId)}
                      >
                        <span>{t.taskKey ?? t.taskId}</span>
                        <span>{t.title}</span>
                        {other ? (
                          <span>{format(messages.linkedTo, { key: t.linkedIssueKey ?? "" })}</span>
                        ) : null}
                      </Button>
                    </li>
                  );
                })}
              </ul>
            ) : null}
            {error ? (
              <p role="alert" className="text-xs text-destructive" data-testid="backlog-link-task-error">
                {noticeText(error, messages)}
              </p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                className={BUTTON}
                data-testid="backlog-link-task-cancel"
                disabled={linking}
                onClick={onClose}
              >
                {messages.cancel}
              </Button>
              <Button
                type="submit"
                className={BUTTON}
                data-testid="backlog-link-task-submit"
                disabled={!chosen || linking}
              >
                {linking ? messages.saving : messages.save}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    );
  };
}
