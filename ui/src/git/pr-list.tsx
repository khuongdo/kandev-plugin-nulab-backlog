import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { showingText, taskHref } from "../issues/issues-state";
import { BUTTON, FIELD, RESULTS, ROW, STACK } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import type { QuickAction } from "../page/quick-actions";
import { createStartTask } from "../page/start-task";
import type { Notice } from "../settings/state";
import {
  gitNotice,
  loadProviders,
  loadRepoOptions,
  noticeText,
  providerName,
  splitRepo,
  stateText,
  usableProviders,
  type ProviderView,
  type RepoOption,
} from "./git-state";
import { createPrToolbar } from "./pr-toolbar";
import { createScmPrList } from "./scm-pr-list";
import { createSaveQueryDialog, type Query } from "./save-query-dialog";

/** One row of git.prs.list (FR4.2). */
interface PullRequestRow {
  number: number;
  title: string;
  status: string;
  author: string;
  assignee?: string;
  updated: string;
  url: string;
  linkedTaskIds: string[];
}

interface PullRequestPage {
  items?: PullRequestRow[];
  page: number;
  pageSize: number;
  total: number;
  hasNext: boolean;
}

interface Filters {
  repo: string; // "PROJ/repo"; "" until one is chosen
  statuses: string[];
  assignee: string;
  creator: string;
}

type Load =
  | { kind: "idle" | "loading" }
  | { kind: "failed"; notice: Notice }
  | { kind: "ready"; page: PullRequestPage; at: string };

const PAGE_SIZE = 20;
const STATUSES = ["open", "closed", "merged"] as const;
const STATUS_KEYS: Record<string, MessageKey> = {
  open: "stateOpen",
  closed: "stateClosed",
  merged: "stateMerged",
};
const ICONS: Record<string, string> = {
  open: "pull-request",
  merged: "merged",
  closed: "pull-request-closed",
};
const START: Filters = { repo: "", statuses: ["open"], assignee: "anyone", creator: "anyone" };

/** The query the list opens on: a saved one, or the "Open, assigned to me" preset. */
export interface PrSelection {
  /** Changes whenever the user picks again, so a pick re-applies. */
  key: string;
  query?: Query;
}

export interface PrListProps {
  workspaceId: string;
  /** The workspace's PR quick actions for the "+ Task" menu (FR1.1). */
  quickActions?: QuickAction[];
  /** Without one the list waits for a repository to be chosen. */
  selection?: PrSelection;
  /** A query saved from the toolbar (BR2.5). */
  onSavedQuery?: (query: Query) => void;
  /** Opens the save dialog each time it changes (the scope bar's Save current). */
  saveRequest?: number;
  /** The selected Backlog projects: other providers list their mapped repositories (FR3.3). */
  selectedProjects?: string[];
}

/**
 * The Pull requests scope of /backlog (WF4, FR2.5, FR2.6, FR3): the PRs of
 * one repository, 20 per page with filters, opened on the selected saved
 * query or the "Open, assigned to me" preset over the first repository, and
 * Save query from the current filters. Rows use the host's change request list.
 */
