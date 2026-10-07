import { act } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { createBacklogPage } from "./BacklogPage";
import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  choose,
  connected,
  expectTestIds,
  fakeHost,
  mount,
  notConnected,
  rawControls,
  setValue,
  text,
  unmount,
  type View,
} from "../testing/harness";

const SETTINGS_HREF = "/settings/workspaces/ws-1/integrations/nulab-backlog";
const VIEW: View = { ...connected, selectedProjects: ["PROJ"] };

function pr(n: number) {
  return {
    number: n,
    title: `Change ${n}`,
    status: n % 2 ? "open" : "merged",
    author: "Test User",
    assignee: "Lan",
    updated: "2026-10-02T09:00:00Z",
    url: `https://example-space.backlog.com/git/PROJ/web-app/pullRequests/${n}`,
    linkedTaskIds: n === 45 ? ["task-9"] : [],
  };
}

const QUERIES = [
  {
    id: "q1",
    name: "Merged by me",
    projectKey: "PROJ",
    repoName: "web-app",
    statuses: ["merged"],
    assignee: "me",
    creator: "anyone",
  },
  {
    id: "q2",
    name: "Old project",
    projectKey: "DEMO",
    repoName: "demo",
    statuses: ["open"],
    assignee: "anyone",
    creator: "anyone",
  },
];

type Handler = (body: Record<string, unknown>) => Promise<unknown>;

function setup(view: View = VIEW, handlers: Record<string, Handler> = {}) {
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    switch (key) {
      case "connection.get":
        return view;
      case "issues.filters":
        return { projects: [], statuses: [], assignees: [] };
      case "issues.list":
        return { items: [], total: 0, page: 1, pageSize: 20 };
      case "git.queries.list":
        return { queries: QUERIES };
      case "git.repositories.list":
        return { repositories: [{ ownerOrProject: "PROJ", repositoryName: "web-app", repositoryId: "11" }] };
      case "git.prs.list": {
        const page = body.page as number;
        const all = Array.from({ length: 45 }, (_, i) => pr(45 - i));
        return {
          items: all.slice((page - 1) * 20, page * 20),
          page,
          pageSize: 20,
          total: 45,
          hasNext: page * 20 < 45,
        };
      }
    }
    throw new Error(`unexpected ${key}`);
  });
}

const render = (host: ReturnType<typeof setup>) => mount(createBacklogPage(host, en), {});
const calls = (host: ReturnType<typeof setup>, key: string) =>
  vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);

async function openPRs(c: HTMLElement) {
  await act(async () => byTestId(c, "backlog-scope-bar-kind-prs")!.click());
}

async function pickRepo(c: HTMLElement) {
  await act(async () => byTestId(c, "backlog-prs-repo-option-PROJ/web-app")!.click());
}

beforeEach(() => window.history.replaceState(null, "", "/backlog"));
afterEach(unmount);

describe("/backlog alert (BR2.3, FR2.7)", () => {
  it.each([
    ["not connected", notConnected, en.pageNotConnected],
    ["off", { ...connected, enabled: false }, en.pageOff],
  ])("shows an alert with the settings link and no lists when %s", async (_, view, message) => {
    const host = setup(view as View);
    const c = await render(host);
    const alert = byTestId(c, "backlog-page-alert")!;
    expect(alert.getAttribute("data-host")).toBe("Alert");
    expect(alert.textContent).toContain(message);
    expect(byTestId(c, "backlog-scope-bar")).toBeNull();
    const link = byTestId(c, "backlog-page-settings-link")!;
    expect(link.getAttribute("data-variant")).toBe("link");
    await act(async () => link.click());
    expect(host.navigate).toHaveBeenCalledWith(SETTINGS_HREF);
  });
});

