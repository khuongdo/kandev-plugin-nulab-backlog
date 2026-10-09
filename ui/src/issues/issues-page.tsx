import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { createSaveQueryDialog } from "../git/save-query-dialog";
import { createErrorAlert } from "../error-alert";
import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON, FILTER_POPOVER, FILTER_TRIGGER, FILTERS, RESULTS, ROW, STACK } from "../layout";
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
  taskRowLinks,
  type IssueItem,
  type IssueQuery,
  type IssuePage,
  type IssuesFailure,
  type TaskLink,
} from "./issues-state";
import { createLinkTaskDialog } from "./link-task-dialog";
import type { LinksStore } from "./links-store";

const PAGE_SIZE = 20;
/** The "Not closed" status choice and the "Me" assignee choice (FR4.1, FR4.2). */
const OPEN = "open";
const ME = "me";
/** The status choice of a saved query that picks several statuses. */
const CUSTOM = "custom";

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

/** The toolbar's last-fetched time: a Date, or null when unknown or not a valid time. */
function fetchedAt(at: string | undefined): Date | null {
  const d = at ? new Date(at) : null;
  return d && !Number.isNaN(d.getTime()) ? d : null;
}

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
 * US4.2; FR1, FR4, FR5.3): the host's list toolbar with the query box
 * (committed on Enter or blur) and GitHub-style dropdown filters, 20 rows per
 * page in the host's change request rows, linked tasks through the host's
 * task indicator, the "+ Task" quick action menu and Link to task per row,
 * and Refresh. Phones get the same toolbar stacked, all filters visible.
 */
