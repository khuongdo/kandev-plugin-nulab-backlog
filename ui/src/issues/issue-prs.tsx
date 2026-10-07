import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText, providerName, scmNotice, stateText } from "../git/git-state";
import { hostUi } from "../host-ui";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";

/** One scm.links.list link: manual to the task or auto to the issue (FR5.1, FR5.2). */
interface ScmLink {
  provider: string;
  repo: string;
  number: number;
  taskId?: string;
  issueKey?: string;
  auto?: boolean;
  title?: string;
  state?: string;
  url: string;
}

/** One pull request Kandev itself attached to a linked task (FR5.4). */
interface KandevPR {
  taskId: string;
  number: number;
  url: string;
  title: string;
  state: string;
  isDraft?: boolean;
  provider: string;
}

export interface IssuePullRequestsProps {
  workspaceId: string;
  taskId: string;
  issueKey: string;
}

const keyOf = (l: ScmLink) => `${l.provider}|${l.repo}|${l.number}`;

/**
 * The pull requests of a task's Backlog issue (FR5): links to any provider,
 * auto-links with a badge and Remove (FR5.3), Kandev's own GitHub and GitLab
 * pull requests (FR5.4), and linking a pasted URL (FR5.1). Each list loads
 * on its own, so one failure never hides the other or the issue.
 */
export function createIssuePullRequests(
  host: PluginHostApi,
  messages: Messages = en,
): Component<IssuePullRequestsProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const { Badge, Button, Input, Label } = hostUi(host);
  const t = (n: Notice) => noticeText(n, messages);
  const line = (provider: string, number: number, state: string) =>
    format(messages.prLine, {
      provider: providerName(provider, messages),
      number,
      state: stateText(state, messages),
    });

  return function IssuePullRequests({ workspaceId, taskId, issueKey }: IssuePullRequestsProps) {
    const [links, setLinks] = useState<ScmLink[] | undefined>(undefined);
    const [kandev, setKandev] = useState<KandevPR[] | undefined>(undefined);
    const [error, setError] = useState<Notice | undefined>(undefined);
    const [url, setUrl] = useState("");
    const [notice, setNotice] = useState<Notice | undefined>(undefined);
    const [busy, setBusy] = useState(false);
    const [adding, setAdding] = useState(false);

    useEffect(() => {
      host.api
        .invokeAction<{ links?: ScmLink[] }>("scm.links.list", { workspaceId, body: { taskId, issueKey } })
        .then((r) => setLinks(r?.links ?? []))
        .catch((e: unknown) => {
          setLinks([]);
          setError(scmNotice(e));
        });
      host.api
        .invokeAction<{ pullRequests?: KandevPR[] }>("scm.task_prs.list", { workspaceId, taskId })
        .then((r) => setKandev(r?.pullRequests ?? []))
        .catch((e: unknown) => {
          setKandev([]);
          setError(scmNotice(e));
        });
    }, [workspaceId, taskId, issueKey]);

    const remove = async (l: ScmLink) => {
      setNotice(undefined);
      try {
        await host.api.invokeAction("scm.prs.unlink", {
          workspaceId,
          taskId,
          body: l.auto ? { key: keyOf(l), issueKey: l.issueKey } : { key: keyOf(l) },
        });
        setLinks((list) =>
          list?.filter((x) => !(keyOf(x) === keyOf(l) && Boolean(x.auto) === Boolean(l.auto))),
        );
      } catch (e) {
        setNotice(scmNotice(e));
      }
    };
    const link = async () => {
      if (busy || !url.trim()) return;
      setBusy(true);
      setNotice(undefined);
      try {
        const l = await host.api.invokeAction<ScmLink>("scm.prs.link", {
          workspaceId,
          taskId,
          body: { url: url.trim() },
        });
        setLinks((list) => [...(list ?? []), l]);
        setUrl("");
      } catch (e) {
        const f = readFailure(e);
        setNotice(f.code === "validation" && f.field === "url" ? { key: "errorPrUrl" } : scmNotice(e));
      }
      setBusy(false);
    };

    const none = links !== undefined && kandev !== undefined && links.length === 0 && kandev.length === 0;
    return (
      <section className={STACK} aria-labelledby="backlog-issue-prs-title" data-testid="backlog-issue-prs">
        <h3 id="backlog-issue-prs-title">{messages.scopePRs}</h3>
        {error ? (
          <p role="alert" data-testid="backlog-issue-prs-error">
            {messages.prsLoadFailed} {t(error)}
          </p>
        ) : null}
        {none ? <p data-testid="backlog-issue-prs-empty">{messages.noPullRequests}</p> : null}
        <ul className={STACK}>
          {(links ?? []).map((l) => {
            const k = keyOf(l);
            return (
              <li
                key={`${k}|${l.auto ? "auto" : "task"}`}
                data-testid={`backlog-issue-pr-${k}`}
                className={ROW}
              >
                <a href={l.url} target="_blank" rel="noreferrer" data-testid={`backlog-issue-pr-link-${k}`}>
                  {l.title || format(messages.prNumber, { number: l.number })}
                </a>
                <span>{line(l.provider, l.number, l.state ?? "")}</span>
                {l.auto ? (
                  <Badge variant="secondary" data-testid={`backlog-issue-pr-auto-${k}`}>
                    {messages.autoLinked}
                  </Badge>
                ) : null}
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className={BUTTON}
                  data-testid={`backlog-issue-pr-remove-${k}`}
                  aria-label={format(messages.removeLink, {
                    name: l.title || format(messages.prNumber, { number: l.number }),
                  })}
                  onClick={() => void remove(l)}
                >
                  {messages.remove}
                </Button>
              </li>
            );
          })}
          {(kandev ?? []).map((p) => (
            <li
              key={`${p.taskId}-${p.number}`}
              data-testid={`backlog-issue-kandev-pr-${p.taskId}-${p.number}`}
              className={ROW}
            >
              <a
                href={p.url}
                target="_blank"
                rel="noreferrer"
                data-testid={`backlog-issue-kandev-pr-link-${p.taskId}-${p.number}`}
              >
                {p.title || format(messages.prNumber, { number: p.number })}
              </a>
              <span>{line(p.provider, p.number, p.isDraft && p.state === "open" ? "draft" : p.state)}</span>
              <span>{messages.fromKandev}</span>
            </li>
          ))}
        </ul>
        {adding ? (
          <div className={ROW}>
            <div className={FIELD}>
              <Label htmlFor="backlog-issue-pr-url">{messages.prUrlLabel}</Label>
              <Input
                id="backlog-issue-pr-url"
                data-testid="backlog-issue-pr-url"
                value={url}
                onChange={(e: { target: { value: string } }) => setUrl(e.target.value)}
              />
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={BUTTON}
              data-testid="backlog-issue-pr-link"
              disabled={busy}
              onClick={() => void link()}
            >
              {messages.linkPr}
            </Button>
          </div>
        ) : (
          <div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={BUTTON}
              data-testid="backlog-issue-pr-add"
              onClick={() => setAdding(true)}
            >
              {messages.linkPrOpen}
            </Button>
          </div>
        )}
        {notice ? (
          <p role="status" data-testid="backlog-issue-pr-notice">
            {t(notice)}
          </p>
        ) : null}
      </section>
    );
  };
}
