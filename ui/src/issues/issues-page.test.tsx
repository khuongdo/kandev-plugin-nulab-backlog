import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  deferred,
  expectOnlyCatalogueText,
  expectTestIds,
  fakeHost,
  mount,
  pseudoCatalogue,
  choose,
  rawControls,
  setValue,
  text,
  unmount,
} from "../testing/harness";
import { createIssuesPage } from "./issues-page";
import type { IssueItem } from "./issues-state";

afterEach(() => {
  unmount();
  vi.useRealTimers();
});

const JP = "ログイン画面でセッションがタイムアウトした後に再ログインすると入力内容が失われる".repeat(4);

function issue(n: number, extra: Partial<IssueItem> = {}): IssueItem {
  return {
    issueKey: `PROJ-${n}`,
    summary: n === 57 ? JP : `Issue ${n}`,
    status: "Open",
    statusId: 1,
    assignee: "Lan",
    updatedAt: "2026-10-01T09:00:00Z",
    url: `https://example-space.backlog.com/view/PROJ-${n}`,
    linkedTasks: [],
    ...extra,
  };
}

const ALL = Array.from({ length: 57 }, (_, i) => issue(57 - i));

type Handlers = Record<string, (body: Record<string, unknown>) => Promise<unknown>>;

function setup(handlers: Handlers = {}, overrides: Record<string, unknown> = {}) {
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    if (key === "issues.filters")
      return {
        projects: [{ key: "PROJ", name: "Test Project" }],
        statuses: [
          { id: 1, name: "Open" },
          { id: 2, name: "In Progress" },
        ],
        assignees: [{ id: 2, name: "Lan" }],
      };
    if (key === "issues.list") {
      const page = (body.page as number) ?? 1;
      return {
        items: ALL.slice((page - 1) * 20, page * 20),
        total: 57,
        page,
        pageSize: 20,
        refreshedAt: "2026-10-06T00:00:00Z",
        connectionEpoch: 1,
      };
    }
    throw new Error(`unexpected ${key}`);
  }, overrides);
}

const calls = (host: ReturnType<typeof setup>, key: string) =>
  vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);

const render = (host: ReturnType<typeof setup>, messages = en) =>
  mount(createIssuesPage(host, messages), { workspaceId: "ws-1" });