describe("/backlog scopes (BR2.2, FR2.3)", () => {
  it("opens Issues by default and switches to Pull requests in the URL", async () => {
    const host = setup();
    const c = await render(host);
    expect(byTestId(c, "backlog-scope-bar-kind-issues")!.getAttribute("aria-pressed")).toBe("true");
    expect(byTestId(c, "backlog-issues")).not.toBeNull();
    expect(byTestId(c, "backlog-prs")).toBeNull();
    await openPRs(c);
    expect(byTestId(c, "backlog-prs")).not.toBeNull();
    expect(host.navigate).toHaveBeenCalledWith("/backlog?scope=prs", { replace: true });
  });

  it("opens Pull requests when the URL says scope=prs", async () => {
    window.history.replaceState(null, "", "/backlog?scope=prs");
    const c = await render(setup());
    expect(byTestId(c, "backlog-scope-bar-kind-prs")!.getAttribute("aria-pressed")).toBe("true");
    expect(byTestId(c, "backlog-prs")).not.toBeNull();
  });
});

describe("Pull requests list (FR2.5, FR2.6, FR4, BR2.4, BR4.1)", () => {
  it("opens on the first repository, lists 20 rows with the linked task and pages", async () => {
    const host = setup();
    const c = await render(host);
    await openPRs(c);
    expect(calls(host, "git.prs.list")).toHaveLength(1);
    expect(calls(host, "git.prs.list")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: {
        projectKey: "PROJ",
        repoName: "web-app",
        statuses: ["open"],
        assignee: "me",
        creator: "anyone",
        page: 1,
      },
    });
    expect(c.querySelectorAll('[data-testid^="backlog-pr-row-"][data-host="ChangeRequestRow"]')).toHaveLength(
      20,
    );
    const row = byTestId(c, "backlog-pr-row-45")!;
    expect(row.textContent).toContain("Change 45");
    expect(row.textContent).toContain("#45");
    expect(row.textContent).toContain("Test User");
    expect(row.textContent).toContain("rel:2026-10-02T09:00:00Z");
    expect(byTestId(c, "backlog-pr-row-45-link")!.getAttribute("href")).toBe(pr(45).url);
    expect(byTestId(c, "backlog-pr-task-45-task-9")!.getAttribute("href")).toBe("/t/task-9");
    expect(byTestId(c, "backlog-prs-showing")!.textContent).toBe("Showing 1-20 of 45");
    expect(byTestId(c, "backlog-prs-count")!.textContent).toContain("45");
    expect((byTestId(c, "backlog-prs-prev") as HTMLButtonElement).disabled).toBe(true);
    await act(async () => byTestId(c, "backlog-prs-next")!.click());
    await act(async () => byTestId(c, "backlog-prs-next")!.click());
    expect(calls(host, "git.prs.list").at(-1)![1]).toMatchObject({ body: { page: 3 } });
    expect(byTestId(c, "backlog-prs-showing")!.textContent).toBe("Showing 41-45 of 45");
    expect((byTestId(c, "backlog-prs-next") as HTMLButtonElement).disabled).toBe(true);
  });

  it("filters by status, assignee and creator and resets the page", async () => {
    const host = setup();
    const c = await render(host);
    await openPRs(c);
    await pickRepo(c);
    await act(async () => byTestId(c, "backlog-prs-next")!.click());
    await act(async () => byTestId(c, "backlog-prs-status-merged")!.click());
    await act(async () => choose(c, "backlog-prs-assignee", "me"));
    await act(async () => choose(c, "backlog-prs-creator", "me"));
    expect(calls(host, "git.prs.list").at(-1)![1]).toMatchObject({
      body: { statuses: ["open", "merged"], assignee: "me", creator: "me", page: 1 },
    });
  });

  it("saves the current filters as a query from the list (BR2.5)", async () => {
    const save = vi.fn(async (body: Record<string, unknown>) => ({ id: "q3", ...body }));
    const host = setup(VIEW, { "git.queries.save": save });
    const c = await render(host);
    await openPRs(c);
    const button = byTestId(c, "backlog-prs-save-query")!;
    expect(button.getAttribute("data-variant")).toBe("outline");
    expect(button.getAttribute("data-size")).toBe("sm");
    await act(async () => button.click());
    await act(async () => setValue(byTestId(c, "backlog-save-query-name") as HTMLInputElement, "Open PRs"));
    await act(async () => byTestId(c, "backlog-save-query-save")!.click());
    expect(save).toHaveBeenCalledWith({
      name: "Open PRs",
      projectKey: "PROJ",
      repoName: "web-app",
      statuses: ["open"],
      assignee: "me",
      creator: "anyone",
    });
    expect(byTestId(c, "backlog-save-query-dialog")).toBeNull();
    expect(byTestId(c, "backlog-saved-menu-item-q3")).not.toBeNull();
  });

  it("shows empty and error states with Retry, asking to reconnect on 401 (BR7.1, BR7.2)", async () => {
    let mode: "empty" | "401" | "429" = "empty";
    const host = setup(VIEW, {
      "git.prs.list": async () => {
        if (mode === "401") throw actionError(401, { code: "reconnect_required" });
        if (mode === "429") throw actionError(429, { code: "rate_limited", retryAfterSeconds: 9 });
        return { items: [], page: 1, pageSize: 20, total: 0, hasNext: false };
      },
    });
    const c = await render(host);
    await openPRs(c);
    await pickRepo(c);
    expect(byTestId(c, "backlog-prs-empty")!.textContent).toContain(en.prsEmpty);
    mode = "401";
    await act(async () => byTestId(c, "backlog-prs-refresh")!.click());
    expect(byTestId(c, "backlog-prs-error")!.textContent).toContain(en.reconnectRequired);
    mode = "429";
    await act(async () => byTestId(c, "backlog-prs-retry")!.click());
    expect(byTestId(c, "backlog-prs-error")!.textContent).toContain("Try again in 9 s");
    expect(text(c)).not.toContain("SECRET");
  });

  it("uses host controls, has accessible icon buttons and test ids (BR5.1, BR5.2, NFR4)", async () => {
    const c = await render(setup());
    expect(rawControls(c)).toEqual([]);
    await openPRs(c);
    await pickRepo(c);
    expect(rawControls(c)).toEqual([]);
    const refresh = byTestId(c, "backlog-prs-refresh")!;
    expect(refresh.getAttribute("data-variant")).toBe("ghost");
    expect(refresh.getAttribute("data-size")).toBe("icon");
    expect(refresh.getAttribute("aria-label")).toBe(en.refresh);
    expect(byTestId(c, "backlog-prs-next")!.getAttribute("data-variant")).toBe("outline");
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });
});

