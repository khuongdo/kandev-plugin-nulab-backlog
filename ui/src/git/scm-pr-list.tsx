import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { taskHref } from "../issues/issues-state";
import { BUTTON, FIELD, RESULTS, ROW, STACK } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import type { QuickAction } from "../page/quick-actions";
import { createStartTask } from "../page/start-task";
import type { Notice } from "../settings/state";
import {
  noticeText,
  scmNotice,
  scmRepoOptions,
  splitScmRepo,
  stateText,
  type ProviderView,
  type ScmProvider,
} from "./git-state";
import { createPrToolbar } from "./pr-toolbar";
import { createSaveQueryDialog, scmQueryFilters, type ScmQuery } from "./save-query-dialog";

/** One row of scm.prs.list (FR4.1). */
interface Row {
  number: number;
  title: string;
  state: string;
  author: string;
  sourceBranch: string;
  targetBranch: string;
  updatedAt: string;
  url: string;
  linkedTaskIds: string[];
}

interface Page {
  items?: Row[];
  hasNext: boolean;
}

interface Filters {
  repo: string; // "PROJ:owner/name"
  statuses: string[];
  author: string;
}

type Load =
  | { kind: "idle" | "loading" }
  | { kind: "failed"; notice: Notice }
  | { kind: "ready"; page: Page; at: string };

const STATUSES = ["open", "closed", "merged"] as const;
const STATUS_KEYS: Record<string, MessageKey> = {
  open: "stateOpen",
  closed: "stateClosed",
  merged: "stateMerged",
};
const ICONS: Record<string, string> = {
  merged: "merged",
  closed: "pull-request-closed",
  declined: "pull-request-closed",
};

export interface ScmPrListProps {
  /** React key: one list per provider. */
  key?: string;
  workspaceId: string;
  provider: ScmProvider;
  view?: ProviderView;
  selectedProjects: string[];
  quickActions?: QuickAction[];
  /** The provider selector, shown first in the toolbar. */
  providerControl?: unknown;
}

/**
 * The pull requests of a GitHub, GitLab or Bitbucket repository mapped to a
 * selected Backlog project (FR3.3, FR4.1): 20 per page with status and
 * author filters, the provider's saved queries with its default applied
 * (FR4.2), and "+ Task" linking the new task to the pull request (FR5.1).
 */