describe("Issues page (M2, US2.1, US2.2)", () => {
  it("shows skeleton rows with the filters while loading, announced once (AC2.1.5)", async () => {
    const pending = deferred<unknown>();
    const c = await render(setup({ "issues.list": () => pending.promise }));
    expect(
      byTestId(c, "backlog-issues-loading")!.querySelectorAll('[data-testid="fake-skeleton"]').length,
    ).toBe(5);
    expect(byTestId(c, "backlog-issues-search")).not.toBeNull();
    expect(byTestId(c, "backlog-issues-announcer")!.textContent).toBe(en.issuesLoading);
    await act(async () => pending.resolve({ items: [issue(1)], total: 1, page: 1, pageSize: 20 }));
    expect(byTestId(c, "backlog-issues-loading")).toBeNull();
  });

  it("lists 20 rows and pages to 41-57 of 57 (AC2.1.1, AC2.1.2)", async () => {
    const c = await render(setup());
    expect(c.querySelectorAll('[data-host="ChangeRequestRow"]')).toHaveLength(20);
    const row = byTestId(c, "backlog-issue-row-PROJ-56")!;
    expect(row.textContent).toContain("PROJ-56");
    expect(row.textContent).toContain("Issue 56");
    expect(row.textContent).toContain("Open");
    expect(row.textContent).toContain("Lan");
    expect(row.textContent).toContain("rel:2026-10-01T09:00:00Z");
    expect(byTestId(c, "backlog-issues-showing")!.textContent).toBe("Showing 1-20 of 57");
    await act(async () => byTestId(c, "backlog-issues-next")!.click());
    await act(async () => byTestId(c, "backlog-issues-next")!.click());
    expect(byTestId(c, "backlog-issues-showing")!.textContent).toBe("Showing 41-57 of 57");
    expect((byTestId(c, "backlog-issues-next") as HTMLButtonElement).disabled).toBe(true);
    expect(document.activeElement).toBe(byTestId(c, "backlog-issues-list"));
    expect(byTestId(c, "backlog-issues-announcer")!.textContent).toBe("Showing 41-57 of 57");
  });

  it("sends the filters and waits 400 ms before a search (AC2.2.1)", async () => {
    vi.useFakeTimers();
    const host = setup();
    const c = await render(host);
    await act(async () => choose(c, "backlog-issues-status", "2"));
    expect(calls(host, "issues.list").at(-1)![1]).toEqual({
      workspaceId: "ws-1",
      body: { page: 1, pageSize: 20, keyword: "", statusIds: [2] },
    });
    await act(async () => choose(c, "backlog-issues-assignee", "2"));
    await act(async () => choose(c, "backlog-issues-project", "PROJ"));
    const before = calls(host, "issues.list").length;
    await act(async () => setValue(byTestId(c, "backlog-issues-search") as HTMLInputElement, "log"));
    await act(async () => setValue(byTestId(c, "backlog-issues-search") as HTMLInputElement, "login"));
    expect(calls(host, "issues.list")).toHaveLength(before);
    await act(async () => vi.advanceTimersByTime(400));
    expect(calls(host, "issues.list")).toHaveLength(before + 1);
    expect(calls(host, "issues.list").at(-1)![1]).toEqual({
      workspaceId: "ws-1",
      body: {
        page: 1,
        pageSize: 20,
        keyword: "login",
        statusIds: [2],
        assigneeIds: [2],
        projectKeys: ["PROJ"],
      },
    });
  });

  it("shows the empty state with Reset filters, and errors with Retry (AC2.1.3, AC2.1.4)", async () => {
    let mode: "empty" | "error" | "ok" = "empty";
    const host = setup({
      "issues.list": async () => {
        if (mode === "error") throw actionError(503, { code: "unreachable" });
        return {
          items: mode === "empty" ? [] : [issue(1)],
          total: mode === "empty" ? 0 : 1,
          page: 1,
          pageSize: 20,
        };
      },
    });
    const c = await render(host);
    expect(byTestId(c, "backlog-issues-empty")!.textContent).toContain(en.issuesEmpty);
    mode = "error";
    await act(async () => byTestId(c, "backlog-issues-reset")!.click());
    expect(byTestId(c, "backlog-issues-error")!.textContent).toContain(en.unreachable);
    mode = "ok";
    await act(async () => byTestId(c, "backlog-issues-retry")!.click());
    expect(byTestId(c, "backlog-issue-row-PROJ-1")).not.toBeNull();
  });

  it("links to the settings when no project, not connected or sign-in needed (AC1.4.3, AC1.7.2)", async () => {
    for (const [error, key] of [
      [actionError(400, { code: "validation", field: "projectKeys" }), "issuesNoProject"],
      [actionError(401, { code: "reconnect_required" }), "issuesSignInAgain"],
    ] as const) {
      const host = setup({
        "issues.list": async () => {
          throw error;
        },
      });
      const c = await render(host);
      expect(text(c)).toContain(en[key]);
      const link = byTestId(c, "backlog-issues-settings-link")!;
      expect(link.getAttribute("href")).toBe("/settings/workspaces/ws-1/integrations/nulab-backlog");
      await act(async () => link.click());
      expect(host.navigate).toHaveBeenCalled();
      unmount();
    }
  });

  it("shows each issue as a host change request row linking to Backlog (AC2.1.6, FR5.3)", async () => {
    const c = await render(setup());
    const row = byTestId(c, "backlog-issue-row-PROJ-57")!;
    expect(row.getAttribute("data-host")).toBe("ChangeRequestRow");
    const link = byTestId(c, "backlog-issue-row-PROJ-57-link")!;
    expect(link.textContent).toBe(JP);
    expect(link.getAttribute("href")).toBe("https://example-space.backlog.com/view/PROJ-57");
  });

  it("shows linked task keys and the row menu (AC2.3.1)", async () => {
    const host = setup({
      "issues.list": async () => ({
        items: [issue(17, { linkedTasks: [{ taskId: "task-17", taskKey: "T-17" }] })],
        total: 1,
        page: 1,
        pageSize: 20,
      }),
    });
    const c = await render(host);
    const link = byTestId(c, "backlog-issue-task-PROJ-17-task-17")!;
    expect(link.textContent).toBe("T-17");
    expect(link.getAttribute("href")).toBe("/t/task-17");
    expect(byTestId(c, "backlog-issue-menu-PROJ-17")!.getAttribute("aria-label")).toBe(
      "More actions for PROJ-17",
    );
    expect(byTestId(c, "backlog-issue-create-PROJ-17")).toBeNull();
    expect(byTestId(c, "backlog-issue-link-PROJ-17")!.textContent).toBe(en.linkToTask);
  });

  it("puts the + Task menu in the row and adds the linked task after the dialog (FR1.1, FR1.3, FR1.4)", async () => {
    const host = setup({ "issues.link": async () => ({ taskId: "t-1", taskKey: "T-1" }) });
    const c = await mount(createIssuesPage(host, en), {
      workspaceId: "ws-1",
      quickActions: [
        { id: "implement", label: "Implement", hint: "", icon: "code", promptTemplate: "Do {{title}}" },
      ],
    });
    const action = byTestId(c, "backlog-issue-row-PROJ-56-action")!;
    expect(action.querySelector('[data-testid="backlog-issue-PROJ-56-start"]')).not.toBeNull();
    await act(async () => action.querySelector<HTMLElement>('[data-preset-id="implement"]')!.click());
    expect(byTestId(c, "fake-task-create-description")!.textContent).toBe("Do Issue 56");
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(calls(host, "issues.link")[0]![1]).toEqual({
      workspaceId: "ws-1",
      taskId: "t-1",
      body: { issueKey: "PROJ-56" },
    });
    expect(byTestId(c, "backlog-issue-task-PROJ-56-t-1")!.textContent).toBe("T-1");
  });

  it("refreshes, then reloads, disabled while running (AC4.2.3)", async () => {
    const pending = deferred<unknown>();
    const host = setup({ "issues.refresh": () => pending.promise });
    const c = await render(host);
    expect(byTestId(c, "backlog-issues-updated")!.textContent).toBe("Updated at rel:2026-10-06T00:00:00Z");
    const button = byTestId(c, "backlog-issues-refresh") as HTMLButtonElement;
    expect(button.getAttribute("data-variant")).toBe("ghost");
    expect(button.getAttribute("aria-label")).toBe(en.refresh);
    await act(async () => button.click());
    expect(button.disabled).toBe(true);
    expect(button.getAttribute("aria-label")).toBe(en.refreshing);
    const before = calls(host, "issues.list").length;
    await act(async () => pending.resolve({ updatedCount: 2, refreshedAt: "2026-10-06T00:05:00Z" }));
    expect(calls(host, "issues.list")).toHaveLength(before + 1);
    expect(button.disabled).toBe(false);
  });

  it("keeps the rows and adds a Filters (n) drawer on mobile (M2m)", async () => {
    const c = await render(setup({}, { useResponsiveBreakpoint: () => ({ isMobile: true }) }));
    expect(c.querySelectorAll('[data-host="ChangeRequestRow"]')).toHaveLength(20);
    const toggle = byTestId(c, "backlog-issues-filters-toggle")!;
    expect(toggle.textContent).toBe("Filters (0)");
    expect(toggle.getAttribute("aria-expanded")).toBe("false");
    expect(byTestId(c, "backlog-issues-status")).toBeNull();
    await act(async () => toggle.click());
    expect(byTestId(c, "backlog-issues-status")).not.toBeNull();
    await act(async () => choose(c, "backlog-issues-status", "1"));
    expect(byTestId(c, "backlog-issues-filters-toggle")!.textContent).toBe("Filters (1)");
  });

  it("is accessible, has test ids and only catalogue text (AC8.2.2)", async () => {
    const c = await render(setup());
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
    expect(rawControls(c)).toEqual([]); // BR5.1
    expect(byTestId(c, "backlog-issue-row-PROJ-57")!.getAttribute("data-host")).toBe("ChangeRequestRow");
    unmount();
    const pseudo = setup({
      "issues.filters": async () => ({
        projects: [{ key: "PROJ", name: "⟦Test Project⟧" }],
        statuses: [],
        assignees: [],
      }),
      "issues.list": async () => ({
        items: [
          {
            issueKey: "⟦PROJ-1⟧",
            summary: "⟦S⟧",
            status: "⟦Open⟧",
            statusId: 1,
            updatedAt: "",
            url: "",
            linkedTasks: [],
          },
        ],
        total: 1,
        page: 1,
        pageSize: 20,
      }),
    });
    const p = await render(pseudo, pseudoCatalogue(en));
    expectOnlyCatalogueText(p, (ok, msg) => expect(ok, msg).toBe(true));
  });
});