const STATUSES = [
  { id: 1, name: "Open" },
  { id: 2, name: "In Progress" },
  { id: 3, name: "Resolved" },
  { id: 4, name: "Closed" },
];
const ISSUE_QUERY = {
  id: "iq1",
  name: "My bugs",
  projectKey: "PROJ",
  statusIds: [2],
  assignee: "7",
  keyword: "login",
  isDefault: true,
};

/** A host for the default-query tests: saved lists, statuses and quick actions are scripted. */
function defaults(
  opts: { issueQueries?: unknown[]; prQueries?: unknown[]; repos?: unknown[] } = {},
  handlers: Record<string, Handler> = {},
) {
  return setup(VIEW, {
    "issues.filters": async () => ({
      projects: [{ key: "PROJ", name: "Test Project" }],
      statuses: STATUSES,
      assignees: [{ id: 7, name: "Lan" }],
    }),
    "issues.queries.list": async () => ({ queries: opts.issueQueries ?? [] }),
    "git.queries.list": async () => ({ queries: opts.prQueries ?? QUERIES }),
    "issues.quick_actions.get": async () => ({ issue: [], pr: [] }),
    ...(opts.repos ? { "git.repositories.list": async () => ({ repositories: opts.repos }) } : {}),
    ...handlers,
  });
}