export function createScmPrList(host: PluginHostApi, messages: Messages = en): Component<ScmPrListProps> {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const ui = hostUi(host);
  const { Alert, AlertDescription, Button, Checkbox, ChangeRequestList, ChangeRequestRow } = ui;
  const { Empty, EmptyHeader, EmptyTitle, IntegrationIcon, IntegrationRepositoryFilter, Label } = ui;
  const { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } = ui;
  const PrToolbar = createPrToolbar(host, messages);
  const SaveQueryDialog = createSaveQueryDialog(host, messages);
  const StartTask = createStartTask(host, messages);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;
  const ID = "backlog-scm-prs";

  return function ScmPrList({
    workspaceId,
    provider,
    view,
    selectedProjects,
    quickActions = [],
    providerControl,
  }: ScmPrListProps) {
    const repos = scmRepoOptions(view, selectedProjects, messages);
    const [saved, setSaved] = useState<ScmQuery[] | undefined>(undefined);
    const [savedId, setSavedId] = useState("");
    const [filters, setFilters] = useState<Filters | undefined>(undefined);
    const [page, setPage] = useState(1);
    const [nonce, setNonce] = useState(0);
    const [load, setLoad] = useState<Load>({ kind: "idle" });
    const [saving, setSaving] = useState(false);
    const latest = useRef(0);

    const apply = (q: ScmQuery) => {
      setSavedId(q.id ?? "");
      setFilters({ repo: `${q.projectKey}:${q.repo}`, statuses: q.statuses, author: q.author || "anyone" });
      setPage(1);
    };

    // FR4.2: open on the provider's default saved query, else the first repository.
    useEffect(() => {
      host.api
        .invokeAction<{ queries?: ScmQuery[] }>("scm.queries.list", { workspaceId })
        .then(
          (r) => r?.queries ?? [],
          () => [],
        )
        .then((all) => {
          const mine = all.filter((q) => q.provider === provider && !q.unmapped);
          setSaved(mine);
          const star = mine.find(
            (q) => q.isDefault && repos.some((o) => o.value === `${q.projectKey}:${q.repo}`),
          );
          if (star) apply(star);
          else setFilters({ repo: repos[0]?.value ?? "", statuses: ["open"], author: "anyone" });
        });
    }, [workspaceId, provider]);

    useEffect(() => {
      if (!filters?.repo) return;
      const seq = ++latest.current;
      setLoad({ kind: "loading" });
      const { projectKey, repo } = splitScmRepo(filters.repo);
      host.api
        .invokeAction<Page>("scm.prs.list", {
          workspaceId,
          body: { provider, projectKey, repo, statuses: filters.statuses, author: filters.author, page },
        })
        .then(
          (r) => seq === latest.current && setLoad({ kind: "ready", page: r, at: new Date().toISOString() }),
        )
        .catch((e: unknown) => seq === latest.current && setLoad({ kind: "failed", notice: scmNotice(e) }));
    }, [filters, page, nonce]);

    const reload = useCallback(() => setNonce((n) => n + 1), []);
    const change = (next: Partial<Filters>) => {
      setFilters((f) => (f ? { ...f, ...next } : f));
      setSavedId("");
      setPage(1);
    };
    const addTask = (number: number, taskId: string) =>
      setLoad((l) =>
        l.kind !== "ready"
          ? l
          : {
              ...l,
              page: {
                ...l.page,
                items: l.page.items?.map((r) =>
                  r.number === number ? { ...r, linkedTaskIds: [...r.linkedTaskIds, taskId] } : r,
                ),
              },
            },
      );
    const current = (): ScmQuery | undefined => {
      if (!filters?.repo) return undefined;
      const { projectKey, repo } = splitScmRepo(filters.repo);
      return { name: "", provider, projectKey, repo, statuses: filters.statuses, author: filters.author };
    };

    const select = (
      id: string,
      label: string,
      value: string,
      set: (v: string) => void,
      items: [string, string][],
    ) => (
      <div className={FIELD}>
        <Label htmlFor={id}>{label}</Label>
        <Select value={value} onValueChange={set}>
          <SelectTrigger id={id} data-testid={id} className="min-w-28">
            <SelectValue placeholder={messages.scmSavedNone} />
          </SelectTrigger>
          <SelectContent>
            {items.map(([v, l]) => (
              <SelectItem key={v} value={v} data-testid={`${id}-${v}`}>
                {l}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    );

    const controls = filters
      ? [
          <div key="provider">{providerControl}</div>,
          <div key="repo" className={FIELD}>
            <span className="text-sm font-medium">{messages.colRepository}</span>
            <IntegrationRepositoryFilter
              value={filters.repo}
              onValueChange={(v: string) => change({ repo: v })}
              options={repos}
              ariaLabel={messages.colRepository}
              allLabel={messages.allRepositories}
              testId={`${ID}-repo`}
            />
          </div>,
          <div key="saved">
            {select(
              `${ID}-saved`,
              messages.scmSavedLabel,
              savedId,
              (id) => {
                const q = saved?.find((x) => x.id === id);
                if (q) apply(q);
              },
              (saved ?? []).map((q) => [q.id ?? "", q.name]),
            )}
          </div>,
          <div key="status" role="group" aria-labelledby={`${ID}-status-label`} className={FIELD}>
            <span id={`${ID}-status-label`} className="text-sm font-medium">
              {messages.watchStatusLegend}
            </span>
            <div className={ROW}>
              {STATUSES.map((s) => (
                <div key={s} className="flex items-center gap-1">
                  <Checkbox
                    id={`${ID}-status-${s}`}
                    data-testid={`${ID}-status-${s}`}
                    checked={filters.statuses.includes(s)}
                    onCheckedChange={() =>
                      change({
                        statuses: filters.statuses.includes(s)
                          ? filters.statuses.filter((x) => x !== s)
                          : [...filters.statuses, s],
                      })
                    }
                  />
                  <Label htmlFor={`${ID}-status-${s}`}>{messages[STATUS_KEYS[s]!]}</Label>
                </div>
              ))}
            </div>
          </div>,
          <div key="author">
            {select(`${ID}-author`, messages.scmAuthorLabel, filters.author, (v) => change({ author: v }), [
              ["anyone", messages.whoAnyone],
              ["me", messages.whoMe],
            ])}
          </div>,
          <div key="save" className="self-end">
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={BUTTON}
              data-testid={`${ID}-save-query`}
              disabled={!filters.repo}
              onClick={() => setSaving(true)}
            >
              {messages.saveQuery}
            </Button>
          </div>,
        ]
      : [<div key="provider">{providerControl}</div>];

    const empty = (title: string) => (
      <Empty data-testid={`${ID}-empty`}>
        <EmptyHeader>
          <EmptyTitle>{title}</EmptyTitle>
        </EmptyHeader>
      </Empty>
    );

    const row = (r: Row) => (
      <ChangeRequestRow
        key={r.number}
        testId={`backlog-scm-pr-row-${r.number}`}
        stateIcon={<IntegrationIcon name={ICONS[r.state] ?? "pull-request"} className="h-4 w-4" />}
        title={r.title}
        href={r.url}
        metadata={[
          <span key="n">{format(messages.prNumber, { number: r.number })}</span>,
          <span key="s">{stateText(r.state, messages)}</span>,
          r.author ? <span key="a">{format(messages.byAuthor, { name: r.author })}</span> : null,
          <span key="b">
            {format(messages.scmBranches, { source: r.sourceBranch, target: r.targetBranch })}
          </span>,
          r.updatedAt ? (
            <span key="u">{format(messages.updatedRelative, { time: relative(r.updatedAt) })}</span>
          ) : null,
        ]}
        taskIndicator={r.linkedTaskIds.map((id) => (
          <a
            key={id}
            href={taskHref(id)}
            data-testid={`backlog-scm-pr-task-${r.number}-${id}`}
            onClick={(e: { preventDefault(): void }) => {
              e.preventDefault();
              host.navigate(taskHref(id));
            }}
          >
            {format(messages.taskLink, { id })}
          </a>
        ))}
        action={
          <StartTask
            workspaceId={workspaceId}
            kind="pr"
            title={r.title}
            url={r.url}
            actions={quickActions}
            testId={`backlog-scm-pr-${r.number}`}
            linkAction="scm.prs.link"
            onLinked={(taskId: string) => addTask(r.number, taskId)}
          />
        }
      />
    );

    const body = (() => {
      if (filters && repos.length === 0) return empty(messages.scmNoRepositories);
      if (load.kind === "failed") {
        return (
          <Alert data-testid={`${ID}-error`} variant="destructive">
            <AlertDescription>
              {messages.prsLoadFailed} {noticeText(load.notice, messages)}
            </AlertDescription>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={BUTTON}
              data-testid={`${ID}-retry`}
              onClick={reload}
            >
              {messages.retry}
            </Button>
          </Alert>
        );
      }
      if (load.kind !== "ready") {
        return <ChangeRequestList loading error={null} emptyMessage="" isEmpty={false} children={null} />;
      }
      const items = load.page.items ?? [];
      if (items.length === 0) return empty(messages.prsEmpty);
      const pager = (id: string, label: string, disabled: boolean, step: number) => (
        <Button
          type="button"
          variant="outline"
          size="sm"
          className={BUTTON}
          data-testid={`${ID}-${id}`}
          disabled={disabled}
          onClick={() => setPage((p) => p + step)}
        >
          {label}
        </Button>
      );
      return (
        <div className={STACK}>
          <ChangeRequestList loading={false} error={null} emptyMessage={messages.prsEmpty} isEmpty={false}>
            {items.map(row)}
          </ChangeRequestList>
          <div className={ROW}>
            {pager("prev", messages.previous, page <= 1, -1)}
            {pager("next", messages.next, !load.page.hasNext, 1)}
          </div>
        </div>
      );
    })();

    const query = current();
    return (
      <div data-testid={ID} className="flex min-w-0 flex-col">
        <PrToolbar
          title={messages.prsTitle}
          loading={load.kind === "loading"}
          lastFetchedAt={load.kind === "ready" ? load.at : undefined}
          onRefresh={reload}
          refreshDisabled={!filters?.repo}
          idPrefix={ID}
        >
          {controls}
        </PrToolbar>
        <div className={RESULTS} data-testid={`${ID}-results`}>
          {body}
        </div>
        {saving && query ? (
          <SaveQueryDialog
            workspaceId={workspaceId}
            query={query}
            action="scm.queries.save"
            description={scmQueryFilters(query, messages)}
            onSaved={(q: ScmQuery) => {
              setSaved((list) => [...(list ?? []), q]);
              setSaving(false);
            }}
            onClose={() => setSaving(false)}
          />
        ) : null}
      </div>
    );
  };
}