export function createIssuesPage(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): Component<IssuesPageProps> {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const ui = hostUi(host);
  const { Button, Skeleton, Pagination, PaginationContent, PaginationItem } = ui;
  const { ChangeRequestList, ChangeRequestRow, IntegrationListToolbar, IntegrationRepositoryFilter } = ui;
  const { TaskRowIndicator } = ui;
  const { RowMenu } = createSectionParts(host, messages);
  const LinkTaskDialog = createLinkTaskDialog(host, messages, store);
  const StartTask = createStartTask(host, messages);
  const ErrorAlert = createErrorAlert(host);
  const SaveQueryDialog = createSaveQueryDialog(host, messages);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;
  const t = (n: Notice) => noticeText(n, messages);

  return function IssuesPage({
    workspaceId,
    quickActions = [],
    selection,
    onSavedQuery,
    saveRequest = 0,
  }: IssuesPageProps) {
    // Undefined until the selection is applied: the preset needs the statuses first.
    const [filters, setFilters] = useState<Filters | undefined>(selection ? undefined : NO_FILTERS);
    // FR4.2: the query box holds a draft; Enter or leaving the box commits it.
    const [draftQuery, setDraftQuery] = useState("");
    const [committedQuery, setCommittedQuery] = useState("");
    const [page, setPage] = useState(1);
    const [nonce, setNonce] = useState(0);
    const [load, setLoad] = useState<Load>({ kind: "loading" });
    // R-04: the toolbar keeps the last count and time while another page or filter loads.
    const [shown, setShown] = useState<IssuePage | undefined>(undefined);
    const [options, setOptions] = useState<FilterOptions | undefined>(undefined);
    const [saving, setSaving] = useState(false);
    const [announcement, setAnnouncement] = useState("");
    const [countdown, setCountdown] = useState<number | undefined>(undefined);
    const [linkDialog, setLinkDialog] = useState<IssueItem | undefined>(undefined);
    const [refreshing, setRefreshing] = useState(false);
    const latest = useRef(0);
    const focusList = useRef(false);
    const results = useRef<HTMLDivElement | null>(null);
    const applied = useRef("");
    // R-01: true from a pointer press on a filter dropdown until the pointer is released.
    const pickingFilter = useRef(false);

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
      setDraftQuery(q?.keyword ?? "");
      setCommittedQuery(q?.keyword ?? "");
      setPage(1);
    }, [selection?.key, options]);

    useEffect(() => {
      if (saveRequest > 0) setSaving(true);
    }, [saveRequest]);

    useEffect(() => {
      if (!filters) return;
      const seq = ++latest.current;
      setLoad({ kind: "loading" });
      setAnnouncement(messages.issuesLoading);
      host.api
        .invokeAction<IssuePage>("issues.list", {
          workspaceId,
          body: listBody(page, committedQuery, filters),
        })
        .then((r) => {
          if (seq !== latest.current) return;
          const ready = { ...r, items: r?.items ?? [] };
          setLoad({ kind: "ready", page: ready });
          setShown(ready);
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
    }, [workspaceId, page, committedQuery, filters, nonce]);

    useEffect(() => {
      if (load.kind === "ready" && focusList.current) {
        focusList.current = false;
        results.current?.focus();
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
    // BR4.5: a filter change reloads from page 1, taking a typed but uncommitted query along (R-04).
    const setFilter = (next: Partial<Filters>) => {
      pickingFilter.current = false;
      setFilters((f) => ({ ...(f ?? NO_FILTERS), ...next }));
      setDraftQuery(draftQuery.trim());
      setCommittedQuery(draftQuery.trim());
      setPage(1);
    };
    // The status choice: All (""), Not closed (every status but Closed), one status, or a saved set.
    const statusValue = (() => {
      const ids = current.statusIds;
      if (ids.length === 0) return "";
      if (ids.length === openIds.length && ids.every((id) => openIds.includes(id))) return OPEN;
      return ids.length === 1 ? String(ids[0]) : CUSTOM;
    })();
    const reset = () => {
      setFilters(NO_FILTERS);
      setDraftQuery("");
      setCommittedQuery("");
      setPage(1);
      reload();
    };
    // BR4.2, BR4.3: commit the trimmed draft; the same value does not reload, blank clears the keyword.
    // R-01: pressing a filter dropdown blurs the box first; that blur leaves the draft for the pick,
    // so a typed query and the following pick make one reload. Enter and other blurs commit at once.
    const commitQuery = () => {
      if (pickingFilter.current) return;
      const next = draftQuery.trim();
      setDraftQuery(next);
      if (next === committedQuery) return;
      setCommittedQuery(next);
      setPage(1);
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
      try {
        await host.api.invokeAction("issues.refresh", { workspaceId });
      } catch (e) {
        host.toast.error(t(issueNotice(e))); // a click error (intent 261009, FR2.1)
      }
      setRefreshing(false);
      reload();
    };

    // BR4.4: GitHub's searchable dropdowns without field labels; "" is the All choice (R-06).
    const filter = (
      id: string,
      text: { label: string; all: string },
      props: { value: string; onChange: (v: string) => void; options: { value: string; label: string }[] },
    ) => (
      <IntegrationRepositoryFilter
        value={props.value}
        onValueChange={props.onChange}
        options={props.options}
        ariaLabel={text.label}
        allLabel={text.all}
        testId={`backlog-issues-${id}`}
        triggerClassName={FILTER_TRIGGER}
        className={FILTER_POPOVER}
      />
    );
    const choices = (list: Option[] | undefined) =>
      (list ?? []).map((o) => ({ value: o.key ?? String(o.id), label: o.name }));
    // A saved query may pick several statuses: shown as "n statuses", not as All.
    const savedSet =
      statusValue === CUSTOM
        ? [{ value: CUSTOM, label: format(messages.statusCount, { count: current.statusIds.length }) }]
        : [];
    const filterFields = (
      <div id="backlog-issues-filters" data-testid="backlog-issues-filters" className={FILTERS}>
        <div
          className="contents"
          onPointerDownCapture={() => {
            pickingFilter.current = true;
            document.addEventListener("pointerup", () => (pickingFilter.current = false), { once: true });
          }}
        >
          {filter(
            "project",
            { label: messages.filterProject, all: messages.filterAllProjects },
            {
              value: current.projectKey,
              onChange: (v) => setFilter({ projectKey: v }),
              options: choices(options?.projects),
            },
          )}
          {filter(
            "status",
            { label: messages.filterStatus, all: messages.filterAllStatuses },
            {
              value: statusValue,
              onChange: (v) => {
                if (v !== CUSTOM) setFilter({ statusIds: !v ? [] : v === OPEN ? openIds : [Number(v)] });
              },
              options: [
                { value: OPEN, label: messages.statusNotClosed },
                ...savedSet,
                ...choices(options?.statuses),
              ],
            },
          )}
          {filter(
            "assignee",
            { label: messages.filterAssignee, all: messages.filterAllAssignees },
            {
              value: current.assignee,
              onChange: (v) => setFilter({ assignee: v }),
              options: [{ value: ME, label: messages.whoMe }, ...choices(options?.assignees)],
            },
          )}
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          className={BUTTON}
          data-testid="backlog-issues-save-query"
          disabled={!filters}
          onClick={() => setSaving(true)}
        >
          {messages.saveQuery}
        </Button>
      </div>
    );

    // FR1, BR1.1-BR1.5: the host shows the task (title, or key/id) and opens it; nothing when empty.
    const tasksOf = (item: IssueItem) => (
      <TaskRowIndicator
        tasks={taskRowLinks(item.linkedTasks)}
        testIdPrefix={`backlog-issue-task-${item.issueKey}`}
      />
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
            <ErrorAlert
              message={text}
              testId="backlog-issues-state"
              action={
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
              }
            />
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
            <ErrorAlert
              message={`${messages.issuesLoadFailed} ${t(load.notice)}`}
              testId="backlog-issues-error"
              action={
                <Button
                  type="button"
                  variant="outline"
                  className={BUTTON}
                  data-testid="backlog-issues-retry"
                  onClick={reload}
                >
                  {messages.retry}
                </Button>
              }
            />
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

    return (
      <div data-testid="backlog-issues" className="flex min-w-0 flex-col">
        <IntegrationListToolbar
          title={messages.issuesListLabel}
          count={shown?.total ?? 0}
          loading={refreshing || (load.kind === "loading" && !shown)}
          lastFetchedAt={fetchedAt(shown?.refreshedAt)}
          customQuery={draftQuery}
          committedQuery={committedQuery}
          onCustomQueryChange={setDraftQuery}
          onCommitCustomQuery={commitQuery}
          onRefresh={() => void refresh()}
          queryPlaceholder={messages.searchIssues}
          titleTestId="backlog-issues-list"
          queryTestId="backlog-issues-search"
          refreshTestId="backlog-issues-refresh"
          filter={filterFields}
        />
        <div className={RESULTS} data-testid="backlog-issues-results" ref={results} tabIndex={-1}>
          {body}
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
                keyword: committedQuery, // BR4.7: never an uncommitted draft
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
