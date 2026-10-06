import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { en, format, type Messages } from "../messages/en";
import type { Notice } from "../settings/state";
import { issueNotice, type TaskLink } from "./issues-state";

type AnyProps = Record<string, unknown>;

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

const STACK = "flex flex-col gap-4";
const ROW = "flex gap-2";

/**
 * Link to task (M3, US3.3): a dialog that searches the workspace's tasks
 * and links the chosen one to the issue. Kandev has no task picker, so the
 * plugin draws it: role="dialog", focus on the search field, Esc closes, and
 * the focus goes back to the element that opened it. A task linked to
 * another issue cannot be chosen (AC3.3.3).
 */
export function createLinkTaskDialog(
  host: PluginHostApi,
  messages: Messages = en,
): Component<LinkTaskDialogProps> {
  const h = host.jsx;
  const { useEffect, useRef, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const Input = host.ui.Input as Component<AnyProps>;
  const Label = host.ui.Label as Component<AnyProps>;

  return function LinkTaskDialog({ workspaceId, issueKey, onLinked, onClose }: LinkTaskDialogProps) {
    const [query, setQuery] = useState("");
    const [tasks, setTasks] = useState<TaskItem[] | undefined>(undefined);
    const [chosen, setChosen] = useState<string | undefined>(undefined);
    const [linking, setLinking] = useState(false);
    const [error, setError] = useState<Notice | undefined>(undefined);
    const root = useRef<HTMLDivElement | null>(null);
    const latest = useRef(0);

    useEffect(() => {
      const opener = document.activeElement as HTMLElement | null;
      root.current?.querySelector<HTMLElement>('[data-testid="backlog-link-task-search"]')?.focus();
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
        onLinked({ taskId: chosen, taskKey: l?.taskKey });
        onClose();
      } catch (e) {
        setError(issueNotice(e));
        setLinking(false);
      }
    };

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !linking) {
        e.preventDefault();
        onClose();
      }
    };

    return (
      <div
        ref={root}
        role="dialog"
        aria-modal="true"
        aria-labelledby="backlog-link-task-title"
        data-testid="backlog-link-task-dialog"
        className={STACK}
        onKeyDown={onKeyDown}
      >
        <h2 id="backlog-link-task-title">{format(messages.linkTaskTitle, { key: issueKey })}</h2>
        <div className="flex flex-col gap-2">
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
                  <button
                    type="button"
                    data-testid={`backlog-link-task-option-${t.taskId}`}
                    className={ROW}
                    aria-pressed={chosen === t.taskId}
                    aria-disabled={other}
                    onClick={() => !other && setChosen(t.taskId)}
                  >
                    <span>{t.taskKey ?? t.taskId}</span>
                    <span>{t.title}</span>
                    {other ? <span>{format(messages.linkedTo, { key: t.linkedIssueKey ?? "" })}</span> : null}
                  </button>
                </li>
              );
            })}
          </ul>
        ) : null}
        {error ? (
          <p role="alert" data-testid="backlog-link-task-error">
            {noticeText(error, messages)}
          </p>
        ) : null}
        <div className={ROW}>
          <Button type="button" data-testid="backlog-link-task-cancel" disabled={linking} onClick={onClose}>
            {messages.cancel}
          </Button>
          <Button
            type="button"
            data-testid="backlog-link-task-submit"
            disabled={!chosen || linking}
            onClick={() => void link()}
          >
            {linking ? messages.linking : messages.link}
          </Button>
        </div>
      </div>
    );
  };
}
