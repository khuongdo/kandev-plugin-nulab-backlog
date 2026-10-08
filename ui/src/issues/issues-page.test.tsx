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
  openFilter,
  pseudoCatalogue,
  press,
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
    expect(c.querySelector('[data-host="IntegrationListToolbar"]')!.getAttribute("data-loading")).toBe(
      "true",
    );
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
    expect(document.activeElement).toBe(byTestId(c, "backlog-issues-results"));
    expect(byTestId(c, "backlog-issues-announcer")!.textContent).toBe("Showing 41-57 of 57");
  });

  it("keeps the toolbar count and Refresh while another page loads; only the first load and Refresh show loading (R-04, FR4.1)", async () => {
    const pending = deferred<unknown>(); // page 2 never answers during the test
    const host = setup({
      "issues.list": async (body) =>
        body.page === 1 ? { items: ALL.slice(0, 20), total: 57, page: 1, pageSize: 20 } : pending.promise,
    });
    const c = await render(host);
    const toolbar = () => c.querySelector('[data-host="IntegrationListToolbar"]')!;
    expect(toolbar().getAttribute("data-loading")).toBe("false");
    await act(async () => byTestId(c, "backlog-issues-next")!.click());
    expect(byTestId(c, "backlog-issues-loading")).not.toBeNull();
    expect(toolbar().getAttribute("data-count")).toBe("57");
    expect(toolbar().getAttribute("data-loading")).toBe("false");
    expect((byTestId(c, "backlog-issues-refresh") as HTMLButtonElement).disabled).toBe(false);
  });

  it("uses the host list toolbar and commits the query on Enter or blur, never while typing (FR4.1, FR4.2, BR4.1-BR4.3)", async () => {
    const host = setup();
    const c = await render(host);
    const toolbar = c.querySelector('[data-host="IntegrationListToolbar"]')!;
    expect(byTestId(c, "backlog-issues-list")!.textContent).toBe(en.issuesListLabel);
    expect(toolbar.getAttribute("data-count")).toBe("57");
    expect(toolbar.getAttribute("data-last-fetched")).toBe("2026-10-06T00:00:00.000Z");
    const input = byTestId(c, "backlog-issues-search") as HTMLInputElement;
    // R-03: the host input has no aria-label prop; the placeholder is its accessible name.
    expect(input.getAttribute("placeholder")).toBe(en.searchIssues);
    await act(async () => byTestId(c, "backlog-issues-next")!.click());
    const before = calls(host, "issues.list").length;
    await act(async () => setValue(input, "log"));
    await act(async () => setValue(input, " login "));
    expect(calls(host, "issues.list")).toHaveLength(before);
    await act(async () => press(input, "Enter"));
    expect(calls(host, "issues.list")).toHaveLength(before + 1);
    expect(calls(host, "issues.list").at(-1)![1]).toEqual({
      workspaceId: "ws-1",
      body: { page: 1, pageSize: 20, keyword: "login" },
    });
    expect(input.value).toBe("login");
    await act(async () => press(input, "Enter")); // the same value: no reload
    expect(calls(host, "issues.list")).toHaveLength(before + 1);
    await act(async () => input.focus());
    await act(async () => setValue(input, "  "));
    await act(async () => input.blur()); // leaving a changed box commits it; blank clears the keyword
    expect(calls(host, "issues.list")).toHaveLength(before + 2);
    expect(calls(host, "issues.list").at(-1)![1]).toMatchObject({ body: { page: 1, keyword: "" } });
  });

  it("filters with label-less searchable host dropdowns that reload from page 1 (FR4.3, FR4.4, BR4.4, BR4.5, R-02, R-06)", async () => {
    const host = setup();
    const c = await render(host);
    for (const [id, label, all] of [
      ["project", en.filterProject, en.filterAllProjects],
      ["status", en.filterStatus, en.filterAllStatuses],
      ["assignee", en.filterAssignee, en.filterAllAssignees],
    ] as const) {
      const filter = byTestId(c, `backlog-issues-${id}`)!;
      expect(filter.getAttribute("data-host")).toBe("IntegrationRepositoryFilter");
      expect(filter.getAttribute("aria-label")).toBe(label);
      expect(byTestId(c, `backlog-issues-${id}-option-all`)!.textContent).toBe(all);
      // R-02: GitHub's widths: full width on phones, 220 px trigger and 360 px list from md.
      for (const cls of ["w-full", "md:w-[220px]"])
        expect(filter.getAttribute("data-trigger-class")).toContain(cls);
      expect(filter.getAttribute("data-popover-class")).toBe("md:min-w-[360px]");
      expect(c.querySelector(`label[for="backlog-issues-${id}"]`)).toBeNull();
    }
    expect(byTestId(c, "backlog-issues-status-option-open")!.textContent).toBe(en.statusNotClosed);
    expect(byTestId(c, "backlog-issues-assignee-option-me")!.textContent).toBe(en.whoMe);
    await act(async () => byTestId(c, "backlog-issues-next")!.click());
    await act(async () => byTestId(c, "backlog-issues-status-option-2")!.click());
    expect(calls(host, "issues.list").at(-1)![1]).toEqual({
      workspaceId: "ws-1",
      body: { page: 1, pageSize: 20, keyword: "", statusIds: [2] },
    });
    expect(byTestId(c, "backlog-issues-status")!.getAttribute("data-value")).toBe("2");
    await act(async () => byTestId(c, "backlog-issues-assignee-option-2")!.click());
    await act(async () => byTestId(c, "backlog-issues-project-option-PROJ")!.click());
    expect(calls(host, "issues.list").at(-1)![1]).toEqual({
      workspaceId: "ws-1",
      body: { page: 1, pageSize: 20, keyword: "", statusIds: [2], assigneeIds: [2], projectKeys: ["PROJ"] },
    });
    await act(async () => byTestId(c, "backlog-issues-status-option-open")!.click());
    expect(calls(host, "issues.list").at(-1)![1]).toMatchObject({ body: { statusIds: [1, 2] } });
    await act(async () => byTestId(c, "backlog-issues-status-option-all")!.click()); // "" is All (R-06)
    expect(calls(host, "issues.list").at(-1)![1]).not.toHaveProperty("body.statusIds");
    expect(byTestId(c, "backlog-issues-status")!.getAttribute("data-value")).toBe("");
    // The filter slot stacks on phones and is one row from md; Save query is last (BR4.8, BR5.4).
    const slot = byTestId(c, "backlog-issues-filters")!;
    for (const cls of ["w-full", "flex-col", "md:w-auto", "md:flex-row"])
      expect(slot.classList).toContain(cls);
    expect(slot.lastElementChild).toBe(byTestId(c, "backlog-issues-save-query"));
  });

  it("commits a typed query with the dropdown pick that follows it in one reload, although opening the dropdown blurs the box first (R-01, BR4.2, BR4.5)", async () => {
    const host = setup();
    const c = await render(host);
    const before = calls(host, "issues.list").length;
    const input = byTestId(c, "backlog-issues-search") as HTMLInputElement;
    await act(async () => input.focus());
    await act(async () => setValue(input, " login"));
    await act(async () => openFilter(c, "backlog-issues-status")); // the host blurs the box here
    expect(calls(host, "issues.list")).toHaveLength(before);
    await act(async () => byTestId(c, "backlog-issues-status-option-2")!.click());
    expect(calls(host, "issues.list")).toHaveLength(before + 1);
    expect(calls(host, "issues.list").at(-1)![1]).toMatchObject({
      body: { keyword: "login", statusIds: [2] },
    });
    expect((byTestId(c, "backlog-issues-search") as HTMLInputElement).value).toBe("login");
    // Once the press is released, leaving the box commits again (the dropdown no longer holds the draft).
    await act(async () => openFilter(c, "backlog-issues-project"));
    await act(async () => document.dispatchEvent(new MouseEvent("pointerup", { bubbles: true })));
    await act(async () => input.focus());
    await act(async () => setValue(input, "crash"));
    await act(async () => input.blur());
    expect(calls(host, "issues.list")).toHaveLength(before + 2);
    expect(calls(host, "issues.list").at(-1)![1]).toMatchObject({
      body: { keyword: "crash", statusIds: [2] },
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

  it("shows linked tasks with the host task indicator and the row menu (FR1, BR1.1-BR1.5, AC2.3.1)", async () => {
    const tasks = (n: number) =>
      Array.from({ length: n }, (_, i) => ({ taskId: `task-${i + 1}`, taskKey: i ? undefined : "T-1" }));
    const host = setup({
      "issues.list": async () => ({
        items: [issue(17, { linkedTasks: tasks(1) }), issue(18, { linkedTasks: tasks(3) }), issue(19)],
        total: 3,
        page: 1,
        pageSize: 20,
      }),
    });
    const c = await render(host);
    const single = byTestId(c, "backlog-issue-task-PROJ-17-single")!;
    expect(single.getAttribute("data-host")).toBe("TaskRowIndicator");
    expect(single.textContent).toBe("T-1");
    const multi = byTestId(c, "backlog-issue-task-PROJ-18-multi")!;
    expect(multi.textContent).toBe("Tasks 3");
    expect(byTestId(c, "backlog-issue-task-PROJ-18-item-task-2")!.textContent).toBe("task-2");
    expect(
      byTestId(c, "backlog-issue-row-PROJ-19")!.querySelector('[data-host="TaskRowIndicator"]'),
    ).toBeNull();
    expect(byTestId(c, "backlog-issue-task-PROJ-19-empty")).toBeNull();
    await act(async () => single.click());
    expect(window.location.pathname).toBe("/t/task-1");
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
    expect(byTestId(c, "backlog-issue-task-PROJ-56-single")!.textContent).toBe("T-1");
  });

  it("refreshes, then reloads, disabled while running (AC4.2.3)", async () => {
    const pending = deferred<unknown>();
    const host = setup({ "issues.refresh": () => pending.promise });
    const c = await render(host);
    const button = byTestId(c, "backlog-issues-refresh") as HTMLButtonElement;
    expect(button.getAttribute("data-host")).toBe("Button");
    await act(async () => button.click());
    expect(button.disabled).toBe(true);
    const before = calls(host, "issues.list").length;
    await act(async () => pending.resolve({ updatedCount: 2, refreshedAt: "2026-10-06T00:05:00Z" }));
    expect(calls(host, "issues.list")).toHaveLength(before + 1);
    expect(button.disabled).toBe(false);
  });

  it("shows every filter at full width on phones, with no Filters (n) toggle (M2m, FR4.6, BR4.8)", async () => {
    const c = await render(setup({}, { useResponsiveBreakpoint: () => ({ isMobile: true }) }));
    expect(c.querySelectorAll('[data-host="ChangeRequestRow"]')).toHaveLength(20);
    expect(byTestId(c, "backlog-issues-filters-toggle")).toBeNull();
    for (const id of ["project", "status", "assignee"])
      expect(byTestId(c, `backlog-issues-${id}`)!.getAttribute("data-trigger-class")).toMatch(
        /(^| )w-full( |$)/,
      );
    expect(byTestId(c, "backlog-issues-filters")!.classList).toContain("flex-col");
    expect(byTestId(c, "backlog-issues-save-query")).not.toBeNull();
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
