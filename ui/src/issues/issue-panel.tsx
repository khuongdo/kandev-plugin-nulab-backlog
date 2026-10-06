import type {
  Component,
  PluginHostApi,
  PluginTaskPanelProps,
  TaskPanelRegistration,
} from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { en, format, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";
import { issueNotice } from "./issues-state";
import type { LinksStore } from "./links-store";

type AnyProps = Record<string, unknown>;

interface Detail {
  issueKey: string;
  spaceHost?: string;
  linkState: string;
  statusUpdatedAt?: string;
  issue?: {
    key: string;
    summary: string;
    status: string;
    assignee?: string;
    priority?: string;
    dueDate?: string;
    url: string;
  };
  attachments?: { name: string; size: number; tooLarge: boolean }[];
  attachmentsError?: string;
}

interface Comment {
  id: number;
  author: string;
  content: string;
  created: string;
}

const STACK = "flex flex-col gap-4";
const ROW = "flex gap-2";

type Load =
  | { kind: "loading" }
  | { kind: "ready"; detail: Detail }
  | { kind: "state"; text: string }
  | { kind: "failed"; notice: Notice };

/** The expanded state of a section, kept per user in this browser. */
function useSection(host: PluginHostApi, name: string): [boolean, () => void] {
  const key = `${host.pluginId}:panel:${name}`;
  const [open, setOpen] = host.React.useState(() => {
    try {
      return localStorage.getItem(key) !== "false";
    } catch {
      return true;
    }
  });
  const toggle = () =>
    setOpen((o) => {
      try {
        localStorage.setItem(key, String(!o));
      } catch {
        // storage blocked: the choice lasts until the panel closes
      }
      return !o;
    });
  return [open, toggle];
}

/**
 * The Backlog issue panel of a task (M8, US3.2, US3.5), opened from the
 * task's panel menu of a linked task. It reads the issue live on every open and never writes
 * to it: key, title, status, assignee, priority, due date, comments (newest
 * first, 20 at a time) and attachments, which link to Backlog.
 */
export function createIssuePanel(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): TaskPanelRegistration {
  const h = host.jsx;
  const { useCallback, useEffect, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;
  const t = (n: Notice) => noticeText(n, messages);

  function Comments({ workspaceId, taskId }: { workspaceId: string; taskId: string }) {
    const [open, toggle] = useSection(host, "comments");
    const [items, setItems] = useState<Comment[] | undefined>(undefined);
    const [next, setNext] = useState<number | undefined>(undefined);
    const [error, setError] = useState<Notice | undefined>(undefined);

    const load = useCallback(
      async (maxId?: number) => {
        setError(undefined);
        try {
          const r = await host.api.invokeAction<{ comments?: Comment[]; nextMaxId?: number }>(
            "issues.comments",
            {
              workspaceId,
              taskId,
              body: maxId ? { maxId } : {},
            },
          );
          setItems((list) => [...(maxId ? (list ?? []) : []), ...(r?.comments ?? [])]);
          setNext(r?.nextMaxId || undefined);
        } catch (e) {
          setError(issueNotice(e));
        }
      },
      [workspaceId, taskId],
    );

    useEffect(() => {
      if (open && items === undefined) void load();
    }, [open]);

    return (
      <section className={STACK}>
        <h3>
          <button
            type="button"
            data-testid="backlog-issue-comments-toggle"
            aria-expanded={open}
            aria-controls="backlog-issue-comments"
            onClick={toggle}
          >
            {messages.comments}
          </button>
        </h3>
        {open ? (
          <div id="backlog-issue-comments" className={STACK}>
            {error ? (
              <div data-testid="backlog-issue-comments-error" className={ROW}>
                <p>{messages.commentsFailed}</p>
                <p>{t(error)}</p>
                <Button type="button" data-testid="backlog-issue-comments-retry" onClick={() => void load()}>
                  {messages.retry}
                </Button>
              </div>
            ) : null}
            {items && items.length === 0 ? <p>{messages.noComments}</p> : null}
            {items && items.length > 0 ? (
              <ul data-testid="backlog-issue-comments-list" className={STACK}>
                {items.map((c) => (
                  <li key={c.id} className="flex flex-col gap-1">
                    <span>{c.author}</span>
                    {c.created ? <span>{relative(c.created)}</span> : null}
                    <p className="whitespace-pre-wrap">{c.content}</p>
                  </li>
                ))}
              </ul>
            ) : null}
            {next ? (
              <div>
                <Button
                  type="button"
                  data-testid="backlog-issue-comments-more"
                  onClick={() => void load(next)}
                >
                  {messages.loadMore}
                </Button>
              </div>
            ) : null}
          </div>
        ) : null}
      </section>
    );
  }

  function Attachments({ detail }: { detail: Detail }) {
    const [open, toggle] = useSection(host, "attachments");
    const list = detail.attachments ?? [];
    const url = detail.issue?.url ?? "";
    return (
      <section className={STACK}>
        <h3>
          <button
            type="button"
            data-testid="backlog-issue-attachments-toggle"
            aria-expanded={open}
            aria-controls="backlog-issue-attachments"
            onClick={toggle}
          >
            {messages.attachments}
          </button>
        </h3>
        {open ? (
          <div id="backlog-issue-attachments" className={STACK}>
            {detail.attachmentsError ? (
              <p>{t(issueNotice({ body: { error: { code: detail.attachmentsError } } }))}</p>
            ) : null}
            {!detail.attachmentsError && list.length === 0 ? <p>{messages.noAttachments}</p> : null}
            <ul className={STACK}>
              {list.map((a, i) => (
                <li key={i} data-testid={`backlog-issue-attachment-${i}`} className={ROW}>
                  <a
                    href={url}
                    target="_blank"
                    rel="noreferrer"
                    data-testid={`backlog-issue-attachment-link-${i}`}
                  >
                    {a.name}
                  </a>
                  <span>{format(messages.sizeMB, { size: (a.size / 1048576).toFixed(1) })}</span>
                  {a.tooLarge ? <span>{messages.tooLargeToPreview}</span> : null}
                </li>
              ))}
            </ul>
          </div>
        ) : null}
      </section>
    );
  }

  function IssuePanel({ taskId }: PluginTaskPanelProps) {
    const workspaceId = host.context.getActiveWorkspaceId() ?? "";
    const [load, setLoad] = useState<Load>({ kind: "loading" });

    const reload = useCallback(async () => {
      setLoad({ kind: "loading" });
      try {
        const detail = await host.api.invokeAction<Detail>("issues.get", { workspaceId, taskId });
        if (detail.linkState !== "active") {
          setLoad({
            kind: "state",
            text: format(messages.reconnectToRestore, { host: detail.spaceHost ?? "" }),
          });
          return;
        }
        setLoad({ kind: "ready", detail });
      } catch (e) {
        const code = readFailure(e).code;
        if (code === "not_found") {
          // The panel is offered for linked tasks only, so not_found is an
          // issue Backlog no longer shows (or a link removed meanwhile).
          setLoad({ kind: "state", text: messages.issueUnavailableLong });
          return;
        }
        setLoad({ kind: "failed", notice: issueNotice(e) });
      }
    }, [workspaceId, taskId]);

    useEffect(() => {
      void reload();
    }, [reload]);

    if (load.kind === "loading") return <p data-testid="backlog-issue-panel">{messages.issueLoading}</p>;
    if (load.kind === "state") {
      return (
        <div data-testid="backlog-issue-panel">
          <p data-testid="backlog-issue-state">{load.text}</p>
        </div>
      );
    }
    if (load.kind === "failed") {
      return (
        <div data-testid="backlog-issue-panel" className={STACK}>
          <p>{messages.issueLoadFailed}</p>
          <p>{t(load.notice)}</p>
          <div>
            <Button type="button" data-testid="backlog-issue-retry" onClick={() => void reload()}>
              {messages.retry}
            </Button>
          </div>
        </div>
      );
    }
    const { detail } = load;
    const issue = detail.issue!;
    const field = (label: string, value: unknown, testId?: string) => (
      <div className={ROW} data-testid={testId}>
        <dt>{label}</dt>
        <dd>{value || messages.noValue}</dd>
      </div>
    );
    const due = issue.dueDate ? (
      <time dateTime={issue.dueDate}>
        {new Date(issue.dueDate).toLocaleDateString(host.i18n?.locale ?? "en", { dateStyle: "long" })}
      </time>
    ) : null;
    return (
      <div data-testid="backlog-issue-panel" className={STACK}>
        <h2 className={ROW}>
          <span>{issue.key}</span>
          <span>{issue.summary}</span>
        </h2>
        <dl className={STACK}>
          {field(messages.fieldStatus, issue.status)}
          {field(messages.fieldAssignee, issue.assignee)}
          {field(messages.fieldPriority, issue.priority)}
          {field(messages.fieldDueDate, due, "backlog-issue-due")}
        </dl>
        {detail.statusUpdatedAt ? (
          <p data-testid="backlog-issue-updated">
            {format(messages.issueUpdatedAt, { time: relative(detail.statusUpdatedAt) })}
          </p>
        ) : null}
        <a href={issue.url} target="_blank" rel="noreferrer" data-testid="backlog-issue-open">
          {messages.openInBacklog}
        </a>
        <Comments workspaceId={workspaceId} taskId={taskId} />
        <Attachments detail={detail} />
      </div>
    );
  }

  return {
    id: "backlog-issue",
    title: messages.issuePanelTitle,
    Component: IssuePanel,
    mobileEnabled: true,
    // Offered only for a linked task, decided synchronously from the shared
    // links (the first call starts their load).
    visible: store
      ? ({ taskId }) => {
          const ws = host.context.getActiveWorkspaceId() ?? "";
          const links = store.get(ws);
          if (!links) void store.load(ws);
          return Boolean(links?.some((l) => l.taskId === taskId));
        }
      : undefined,
  };
}