describe("Default queries and scope bar (FR3, FR4, FR5.1, NFR4)", () => {
  it("replaces the tabs with the host scope bar holding the kinds, a preset and the saved menu", async () => {
    const c = await render(defaults());
    expect(byTestId(c, "backlog-scope-tabs")).toBeNull();
    const bar = byTestId(c, "backlog-scope-bar")!;
    expect(bar.getAttribute("data-host")).toBe("IntegrationScopeBar");
    expect(byTestId(c, "backlog-scope-bar-kind-issues")!.getAttribute("aria-pressed")).toBe("true");
    expect(byTestId(c, "backlog-scope-bar-preset-assigned-open")!.textContent).toBe(en.presetAssignedOpen);
    expect(byTestId(c, "backlog-saved-menu")).not.toBeNull();
  });

  it("opens Issues on 'Assigned to me, open' without a starred issue query, with one list call", async () => {
    const host = defaults();
    await render(host);
    expect(calls(host, "issues.list")).toHaveLength(1);
    expect(calls(host, "issues.list")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { page: 1, pageSize: 20, keyword: "", statusIds: [1, 2, 3], assignee: "me" },
    });
  });

  it("opens Issues on the starred issue query instead", async () => {
    const host = defaults({ issueQueries: [ISSUE_QUERY] });
    const c = await render(host);
    expect(calls(host, "issues.list")).toHaveLength(1);
    expect(calls(host, "issues.list")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: {
        page: 1,
        pageSize: 20,
        keyword: "login",
        projectKeys: ["PROJ"],
        statusIds: [2],
        assigneeIds: [7],
      },
    });
    expect(byTestId(c, "backlog-saved-menu-item-iq1")!.getAttribute("aria-pressed")).toBe("true");
  });

  it("opens Pull requests on 'Open, assigned to me' in the first repository without a starred PR query", async () => {
    const host = defaults();
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scope-bar-kind-prs")!.click());
    expect(calls(host, "git.prs.list")).toHaveLength(1);
    expect(calls(host, "git.prs.list")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: {
        projectKey: "PROJ",
        repoName: "web-app",
        statuses: ["open"],
        assignee: "me",
        creator: "anyone",
        page: 1,
      },
    });
    expect(byTestId(c, "backlog-scope-bar-preset-open-assigned")!.getAttribute("aria-pressed")).toBe("true");
  });

  it("opens Pull requests on the starred PR query instead", async () => {
    const host = defaults({ prQueries: [{ ...QUERIES[0], isDefault: true }, QUERIES[1]] });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scope-bar-kind-prs")!.click());
    expect(calls(host, "git.prs.list")).toHaveLength(1);
    expect(calls(host, "git.prs.list")[0]![1]).toMatchObject({
      body: {
        projectKey: "PROJ",
        repoName: "web-app",
        statuses: ["merged"],
        assignee: "me",
        creator: "anyone",
      },
    });
  });

  it("shows the repository guidance when the projects have no repository (FR3.4)", async () => {
    const host = defaults({ repos: [] });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scope-bar-kind-prs")!.click());
    expect(byTestId(c, "backlog-prs-empty")!.textContent).toContain(en.prsChooseRepository);
    expect(calls(host, "git.prs.list")).toHaveLength(0);
  });

  it("stars and un-stars saved queries through set_default", async () => {
    const host = defaults(
      { issueQueries: [{ ...ISSUE_QUERY, isDefault: false }] },
      {
        "issues.queries.set_default": async (b) => ({
          queries: [{ ...ISSUE_QUERY, isDefault: b.isDefault }],
        }),
        "git.queries.set_default": async (b) => ({
          queries: [{ ...QUERIES[0], isDefault: b.isDefault }, QUERIES[1]],
        }),
      },
    );
    const c = await render(host);
    await act(async () => byTestId(c, "saved-query-default-iq1")!.click());
    expect(calls(host, "issues.queries.set_default")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { id: "iq1", isDefault: true },
    });
    expect(byTestId(c, "saved-query-default-iq1")!.getAttribute("aria-pressed")).toBe("true");
    await act(async () => byTestId(c, "backlog-scope-bar-kind-prs")!.click());
    await act(async () => byTestId(c, "saved-query-default-q1")!.click());
    expect(calls(host, "git.queries.set_default")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { id: "q1", isDefault: true },
    });
  });

  it("applies a saved query from the menu and ignores one of an unselected project (BR2.4)", async () => {
    const host = defaults();
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scope-bar-kind-prs")!.click());
    expect(byTestId(c, "backlog-saved-menu-item-q2")!.textContent).toContain(en.projectNotSelected);
    await act(async () => byTestId(c, "backlog-saved-menu-item-q2")!.click());
    expect(calls(host, "git.prs.list")).toHaveLength(1);
    await act(async () => byTestId(c, "backlog-saved-menu-item-q1")!.click());
    expect(calls(host, "git.prs.list").at(-1)![1]).toMatchObject({
      body: { statuses: ["merged"], assignee: "me" },
    });
  });

  it("saves the issue filters as a saved issue query and lists it in the menu (FR4.3)", async () => {
    const save = vi.fn(async (b: Record<string, unknown>) => ({ id: "iq9", ...b }));
    const host = defaults({}, { "issues.queries.save": save });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-issues-save-query")!.click());
    await act(async () => setValue(byTestId(c, "backlog-save-query-name") as HTMLInputElement, "Mine open"));
    await act(async () => byTestId(c, "backlog-save-query-save")!.click());
    expect(save).toHaveBeenCalledWith({
      name: "Mine open",
      projectKey: "",
      statusIds: [1, 2, 3],
      assignee: "me",
      keyword: "",
    });
    expect(byTestId(c, "backlog-saved-menu-item-iq9")!.textContent).toBe("Mine open");
    await act(async () => byTestId(c, "backlog-saved-menu-save")!.click());
    expect(byTestId(c, "backlog-save-query-dialog")).not.toBeNull();
  });

  it("deletes a saved query from the menu and falls back to the preset", async () => {
    const del = vi.fn(async () => ({ ok: true }));
    const host = defaults({ issueQueries: [ISSUE_QUERY] }, { "issues.queries.delete": del });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-saved-menu-delete-iq1")!.click());
    expect(del).toHaveBeenCalledWith({ id: "iq1" });
    expect(byTestId(c, "backlog-saved-menu-item-iq1")).toBeNull();
    expect(byTestId(c, "backlog-scope-bar-preset-assigned-open")!.getAttribute("aria-pressed")).toBe("true");
    expect(calls(host, "issues.list").at(-1)![1]).toMatchObject({
      body: { assignee: "me", statusIds: [1, 2, 3] },
    });
  });
});

describe("GitHub-aligned layout (FR5.1-FR5.3)", () => {
  it("puts the host scope bar first, a bordered toolbar with refresh last, and padded results", async () => {
    const c = await render(defaults());
    const root = byTestId(c, "backlog-page")!;
    expect(root.firstElementChild).toBe(byTestId(c, "backlog-scope-bar"));
    expect(root.className).not.toMatch(/\bp[xy]?-\d/);
    for (const kind of ["issues", "prs"]) {
      if (kind === "prs") await act(async () => byTestId(c, "backlog-scope-bar-kind-prs")!.click());
      const toolbar = byTestId(c, `backlog-${kind}-toolbar`)!;
      for (const cls of ["border-b", "px-4", "py-2.5", "sm:px-6"]) expect(toolbar.classList).toContain(cls);
      expect(toolbar.lastElementChild!.lastElementChild).toBe(byTestId(c, `backlog-${kind}-refresh`));
      const results = byTestId(c, `backlog-${kind}-results`)!;
      for (const cls of ["px-3", "py-4", "md:px-6"]) expect(results.classList).toContain(cls);
    }
  });
});
