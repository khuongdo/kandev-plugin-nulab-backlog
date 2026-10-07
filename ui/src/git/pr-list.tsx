import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { showingText, taskHref } from "../issues/issues-state";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import type { Notice } from "../settings/state";
import { gitNotice, loadRepoOptions, noticeText, splitRepo, stateText, type RepoOption } from "./git-state";
import { createPrToolbar } from "./pr-toolbar";
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

export interface PrListProps {
  workspaceId: string;
  /** The workspace's selected projects: a saved query of another project is disabled (BR2.4). */
  selectedProjects: string[];
}

/**
 * The Pull requests scope of /backlog (WF4, FR2.5, FR2.6): the PRs of one
 * repository, 20 per page with filters, saved queries as presets, and Save
 * query from the current filters. Rows use the host's change request list.
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
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;

  return function PrList({ workspaceId, selectedProjects }: PrListProps) {
    const [filters, setFilters] = useState<Filters>(START);
    const [page, setPage] = useState(1);
    const [preset, setPreset] = useState("");
    const [nonce, setNonce] = useState(0);
    const [load, setLoad] = useState<Load>({ kind: "idle" });
    const [repos, setRepos] = useState<RepoOption[]>([]);
    const [queries, setQueries] = useState<Query[]>([]);
    const [saving, setSaving] = useState(false);
    const latest = useRef(0);

    useEffect(() => {
      loadRepoOptions(host, workspaceId, messages).then(setRepos, () => setRepos([]));
      host.api.invokeAction<{ queries?: Query[] }>("git.queries.list", { workspaceId }).then(
        (r) => setQueries(r?.queries ?? []),
        () => setQueries([]),
      );
    }, [workspaceId]);

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
    const change = (next: Partial<Filters>) => {
      setFilters((f) => ({ ...f, ...next }));
      setPreset("");
      setPage(1);
    };
    const applyPreset = (id: string) => {
      const q = queries.find((x) => x.id === id);
      if (!q) return;
      setPreset(id);
      setFilters({
        repo: `${q.projectKey}/${q.repoName}`,
        statuses: q.statuses,
        assignee: q.assignee,
        creator: q.creator ?? "anyone",
      });
      setPage(1); // BR2.4
    };
    const toggleStatus = (s: string) =>
      change({
        statuses: filters.statuses.includes(s)
          ? filters.statuses.filter((x) => x !== s)
          : [...filters.statuses, s],
      });

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
      <div key="repo" className={FIELD}>
        <span className="text-sm font-medium">{messages.colRepository}</span>
        <IntegrationRepositoryFilter
          value={filters.repo}
          onValueChange={(v: string) => change({ repo: v })}
          options={repos}
          ariaLabel={messages.colRepository}
          allLabel={messages.allRepositories}
          testId="backlog-prs-repo"
        />
      </div>,
      <div key="preset" className={FIELD}>
        <Label htmlFor="backlog-prs-preset">{messages.queryLabel}</Label>
        <Select value={preset} onValueChange={applyPreset}>
          <SelectTrigger id="backlog-prs-preset" data-testid="backlog-prs-preset" className="min-w-40">
            <SelectValue placeholder={messages.chooseQuery} />
          </SelectTrigger>
          <SelectContent>
            {queries.map((q) => {
              const off = !selectedProjects.includes(q.projectKey);
              return (
                <SelectItem key={q.id} value={q.id} disabled={off} data-testid={`backlog-prs-preset-${q.id}`}>
                  {off ? `${q.name} (${messages.projectNotSelected})` : q.name}
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
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
      <div data-testid="backlog-prs" className={STACK}>
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
        {body}
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
              setQueries((list) => [...list, q]);
              setSaving(false);
            }}
            onClose={() => setSaving(false)}
          />
        ) : null}
      </div>
    );
  };
}
