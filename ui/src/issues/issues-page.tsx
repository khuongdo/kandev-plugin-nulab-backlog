import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { settingsHref } from "../page/BacklogPage";
import { createConfirmDialog } from "../settings/confirm-dialog";
import { createSectionParts } from "../settings/section-parts";
import type { Notice } from "../settings/state";
import {
  issueNotice,
  issuesFailure,
  rowLabel,
  showingText,
  taskHref,
  type IssueItem,
  type IssuePage,
  type IssuesFailure,
  type TaskLink,
} from "./issues-state";
import { createLinkTaskDialog } from "./link-task-dialog";
import type { LinksStore } from "./links-store";

const PAGE_SIZE = 20;
/** The "All" choice of a filter: Radix Select items cannot have an empty value. */
const ALL = "all";
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

interface Filters {
  projectKey: string;
  statusId: string;
  assigneeId: string;
}

const NO_FILTERS: Filters = { projectKey: "", statusId: "", assigneeId: "" };

/** The issues.list body: only the filters in use. */
function listBody(page: number, keyword: string, f: Filters): Record<string, unknown> {
  const body: Record<string, unknown> = { page, pageSize: PAGE_SIZE, keyword };
  if (f.projectKey) body.projectKeys = [f.projectKey];
  if (f.statusId) body.statusIds = [Number(f.statusId)];
  if (f.assigneeId) body.assigneeIds = [Number(f.assigneeId)];
  return body;
}

/**
 * The issue list at /backlog once connected (M2, M2m; US2.1, US2.2, US2.3,
 * US3.1, US4.2): filters, a search that waits 400 ms, 20 rows per page,
 * linked tasks, Create task and Link to task per row, and Refresh. Phones
 * get cards and a Filters (n) drawer.
 */
