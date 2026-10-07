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
  await act(async () => byTestId(c, "backlog-scope-prs")!.click());
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
    expect(byTestId(c, "backlog-scope-tabs")).toBeNull();
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
    expect(byTestId(c, "backlog-scope-issues")!.getAttribute("aria-selected")).toBe("true");
    expect(byTestId(c, "backlog-issues")).not.toBeNull();
    expect(byTestId(c, "backlog-prs")).toBeNull();
    await openPRs(c);
    expect(byTestId(c, "backlog-prs")).not.toBeNull();
    expect(host.navigate).toHaveBeenCalledWith("/backlog?scope=prs", { replace: true });
  });

  it("opens Pull requests when the URL says scope=prs", async () => {
    window.history.replaceState(null, "", "/backlog?scope=prs");
    const c = await render(setup());
    expect(byTestId(c, "backlog-scope-prs")!.getAttribute("aria-selected")).toBe("true");
    expect(byTestId(c, "backlog-prs")).not.toBeNull();
  });
});

describe("Pull requests list (FR2.5, FR2.6, FR4, BR2.4, BR4.1)", () => {
  it("asks for a repository first, then lists 20 rows with the linked task and pages", async () => {
    const host = setup();
    const c = await render(host);
    await openPRs(c);
    expect(byTestId(c, "backlog-prs-empty")!.textContent).toContain(en.prsChooseRepository);
    expect(calls(host, "git.prs.list")).toHaveLength(0);
    await pickRepo(c);
    expect(calls(host, "git.prs.list")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: {
        projectKey: "PROJ",
        repoName: "web-app",
        statuses: ["open"],
        assignee: "anyone",
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

  it("applies a saved query as a preset; a query of an unselected project is disabled (BR2.4)", async () => {
    const host = setup();
    const c = await render(host);
    await openPRs(c);
    const old = c.querySelector<HTMLElement>('[role="option"][data-value="q2"]')!;
    expect(old.getAttribute("aria-disabled")).toBe("true");
    expect(old.textContent).toContain(en.projectNotSelected);
    await act(async () => choose(c, "backlog-prs-preset", "q1"));
    expect(calls(host, "git.prs.list").at(-1)![1]).toEqual({
      workspaceId: "ws-1",
      body: {
        projectKey: "PROJ",
        repoName: "web-app",
        statuses: ["merged"],
        assignee: "me",
        creator: "anyone",
        page: 1,
      },
    });
  });

  it("saves the current filters as a query from the list (BR2.5)", async () => {
    const save = vi.fn(async (body: Record<string, unknown>) => ({ id: "q3", ...body }));
    const host = setup(VIEW, { "git.queries.save": save });
    const c = await render(host);
    await openPRs(c);
    expect(byTestId(c, "backlog-prs-save-query")!.hasAttribute("disabled")).toBe(true);
    await pickRepo(c);
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
      assignee: "anyone",
      creator: "anyone",
    });
    expect(byTestId(c, "backlog-save-query-dialog")).toBeNull();
    expect(c.querySelector('[role="option"][data-value="q3"]')).not.toBeNull();
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