export function createPrList(host: PluginHostApi, messages: Messages = en): Component<PrListProps> {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const ui = hostUi(host);
  const {
    Alert,
    AlertDescription,
    Button,
    Checkbox,
    ChangeRequestList,
    ChangeRequestRow,
    Empty,
    EmptyHeader,
  } = ui;
  const { EmptyTitle, IntegrationIcon, IntegrationRepositoryFilter, Label, Pagination, PaginationContent } =
    ui;
  const { PaginationItem, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } = ui;
  const PrToolbar = createPrToolbar(host, messages);
  const SaveQueryDialog = createSaveQueryDialog(host, messages);
  const StartTask = createStartTask(host, messages);
  const ScmPrList = createScmPrList(host, messages);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;

  return function PrList({
    workspaceId,
    quickActions = [],
    selection,
    onSavedQuery,
    saveRequest = 0,
    selectedProjects = [],
  }: PrListProps) {
    const [filters, setFilters] = useState<Filters>(START);
    const [page, setPage] = useState(1);
    const [nonce, setNonce] = useState(0);
    const [load, setLoad] = useState<Load>({ kind: "idle" });
    const [repos, setRepos] = useState<RepoOption[] | undefined>(undefined);
    const [saving, setSaving] = useState(false);
    const latest = useRef(0);
    const applied = useRef("");
    // Intent 261007-source-control-agnostic (FR4.1): Backlog Git or a connected provider.
    const [provider, setProvider] = useState("backlog");
    const [providers, setProviders] = useState<ProviderView[]>([]);

    useEffect(() => {
      loadRepoOptions(host, workspaceId, messages).then(setRepos, () => setRepos([]));
      void loadProviders(host, workspaceId).then((v) => setProviders(usableProviders(v)));
    }, [workspaceId]);

    // FR3.1, FR3.2: apply the selected saved query, or the preset once the repositories are known.
    useEffect(() => {
      if (!selection || applied.current === selection.key) return;
      const q = selection.query;
      if (!q && !repos) return;
      applied.current = selection.key;
      setFilters(
        q
          ? {
              repo: `${q.projectKey}/${q.repoName}`,
              statuses: q.statuses,
              assignee: q.assignee,
              creator: q.creator ?? "anyone",
            }
          : { repo: repos?.[0]?.value ?? "", statuses: ["open"], assignee: "me", creator: "anyone" },
      );
      setPage(1); // BR2.4
    }, [selection?.key, repos]);

    useEffect(() => {
      if (saveRequest > 0 && filters.repo) setSaving(true);
    }, [saveRequest]);

    useEffect(() => {
      if (!filters.repo) return;
      const seq = ++latest.current;
      setLoad({ kind: "loading" });
      host.api
        .invokeAction<PullRequestPage>("git.prs.list", {
          workspaceId,
          body: {
            ...splitRepo(filters.repo),
            statuses: filters.statuses,
            assignee: filters.assignee,
            creator: filters.creator,
            page,
          },
        })
        .then(
          (r) => seq === latest.current && setLoad({ kind: "ready", page: r, at: new Date().toISOString() }),
        )
        .catch((e: unknown) => seq === latest.current && setLoad({ kind: "failed", notice: gitNotice(e) }));
    }, [workspaceId, filters, page, nonce]);

    const reload = useCallback(() => setNonce((n) => n + 1), []);
    const addTask = (number: number, taskId: string) =>
      setLoad((l) =>
        l.kind !== "ready"
          ? l
          : {
              ...l,
              page: {
                ...l.page,
                items: l.page.items?.map((pr) =>
                  pr.number === number ? { ...pr, linkedTaskIds: [...pr.linkedTaskIds, taskId] } : pr,
                ),
              },
            },
      );
    const change = (next: Partial<Filters>) => {
      setFilters((f) => ({ ...f, ...next }));
      setPage(1);
    };
    const toggleStatus = (s: string) =>
      change({
        statuses: filters.statuses.includes(s)
          ? filters.statuses.filter((x) => x !== s)
          : [...filters.statuses, s],
      });

    const providerControl = (
      <div className={FIELD}>
        <Label htmlFor="backlog-prs-provider">{messages.scmProviderLabel}</Label>
        <Select value={provider} onValueChange={setProvider}>
          <SelectTrigger id="backlog-prs-provider" data-testid="backlog-prs-provider" className="min-w-28">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {["backlog", ...providers.map((v) => v.provider)].map((p) => (
              <SelectItem key={p} value={p} data-testid={`backlog-prs-provider-${p}`}>
                {providerName(p, messages)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    );

    if (provider !== "backlog") {
      return (
        <div data-testid="backlog-prs" className="flex min-w-0 flex-col">
          <ScmPrList
            key={provider}
            workspaceId={workspaceId}
            provider={provider as ProviderView["provider"]}
            view={providers.find((v) => v.provider === provider)}
            selectedProjects={selectedProjects}
            quickActions={quickActions}
            providerControl={providerControl}
          />
        </div>
      );
    }

    const who = (field: "assignee" | "creator", label: string) => (
      <div className={FIELD}>
        <Label htmlFor={`backlog-prs-${field}`}>{label}</Label>
        <Select value={filters[field]} onValueChange={(v: string) => change({ [field]: v })}>
          <SelectTrigger
            id={`backlog-prs-${field}`}
            data-testid={`backlog-prs-${field}`}
            className="min-w-28"
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="anyone" data-testid={`backlog-prs-${field}-anyone`}>
              {messages.whoAnyone}
            </SelectItem>
            <SelectItem value="me" data-testid={`backlog-prs-${field}-me`}>
              {messages.whoMe}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    );

    const filterControls = [
      <div key="provider">{providerControl}</div>,
      <div key="repo" className={FIELD}>
        <span className="text-sm font-medium">{messages.colRepository}</span>
        <IntegrationRepositoryFilter
          value={filters.repo}
          onValueChange={(v: string) => change({ repo: v })}
          options={repos ?? []}
          ariaLabel={messages.colRepository}
          allLabel={messages.allRepositories}
          testId="backlog-prs-repo"
        />
      </div>,
      <div key="status" role="group" aria-labelledby="backlog-prs-status-label" className={FIELD}>
        <span id="backlog-prs-status-label" className="text-sm font-medium">
          {messages.watchStatusLegend}
        </span>
        <div className={ROW}>
          {STATUSES.map((s) => (
            <div key={s} className="flex items-center gap-1">
              <Checkbox
                id={`backlog-prs-status-${s}`}
                data-testid={`backlog-prs-status-${s}`}
                checked={filters.statuses.includes(s)}
                onCheckedChange={() => toggleStatus(s)}
              />
              <Label htmlFor={`backlog-prs-status-${s}`}>{messages[STATUS_KEYS[s]!]}</Label>
            </div>
          ))}
        </div>
      </div>,
      <div key="assignee">{who("assignee", messages.assigneeLabel)}</div>,
      <div key="creator">{who("creator", messages.creatorLabel)}</div>,
      <div key="save" className="self-end">
        <Button
          type="button"
          variant="outline"
          size="sm"
          className={BUTTON}
          data-testid="backlog-prs-save-query"
          disabled={!filters.repo}
          onClick={() => setSaving(true)}
        >
          {messages.saveQuery}
        </Button>
      </div>,
    ];

    const empty = (title: string) => (
      <Empty data-testid="backlog-prs-empty">
        <EmptyHeader>
          <EmptyTitle>{title}</EmptyTitle>
        </EmptyHeader>
      </Empty>
    );

    const row = (pr: PullRequestRow) => (
      <ChangeRequestRow
        key={pr.number}
        testId={`backlog-pr-row-${pr.number}`}
        stateIcon={<IntegrationIcon name={ICONS[pr.status] ?? "pull-request"} className="h-4 w-4" />}
        title={pr.title}
        href={pr.url}
        metadata={[
          <span key="n">{format(messages.prNumber, { number: pr.number })}</span>,
          <span key="s">{stateText(pr.status, messages)}</span>,
          pr.author ? <span key="a">{format(messages.byAuthor, { name: pr.author })}</span> : null,
          pr.assignee ? <span key="as">{format(messages.assignedTo, { name: pr.assignee })}</span> : null,
          pr.updated ? (
            <span key="u">{format(messages.updatedRelative, { time: relative(pr.updated) })}</span>
          ) : null,
        ]}
        taskIndicator={pr.linkedTaskIds.map((id) => (
          <a
            key={id}
            href={taskHref(id)}
            data-testid={`backlog-pr-task-${pr.number}-${id}`}
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
            title={pr.title}
            url={pr.url}
            actions={quickActions}
            testId={`backlog-pr-${pr.number}`}
            onLinked={(taskId: string) => addTask(pr.number, taskId)}
          />
        }
      />
    );

    const body = (() => {
      if (!filters.repo) return empty(messages.prsChooseRepository);
      if (load.kind === "failed") {
        return (
          <Alert data-testid="backlog-prs-error" variant="destructive">
            <AlertDescription>
              {messages.prsLoadFailed} {noticeText(load.notice, messages)}
            </AlertDescription>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={BUTTON}
              data-testid="backlog-prs-retry"
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
      return (
        <div className={STACK}>
          <ChangeRequestList loading={false} error={null} emptyMessage={messages.prsEmpty} isEmpty={false}>
            {items.map(row)}
          </ChangeRequestList>
          <div className={ROW}>
            <p data-testid="backlog-prs-showing">
              {showingText({ page, pageSize: PAGE_SIZE, total: load.page.total }, messages)}
            </p>
            <Pagination className="mx-0 w-auto">
              <PaginationContent>
                <PaginationItem>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className={BUTTON}
                    data-testid="backlog-prs-prev"
                    disabled={page <= 1}
                    onClick={() => setPage((p) => p - 1)}
                  >
                    {messages.previous}
                  </Button>
                </PaginationItem>
                <PaginationItem>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className={BUTTON}
                    data-testid="backlog-prs-next"
                    disabled={!load.page.hasNext}
                    onClick={() => setPage((p) => p + 1)}
                  >
                    {messages.next}
                  </Button>
                </PaginationItem>
              </PaginationContent>
            </Pagination>
          </div>
        </div>
      );
    })();

    return (
      <div data-testid="backlog-prs" className="flex min-w-0 flex-col">
        <PrToolbar
          title={messages.prsTitle}
          count={load.kind === "ready" && filters.repo ? load.page.total : undefined}
          loading={load.kind === "loading"}
          lastFetchedAt={load.kind === "ready" ? load.at : undefined}
          onRefresh={reload}
          refreshDisabled={!filters.repo}
        >
          {filterControls}
        </PrToolbar>
        <div className={RESULTS} data-testid="backlog-prs-results">
          {body}
        </div>
        {saving ? (
          <SaveQueryDialog
            workspaceId={workspaceId}
            query={{
              name: "",
              ...splitRepo(filters.repo),
              statuses: filters.statuses,
              assignee: filters.assignee,
              creator: filters.creator,
            }}
            onSaved={(q: Query) => {
              onSavedQuery?.(q);
              setSaving(false);
            }}
            onClose={() => setSaving(false)}
          />
        ) : null}
      </div>
    );
  };
}
