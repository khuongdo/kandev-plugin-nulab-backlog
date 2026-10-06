import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "../settings/confirm-dialog";
import type { Notice } from "../settings/state";
import { gitNotice, loadRepoOptions, noticeText, splitRepo, stateText, type RepoOption } from "./git-state";

type AnyProps = Record<string, unknown>;

interface Query {
  id: string;
  name: string;
}

interface Row {
  number: number;
  title: string;
  state: string;
  url: string;
  linkedTaskIds: string[];
}

const STACK = "flex flex-col gap-4";
const FIELD = "flex flex-col gap-2";
const ROW = "flex gap-2";

type Result =
  { kind: "idle" | "loading" } | { kind: "failed"; notice: Notice } | { kind: "ready"; rows: Row[] };

/** The PR dashboard at /backlog/dashboard: saved queries and their PRs (M5, US6.3). */
export function createDashboardPage(host: PluginHostApi, messages: Messages = en): Component {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const Input = host.ui.Input as Component<AnyProps>;
  const Label = host.ui.Label as Component<AnyProps>;
  const ConfirmDialog = createConfirmDialog(host, messages);
  const t = (n: Notice) => noticeText(n, messages);

  return function DashboardPage() {
    const [workspaceId, setWorkspaceId] = useState(host.context.getActiveWorkspaceId());
    const [queries, setQueries] = useState<Query[] | undefined>(undefined);
    const [loadNotice, setLoadNotice] = useState<Notice | undefined>(undefined);
    const [selected, setSelected] = useState("");
    const [result, setResult] = useState<Result>({ kind: "idle" });
    const [form, setForm] = useState(false);
    const [repos, setRepos] = useState<RepoOption[]>([]);
    const [name, setName] = useState("");
    const [repo, setRepo] = useState("");
    const [formError, setFormError] = useState<"name" | "repo" | undefined>(undefined);
    const [deleting, setDeleting] = useState(false);

    useEffect(() => host.context.subscribeActiveWorkspace(setWorkspaceId), []);
    useEffect(() => {
      if (!workspaceId) return;
      host.api.invokeAction<{ queries?: Query[] }>("git.queries.list", { workspaceId }).then(
        (r) => setQueries(r?.queries ?? []),
        (e: unknown) => setLoadNotice(gitNotice(e)),
      );
    }, [workspaceId]);

    const run = async (id: string) => {
      setSelected(id);
      if (!id) return setResult({ kind: "idle" });
      setResult({ kind: "loading" });
      try {
        const reply = await host.api.invokeAction<{ rows?: Row[] }>("git.queries.run", {
          workspaceId,
          body: { id },
        });
        setResult({ kind: "ready", rows: reply?.rows ?? [] });
      } catch (error) {
        setResult({ kind: "failed", notice: gitNotice(error) });
      }
    };

    const openForm = () => {
      setForm(true);
      if (workspaceId) loadRepoOptions(host, workspaceId, messages).then(setRepos, () => setRepos([]));
    };

    const saveQuery = async () => {
      const trimmed = name.trim();
      const error = trimmed === "" || trimmed.length > 100 ? "name" : !repo ? "repo" : undefined;
      setFormError(error);
      if (error) {
        document.getElementById(error === "name" ? "backlog-query-name" : "backlog-query-repo")?.focus();
        return;
      }
      try {
        const q = await host.api.invokeAction<Query>("git.queries.save", {
          workspaceId,
          body: { name: trimmed, ...splitRepo(repo), statuses: ["open"], assignee: "anyone" },
        });
        setQueries((list) => [...(list ?? []), q]);
        setForm(false);
        setName("");
        setRepo("");
      } catch (error) {
        setLoadNotice(gitNotice(error));
      }
    };

    const remove = async () => {
      await host.api.invokeAction("git.queries.delete", { workspaceId, body: { id: selected } });
      setQueries((list) => list?.filter((q) => q.id !== selected));
      setSelected("");
      setResult({ kind: "idle" });
    };

    const fieldError = (field: "name" | "repo", id: string) =>
      formError === field ? (
        <p id={`${id}-error`} data-testid={`${id}-error`}>
          {field === "name" ? messages.errorName : messages.errorRepository}
        </p>
      ) : null;

    return (
      <div data-testid="backlog-dashboard" className={STACK}>
        {loadNotice ? (
          <p data-testid="backlog-dashboard-notice">
            {messages.queriesLoadFailed} {t(loadNotice)}
          </p>
        ) : null}
        {queries && queries.length === 0 ? (
          <p data-testid="backlog-queries-empty">{messages.noQueries}</p>
        ) : null}
        <div className={FIELD}>
          <Label htmlFor="backlog-query-select">{messages.queryLabel}</Label>
          <select
            id="backlog-query-select"
            data-testid="backlog-query-select"
            value={selected}
            onChange={(e: { target: { value: string } }) => void run(e.target.value)}
          >
            <option value="">{messages.chooseQuery}</option>
            {(queries ?? []).map((q) => (
              <option key={q.id} value={q.id}>
                {q.name}
              </option>
            ))}
          </select>
        </div>
        <div className={ROW}>
          <Button type="button" data-testid="backlog-query-new" onClick={openForm}>
            {messages.newQuery}
          </Button>
          {selected ? (
            <Button type="button" data-testid="backlog-query-delete" onClick={() => setDeleting(true)}>
              {messages.deleteQuery}
            </Button>
          ) : null}
        </div>
        {form ? (
          <div data-testid="backlog-query-form" className={STACK}>
            <div className={FIELD}>
              <Label htmlFor="backlog-query-name">
                {format(messages.required, { label: messages.watchNameLabel })}
              </Label>
              <Input
                id="backlog-query-name"
                data-testid="backlog-query-name"
                value={name}
                aria-invalid={formError === "name" ? "true" : undefined}
                aria-describedby={formError === "name" ? "backlog-query-name-error" : undefined}
                onChange={(e: { target: { value: string } }) => setName(e.target.value)}
              />
              {fieldError("name", "backlog-query-name")}
            </div>
            <div className={FIELD}>
              <Label htmlFor="backlog-query-repo">
                {format(messages.required, { label: messages.watchRepoLabel })}
              </Label>
              <select
                id="backlog-query-repo"
                data-testid="backlog-query-repo"
                value={repo}
                aria-invalid={formError === "repo" ? "true" : undefined}
                aria-describedby={formError === "repo" ? "backlog-query-repo-error" : undefined}
                onChange={(e: { target: { value: string } }) => setRepo(e.target.value)}
              >
                <option value="">{messages.chooseRepository}</option>
                {repos.map((r) => (
                  <option key={r.value} value={r.value}>
                    {r.label}
                  </option>
                ))}
              </select>
              {fieldError("repo", "backlog-query-repo")}
            </div>
            <div className={ROW}>
              <Button type="button" data-testid="backlog-query-save" onClick={() => void saveQuery()}>
                {messages.save}
              </Button>
              <Button type="button" data-testid="backlog-query-cancel" onClick={() => setForm(false)}>
                {messages.cancel}
              </Button>
            </div>
          </div>
        ) : null}
        {result.kind === "loading" ? <p data-testid="backlog-query-loading">{messages.pageLoading}</p> : null}
        {result.kind === "failed" ? (
          <div data-testid="backlog-query-error" className={STACK}>
            <p>{t(result.notice)}</p>
            <Button type="button" data-testid="backlog-query-retry" onClick={() => void run(selected)}>
              {messages.retry}
            </Button>
          </div>
        ) : null}
        {result.kind === "ready" && result.rows.length === 0 ? (
          <p data-testid="backlog-query-empty">{messages.queryEmpty}</p>
        ) : null}
        {result.kind === "ready" && result.rows.length > 0 ? (
          <table data-testid="backlog-query-table">
            <thead>
              <tr>
                <th scope="col">{messages.colNumber}</th>
                <th scope="col">{messages.colTitle}</th>
                <th scope="col">{messages.colState}</th>
                <th scope="col">{messages.colTasks}</th>
              </tr>
            </thead>
            <tbody>
              {result.rows.map((r) => (
                <tr key={r.number} data-testid={`backlog-query-row-${r.number}`}>
                  <td>
                    <a
                      href={r.url}
                      data-testid={`backlog-query-pr-${r.number}`}
                      target="_blank"
                      rel="noreferrer"
                    >
                      {format(messages.prNumber, { number: r.number })}
                    </a>
                  </td>
                  <td>{r.title}</td>
                  <td>{stateText(r.state, messages)}</td>
                  <td>
                    {r.linkedTaskIds.map((id) => (
                      <a
                        key={id}
                        href={`/t/${encodeURIComponent(id)}`}
                        data-testid={`backlog-query-task-${id}`}
                      >
                        {format(messages.taskLink, { id })}
                      </a>
                    ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : null}
        {deleting ? (
          <ConfirmDialog
            testId="backlog-query-delete-dialog"
            title={messages.deleteQueryTitle}
            body={messages.deleteQueryBody}
            confirmLabel={messages.delete}
            onConfirm={remove}
            onClose={() => setDeleting(false)}
          />
        ) : null}
      </div>
    );
  };
}
