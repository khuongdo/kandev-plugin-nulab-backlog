import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { createPrToolbar } from "../git/pr-toolbar";
import { createSaveQueryDialog } from "../git/save-query-dialog";
import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON, FIELD, RESULTS, ROW, STACK } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { settingsHref } from "../page/BacklogPage";
import type { QuickAction } from "../page/quick-actions";
import { createStartTask } from "../page/start-task";
import { createSectionParts } from "../settings/section-parts";
import type { Notice } from "../settings/state";
import {
  issueNotice,
  issueQueryFilters,
  issuesFailure,
  openStatusIds,
  showingText,
  taskHref,
  type IssueItem,
  type IssueQuery,
  type IssuePage,
  type IssuesFailure,
  type TaskLink,
} from "./issues-state";
import { createLinkTaskDialog } from "./link-task-dialog";
import type { LinksStore } from "./links-store";

const PAGE_SIZE = 20;
/** The "All" choice of a filter: Radix Select items cannot have an empty value. */
const ALL = "all";
/** The "Not closed" status choice and the "Me" assignee choice (FR4.1, FR4.2). */
const OPEN = "open";
const ME = "me";
const SEARCH_WAIT = 400;

type Load =
  | { kind: "loading" }
  | { kind: "ready"; page: Required<Pick<IssuePage, "items">> & IssuePage }
  | IssuesFailure;

interface Option {
  id?: number;
  key?: string;
  name: string;
}

interface FilterOptions {
  projects?: Option[];
  statuses?: Option[];
  assignees?: Option[];
}

/** The filters, shaped like a saved issue query: assignee "", "me" or a user id. */
interface Filters {
  projectKey: string;
  statusIds: number[];
  assignee: string;
}

const NO_FILTERS: Filters = { projectKey: "", statusIds: [], assignee: "" };

/** The issues.list body: only the filters in use; "me" is resolved by the server. */
function listBody(page: number, keyword: string, f: Filters): Record<string, unknown> {
  const body: Record<string, unknown> = { page, pageSize: PAGE_SIZE, keyword };
  if (f.projectKey) body.projectKeys = [f.projectKey];
  if (f.statusIds.length) body.statusIds = f.statusIds;
  if (f.assignee === ME) body.assignee = ME;
  else if (f.assignee) body.assigneeIds = [Number(f.assignee)];
  return body;
}

/** The query the page opens on: a saved one, or the "Assigned to me, open" preset. */
export interface IssueSelection {
  /** Changes whenever the user picks again, so a pick re-applies. */
  key: string;
  query?: IssueQuery;
}

export interface IssuesPageProps {
  workspaceId: string;
  /** The workspace's issue quick actions for the "+ Task" menu (FR1.1). */
  quickActions?: QuickAction[];
  /** Without one the page opens with no filters. */
  selection?: IssueSelection;
  /** A query saved from the toolbar (FR4.3). */
  onSavedQuery?: (query: IssueQuery) => void;
  /** Opens the save dialog each time it changes (the scope bar's Save current). */
  saveRequest?: number;
}

/**
 * The issue list at /backlog once connected (M2, M2m; US2.1, US2.2, US2.3,
 * US4.2; FR1, FR5.3): filters, a search that waits 400 ms, 20 rows per page
 * in the host's change request rows, linked tasks, the "+ Task" quick action
 * menu and Link to task per row, and Refresh. Phones get a Filters (n) drawer.
 */