export function createIssuesPage(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): Component<{ workspaceId: string }> {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const ui = hostUi(host);
  const { Button, Input, Label, Skeleton, Pagination, PaginationContent, PaginationItem } = ui;
  const { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } = ui;
  const { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } = ui;
  const { RowMenu } = createSectionParts(host, messages);
  const ConfirmDialog = createConfirmDialog(host, messages);
  const LinkTaskDialog = createLinkTaskDialog(host, messages);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;
  const t = (n: Notice) => noticeText(n, messages);

  return function IssuesPage({ workspaceId }: { workspaceId: string }) {
    const { isMobile } = host.useResponsiveBreakpoint?.() ?? { isMobile: false };
    const [filters, setFilters] = useState<Filters>(NO_FILTERS);
    const [search, setSearch] = useState("");
    const [keyword, setKeyword] = useState("");
    const [page, setPage] = useState(1);
    const [nonce, setNonce] = useState(0);
    const [load, setLoad] = useState<Load>({ kind: "loading" });
    const [options, setOptions] = useState<FilterOptions>({});
    const [announcement, setAnnouncement] = useState("");
    const [countdown, setCountdown] = useState<number | undefined>(undefined);
    const [creating, setCreating] = useState<string[]>([]);
    const [created, setCreated] = useState<Record<string, string>>({});
    const [linkedDialog, setLinkedDialog] = useState<IssueItem | undefined>(undefined);
    const [linkDialog, setLinkDialog] = useState<IssueItem | undefined>(undefined);
    const [refreshing, setRefreshing] = useState(false);
    const [showFilters, setShowFilters] = useState(false);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);
    const latest = useRef(0);
    const focusList = useRef(false);
    const busy = useRef(new Set<string>());
    const list = useRef<HTMLHeadingElement | null>(null);

    useEffect(() => {
      host.api
        .invokeAction<FilterOptions>("issues.filters", { workspaceId })
        .then((o) => setOptions(o ?? {}))
        .catch(() => setOptions({})); // the list still works without filter choices
    }, [workspaceId]);

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
    const setFilter = (name: keyof Filters, value: string) => {
      setFilters((f) => ({ ...f, [name]: value }));
      setPage(1);
    };
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

    /** Creates a task from the issue; rejects so the dialog can show the error. */
    const createTask = async (item: IssueItem, force: boolean) => {
      const key = item.issueKey;
      if (busy.current.has(key)) return;
      const ctx = host.context.getTaskCreationContext?.(workspaceId);
      if (!ctx?.workflowId) {
        setNotice({ key: "errorWorkflow" });
        return;
      }
      busy.current.add(key);
      setCreating((c) => [...c, key]);
      setNotice(undefined);
      try {
        const body: Record<string, unknown> = {
          issueKey: key,
          workflowId: ctx.workflowId,
          workflowStepId: ctx.defaultStepId,
        };
        if (force) body.force = true;
        const r = await host.api.invokeAction<{ taskId: string; taskKey: string }>("issues.create_task", {
          workspaceId,
          body,
        });
        addTask(key, { taskId: r.taskId, taskKey: r.taskKey });
        setCreated((c) => ({ ...c, [key]: r.taskKey }));
      } finally {
        busy.current.delete(key);
        setCreating((c) => c.filter((k) => k !== key));
      }
    };

    const onCreate = (item: IssueItem) => {
      if (item.linkedTasks.length > 0) {
        setLinkedDialog(item);
        return;
      }
      createTask(item, false).catch((e) => setNotice(issueNotice(e)));
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

    const filterCount = [filters.projectKey, filters.statusId, filters.assigneeId].filter(Boolean).length;
    const select = (name: keyof Filters, label: string, choices: Option[] | undefined, testId: string) => (
      <div className={FIELD}>
        <Label htmlFor={testId}>{label}</Label>
        <Select
          value={filters[name] || ALL}
          onValueChange={(v: string) => setFilter(name, v === ALL ? "" : v)}
        >
          <SelectTrigger id={testId} data-testid={testId} className="min-w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL} data-testid={`${testId}-${ALL}`}>
              {messages.filterAll}
            </SelectItem>
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
        {select("projectKey", messages.filterProject, options.projects, "backlog-issues-project")}
        {select("statusId", messages.filterStatus, options.statuses, "backlog-issues-status")}
        {select("assigneeId", messages.filterAssignee, options.assignees, "backlog-issues-assignee")}
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
        {created[item.issueKey] ? (
          <span role="status" data-testid={`backlog-issue-created-${item.issueKey}`}>
            {format(messages.createdTask, { key: created[item.issueKey]! })}
          </span>
        ) : null}
      </span>
    );

    const actionsOf = (item: IssueItem) => {
      const key = item.issueKey;
      const isCreating = creating.includes(key);
      return (
        <RowMenu
          testId={`backlog-issue-menu-${key}`}
          label={format(messages.moreActions, { key })}
          items={[
            {
              testId: `backlog-issue-create-${key}`,
              label: isCreating ? messages.creatingTask : messages.createTask,
              disabled: isCreating,
              onSelect: () => onCreate(item),
            },
            {
              testId: `backlog-issue-link-${key}`,
              label: messages.linkToTask,
              onSelect: () => setLinkDialog(item),
            },
          ]}
        />
      );
    };

    const title = (item: IssueItem) => (
      <span
        data-testid={`backlog-issue-title-${item.issueKey}`}
        className="line-clamp-2"
        title={item.summary}
        tabIndex={0}
      >
        {item.summary}
      </span>
    );

    const keyLink = (item: IssueItem) => (
      <a href={item.url} target="_blank" rel="noreferrer" data-testid={`backlog-issue-key-${item.issueKey}`}>
        {item.issueKey}
      </a>
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
          {isMobile ? (
            <ul data-testid="backlog-issues-cards" className={STACK}>
              {items.map((item) => (
                <li
                  key={item.issueKey}
                  data-testid={`backlog-issue-row-${item.issueKey}`}
                  aria-label={rowLabel(item, messages)}
                  className={STACK}
                >
                  <span className={ROW}>
                    {keyLink(item)}
                    <span>{item.status}</span>
                  </span>
                  {title(item)}
                  {item.assignee ? <span>{item.assignee}</span> : null}
                  {tasksOf(item)}
                  {actionsOf(item)}
                </li>
              ))}
            </ul>
          ) : (
            <Table data-testid="backlog-issues-table">
              <TableHeader>
                <TableRow>
                  <TableHead scope="col">{messages.colKey}</TableHead>
                  <TableHead scope="col">{messages.colTitle}</TableHead>
                  <TableHead scope="col">{messages.filterStatus}</TableHead>
                  <TableHead scope="col">{messages.filterAssignee}</TableHead>
                  <TableHead scope="col">{messages.colUpdated}</TableHead>
                  <TableHead scope="col">{messages.colTasks}</TableHead>
                  <TableHead scope="col">{messages.colActions}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow
                    key={item.issueKey}
                    data-testid={`backlog-issue-row-${item.issueKey}`}
                    aria-label={rowLabel(item, messages)}
                  >
                    <TableCell>{keyLink(item)}</TableCell>
                    <TableCell>{title(item)}</TableCell>
                    <TableCell>{item.status}</TableCell>
                    <TableCell>{item.assignee ?? ""}</TableCell>
                    <TableCell>{item.updatedAt ? relative(item.updatedAt) : ""}</TableCell>
                    <TableCell>{tasksOf(item)}</TableCell>
                    <TableCell>{actionsOf(item)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
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
      <div data-testid="backlog-issues" className={STACK}>
        <div className={ROW}>
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
              className={BUTTON}
              data-testid="backlog-issues-filters-toggle"
              aria-expanded={showFilters}
              aria-controls="backlog-issues-filters"
              onClick={() => setShowFilters((s) => !s)}
            >
              {format(messages.filtersButton, { count: filterCount })}
            </Button>
          ) : null}
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className={BUTTON}
            data-testid="backlog-issues-refresh"
            aria-label={refreshing ? messages.refreshing : messages.refresh}
            disabled={refreshing}
            onClick={() => void refresh()}
          >
            {icon(host, "refresh", refreshing ? "h-4 w-4 animate-spin" : "h-4 w-4")}
          </Button>
        </div>
        {!isMobile || showFilters ? <div id="backlog-issues-filters">{filterFields}</div> : null}
        {refreshedAt ? (
          <p data-testid="backlog-issues-updated">
            {format(messages.updatedAt, { time: relative(refreshedAt) })}
          </p>
        ) : null}
        <h2 ref={list} tabIndex={-1} data-testid="backlog-issues-list">
          {messages.issuesListLabel}
        </h2>
        {body}
        {notice ? (
          <p role="alert" data-testid="backlog-issues-notice">
            {t(notice)}
          </p>
        ) : null}
        <div role="status" aria-live="polite" className="sr-only" data-testid="backlog-issues-announcer">
          {announcement}
        </div>
        {linkedDialog ? (
          <ConfirmDialog
            testId="backlog-issue-linked-dialog"
            title={messages.alreadyLinkedTitle}
            body={format(messages.alreadyLinkedBody, { key: linkedDialog.issueKey })}
            confirmLabel={messages.createAnother}
            onConfirm={() => createTask(linkedDialog, true)}
            onClose={() => setLinkedDialog(undefined)}
          >
            {linkedDialog.linkedTasks.slice(0, 1).map((task) => (
              <a
                key={task.taskId}
                href={taskHref(task.taskId)}
                data-testid="backlog-issue-linked-dialog-open"
                onClick={(e: { preventDefault(): void }) => {
                  e.preventDefault();
                  host.navigate(taskHref(task.taskId));
                }}
              >
                {format(messages.openTask, { key: task.taskKey ?? task.taskId })}
              </a>
            ))}
          </ConfirmDialog>
        ) : null}
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