export function createIssuesPage(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): Component<IssuesPageProps> {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const ui = hostUi(host);
  const { Button, Input, Label, Skeleton, Pagination, PaginationContent, PaginationItem } = ui;
  const { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } = ui;
  const { ChangeRequestList, ChangeRequestRow } = ui;
  const { RowMenu } = createSectionParts(host, messages);
  const LinkTaskDialog = createLinkTaskDialog(host, messages);
  const StartTask = createStartTask(host, messages);
  const SaveQueryDialog = createSaveQueryDialog(host, messages);
  const ListToolbar = createPrToolbar(host, messages);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;
  const t = (n: Notice) => noticeText(n, messages);

  return function IssuesPage({
    workspaceId,
    quickActions = [],
    selection,
    onSavedQuery,
    saveRequest = 0,
  }: IssuesPageProps) {
    const { isMobile } = host.useResponsiveBreakpoint?.() ?? { isMobile: false };
    // Undefined until the selection is applied: the preset needs the statuses first.
    const [filters, setFilters] = useState<Filters | undefined>(selection ? undefined : NO_FILTERS);
    const [search, setSearch] = useState("");
    const [keyword, setKeyword] = useState("");
    const [page, setPage] = useState(1);
    const [nonce, setNonce] = useState(0);
    const [load, setLoad] = useState<Load>({ kind: "loading" });
    const [options, setOptions] = useState<FilterOptions | undefined>(undefined);
    const [saving, setSaving] = useState(false);
    const [announcement, setAnnouncement] = useState("");
    const [countdown, setCountdown] = useState<number | undefined>(undefined);
    const [linkDialog, setLinkDialog] = useState<IssueItem | undefined>(undefined);
    const [refreshing, setRefreshing] = useState(false);
    const [showFilters, setShowFilters] = useState(false);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);
    const latest = useRef(0);
    const focusList = useRef(false);
    const list = useRef<HTMLHeadingElement | null>(null);
    const applied = useRef("");

    useEffect(() => {
      host.api
        .invokeAction<FilterOptions>("issues.filters", { workspaceId })
        .then((o) => setOptions(o ?? {}))
        .catch(() => setOptions({})); // the list still works without filter choices
    }, [workspaceId]);

    // FR4.1, FR4.5: apply the selected saved query, or the preset once the statuses are known.
    useEffect(() => {
      if (!selection || applied.current === selection.key) return;
      const q = selection.query;
      if (!q && !options) return;
      applied.current = selection.key;
      setFilters(
        q
          ? { projectKey: q.projectKey ?? "", statusIds: q.statusIds ?? [], assignee: q.assignee ?? "" }
          : { projectKey: "", statusIds: openStatusIds(options?.statuses), assignee: ME },
      );
      setSearch(q?.keyword ?? "");
      setKeyword(q?.keyword ?? "");
      setPage(1);
    }, [selection?.key, options]);

    useEffect(() => {
      if (saveRequest > 0) setSaving(true);
    }, [saveRequest]);

    useEffect(() => {
      const id = setTimeout(() => {
        if (search.trim() !== keyword) {
          setKeyword(search.trim());
          setPage(1);
        }
      }, SEARCH_WAIT);
      return () => clearTimeout(id);
    }, [search]);

    useEffect(() => {
      if (!filters) return;
      const seq = ++latest.current;
      setLoad({ kind: "loading" });
      setAnnouncement(messages.issuesLoading);
      host.api
        .invokeAction<IssuePage>("issues.list", { workspaceId, body: listBody(page, keyword, filters) })
        .then((r) => {
          if (seq !== latest.current) return;
          const ready = { ...r, items: r?.items ?? [] };
          setLoad({ kind: "ready", page: ready });
          setAnnouncement(
            showingText({ page, pageSize: PAGE_SIZE, total: ready.total ?? 0 }, messages) ||
              messages.issuesEmpty,
          );
        })
        .catch((e) => {
          if (seq !== latest.current) return;
          const f = issuesFailure(e);
          setLoad(f);
          if (f.kind === "rate_limited") {
            setCountdown(f.retryAfterSeconds);
            setAnnouncement(t(f.notice)); // once; the visible countdown changes every second
          }
        });
    }, [workspaceId, page, keyword, filters, nonce]);

    useEffect(() => {
      if (load.kind === "ready" && focusList.current) {
        focusList.current = false;
        list.current?.focus();
      }
    }, [load]);

    const waiting = countdown !== undefined;
    useEffect(() => {
      if (!waiting) return;
      const id = setInterval(() => setCountdown((s) => (s === undefined ? s : s - 1)), 1000);
      return () => clearInterval(id);
    }, [waiting]);
    useEffect(() => {
      if (countdown !== undefined && countdown <= 0) {
        setCountdown(undefined);
        setNonce((n) => n + 1);
      }
    }, [countdown]);

    const reload = useCallback(() => setNonce((n) => n + 1), []);
    const current = filters ?? NO_FILTERS;
    const openIds = openStatusIds(options?.statuses);
    const setFilter = (next: Partial<Filters>) => {
      setFilters((f) => ({ ...(f ?? NO_FILTERS), ...next }));
      setPage(1);
    };
    // The status choice: All, Not closed (every status but Closed), or one status.
    const statusValue = (() => {
      const ids = current.statusIds;
      if (ids.length === 0) return ALL;
      if (ids.length === openIds.length && ids.every((id) => openIds.includes(id))) return OPEN;
      return ids.length === 1 ? String(ids[0]) : "";
    })();
    const reset = () => {
      setFilters(NO_FILTERS);
      setSearch("");
      setKeyword("");
      setPage(1);
      reload();
    };
    const goTo = (p: number) => {
      focusList.current = true;
      setPage(p);
    };

    const addTask = (key: string, task: TaskLink) => {
      setLoad((l) =>
        l.kind !== "ready"
          ? l
          : {
              ...l,
              page: {
                ...l.page,
                items: l.page.items.map((i) =>
                  i.issueKey === key ? { ...i, linkedTasks: [...i.linkedTasks, task] } : i,
                ),
              },
            },
      );
      if (store) void store.refresh(workspaceId);
    };

    const refresh = async () => {
      setRefreshing(true);
      setNotice(undefined);
      try {
        await host.api.invokeAction("issues.refresh", { workspaceId });
      } catch (e) {
        setNotice(issueNotice(e));
      }
      setRefreshing(false);
      reload();
    };

    const filterCount = [current.projectKey, current.statusIds.length > 0, current.assignee].filter(
      Boolean,
    ).length;
    const select = (
      label: string,
      value: string,
      onChange: (v: string) => void,
      choices: Option[] | undefined,
      testId: string,
      extra?: { value: string; label: string },
    ) => (
      <div className={FIELD}>
        <Label htmlFor={testId}>{label}</Label>
        <Select value={value} onValueChange={onChange}>
          <SelectTrigger id={testId} data-testid={testId} className="min-w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL} data-testid={`${testId}-${ALL}`}>
              {messages.filterAll}
            </SelectItem>
            {extra ? (
              <SelectItem value={extra.value} data-testid={`${testId}-${extra.value}`}>
                {extra.label}
              </SelectItem>
            ) : null}
            {(choices ?? []).map((o) => (
              <SelectItem
                key={o.key ?? o.id}
                value={o.key ?? String(o.id)}
                data-testid={`${testId}-${o.key ?? o.id}`}
              >
                {o.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    );
    const filterFields = (
      <div className={ROW}>
        {select(
          messages.filterProject,
          current.projectKey || ALL,
          (v) => setFilter({ projectKey: v === ALL ? "" : v }),
          options?.projects,
          "backlog-issues-project",
        )}
        {select(
          messages.filterStatus,
          statusValue,
          (v) => setFilter({ statusIds: v === ALL ? [] : v === OPEN ? openIds : [Number(v)] }),
          options?.statuses,
          "backlog-issues-status",
          { value: OPEN, label: messages.statusNotClosed },
        )}
        {select(
          messages.filterAssignee,
          current.assignee || ALL,
          (v) => setFilter({ assignee: v === ALL ? "" : v }),
          options?.assignees,
          "backlog-issues-assignee",
          { value: ME, label: messages.whoMe },
        )}
      </div>
    );

    const tasksOf = (item: IssueItem) => (
      <span className={ROW}>
        {item.linkedTasks.map((task) => (
          <a
            key={task.taskId}
            href={taskHref(task.taskId)}
            data-testid={`backlog-issue-task-${item.issueKey}-${task.taskId}`}
            onClick={(e: { preventDefault(): void }) => {
              e.preventDefault();
              host.navigate(taskHref(task.taskId));
            }}
          >
            {task.taskKey ?? task.taskId}
          </a>
        ))}
      </span>
    );

    const actionsOf = (item: IssueItem) => {
      const key = item.issueKey;
      return (
        <span className={ROW}>
          <StartTask
            workspaceId={workspaceId}
            kind="issue"
            title={item.summary}
            url={item.url}
            issueKey={key}
            actions={quickActions}
            testId={`backlog-issue-${key}`}
            onLinked={(taskId: string, taskKey?: string) => addTask(key, { taskId, taskKey })}
          />
          <RowMenu
            testId={`backlog-issue-menu-${key}`}
            label={format(messages.moreActions, { key })}
            items={[
              {
                testId: `backlog-issue-link-${key}`,
                label: messages.linkToTask,
                onSelect: () => setLinkDialog(item),
              },
            ]}
          />
        </span>
      );
    };

    const row = (item: IssueItem) => (
      <ChangeRequestRow
        key={item.issueKey}
        testId={`backlog-issue-row-${item.issueKey}`}
        stateIcon={icon(host, "issue", "h-4 w-4 text-muted-foreground")}
        title={item.summary}
        href={item.url}
        metadata={[
          <span key="k">{item.issueKey}</span>,
          <span key="s">{item.status}</span>,
          item.assignee ? <span key="a">{item.assignee}</span> : null,
          item.updatedAt ? <span key="u">{relative(item.updatedAt)}</span> : null,
        ]}
        taskIndicator={tasksOf(item)}
        action={actionsOf(item)}
      />
    );

    const body = (() => {
      switch (load.kind) {
        case "loading":
          return (
            <div data-testid="backlog-issues-loading" className={STACK}>
              {[0, 1, 2, 3, 4].map((i) => (
                <Skeleton key={i} className="h-6 w-full" />
              ))}
            </div>
          );
        case "not_connected":
        case "no_project":
        case "sign_in_again": {
          const href = settingsHref(workspaceId);
          const text = {
            not_connected: messages.pageNotConnected,
            no_project: messages.issuesNoProject,
            sign_in_again: messages.issuesSignInAgain,
          }[load.kind];
          return (
            <div className={STACK}>
              <p>{text}</p>
              <a
                href={href}
                data-testid="backlog-issues-settings-link"
                onClick={(e: { preventDefault(): void }) => {
                  e.preventDefault();
                  host.navigate(href);
                }}
              >
                {messages.openSettings}
              </a>
            </div>
          );
        }
        case "rate_limited":
          return (
            <p data-testid="backlog-issues-wait">
              {t({ key: "rateLimitedRetrying", params: { seconds: countdown ?? load.retryAfterSeconds } })}
            </p>
          );
        case "error":
          return (
            <div data-testid="backlog-issues-error" className={STACK}>
              <p>{messages.issuesLoadFailed}</p>
              <p>{t(load.notice)}</p>
              <div>
                <Button
                  type="button"
                  variant="outline"
                  className={BUTTON}
                  data-testid="backlog-issues-retry"
                  onClick={reload}
                >
                  {messages.retry}
                </Button>
              </div>
            </div>
          );
      }
      const items = load.page.items;
      if (items.length === 0) {
        return (
          <div data-testid="backlog-issues-empty" className={STACK}>
            <p>{messages.issuesEmpty}</p>
            <div>
              <Button
                type="button"
                variant="outline"
                className={BUTTON}
                data-testid="backlog-issues-reset"
                onClick={reset}
              >
                {messages.resetFilters}
              </Button>
            </div>
          </div>
        );
      }
      const total = load.page.total ?? items.length;
      const showing = showingText({ page, pageSize: PAGE_SIZE, total }, messages);
      return (
        <div className={STACK}>
          <ChangeRequestList loading={false} error={null} emptyMessage={messages.issuesEmpty} isEmpty={false}>
            {items.map(row)}
          </ChangeRequestList>
          <div className={ROW}>
            <p data-testid="backlog-issues-showing">{showing}</p>
            <Pagination className="mx-0 w-auto">
              <PaginationContent>
                <PaginationItem>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className={BUTTON}
                    data-testid="backlog-issues-prev"
                    disabled={page <= 1}
                    onClick={() => goTo(page - 1)}
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
                    data-testid="backlog-issues-next"
                    disabled={page * PAGE_SIZE >= total}
                    onClick={() => goTo(page + 1)}
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

    const refreshedAt = load.kind === "ready" ? load.page.refreshedAt : undefined;
    return (
      <div data-testid="backlog-issues" className="flex min-w-0 flex-col">
        <ListToolbar
          idPrefix="backlog-issues"
          title={messages.issuesListLabel}
          headingRef={list}
          count={load.kind === "ready" ? load.page.total : undefined}
          loading={refreshing}
          lastFetchedAt={refreshedAt}
          refreshLabel={refreshing ? messages.refreshing : messages.refresh}
          onRefresh={() => void refresh()}
        >
          <div className={FIELD}>
            <Label htmlFor="backlog-issues-search">{messages.searchIssues}</Label>
            <Input
              id="backlog-issues-search"
              data-testid="backlog-issues-search"
              type="search"
              value={search}
              onChange={(e: { target: { value: string } }) => setSearch(e.target.value)}
            />
          </div>
          {isMobile ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={`${BUTTON} self-end`}
              data-testid="backlog-issues-filters-toggle"
              aria-expanded={showFilters}
              aria-controls="backlog-issues-filters"
              onClick={() => setShowFilters((s) => !s)}
            >
              {format(messages.filtersButton, { count: filterCount })}
            </Button>
          ) : null}
          {!isMobile || showFilters ? <div id="backlog-issues-filters">{filterFields}</div> : null}
          <Button
            type="button"
            variant="outline"
            size="sm"
            className={`${BUTTON} self-end`}
            data-testid="backlog-issues-save-query"
            disabled={!filters}
            onClick={() => setSaving(true)}
          >
            {messages.saveQuery}
          </Button>
        </ListToolbar>
        <div className={RESULTS} data-testid="backlog-issues-results">
          {body}
          {notice ? (
            <p role="alert" data-testid="backlog-issues-notice">
              {t(notice)}
            </p>
          ) : null}
        </div>
        <div role="status" aria-live="polite" className="sr-only" data-testid="backlog-issues-announcer">
          {announcement}
        </div>
        {saving
          ? (() => {
              const draft: IssueQuery = {
                name: "",
                projectKey: current.projectKey,
                statusIds: current.statusIds,
                assignee: current.assignee,
                keyword,
              };
              return (
                <SaveQueryDialog
                  workspaceId={workspaceId}
                  query={draft}
                  action="issues.queries.save"
                  description={issueQueryFilters(draft, messages)}
                  onSaved={(q: IssueQuery) => {
                    onSavedQuery?.(q);
                    setSaving(false);
                  }}
                  onClose={() => setSaving(false)}
                />
              );
            })()
          : null}
        {linkDialog ? (
          <LinkTaskDialog
            workspaceId={workspaceId}
            issueKey={linkDialog.issueKey}
            onLinked={(task: TaskLink) => addTask(linkDialog.issueKey, task)}
            onClose={() => setLinkDialog(undefined)}
          />
        ) : null}
      </div>
    );
  };
}
