import { act } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { createSettingsScreen } from "./SettingsScreen";
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
  optionValues,
  rawControls,
  setValue,
  text,
  unmount,
  type View,
} from "../testing/harness";

const SECTIONS = [
  "backlog-section-connection",
  "backlog-section-pr-watches",
  "backlog-section-issue-watches",
  "backlog-section-saved-queries",
  "backlog-section-issue-sync",
  "backlog-section-git-access",
  "backlog-section-projects",
];

const PR_WATCH = {
  id: "w1",
  name: "Reviews",
  projectKey: "PROJ",
  repoName: "web-app",
  statuses: ["open"],
  assignee: "anyone",
  creator: "anyone",
  state: "active",
  createdCount: 2,
  pendingCount: 0,
};
const ISSUE_WATCH = {
  id: "iw1",
  name: "Open bugs",
  projectKey: "PROJ",
  statusIds: [1],
  assignee: "anyone",
  creator: "me",
  workflowId: "wf-1",
  intervalMinutes: 5,
  state: "active",
  createdCount: 3,
  pendingCount: 1,
  lastError: "workflow_missing",
};
const QUERY = {
  id: "q1",
  name: "Merged by me",
  projectKey: "PROJ",
  repoName: "web-app",
  statuses: ["merged"],
  assignee: "me",
  creator: "anyone",
};

type Handler = (body: Record<string, unknown>) => Promise<unknown>;

function scripted(view: View, handlers: Record<string, Handler> = {}, lists: Record<string, unknown[]> = {}) {
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    switch (key) {
      case "connection.get":
        return view;
      case "connection.list_projects":
        return {
          projects: [{ projectKey: "PROJ", projectId: 101, projectName: "Test Project", selected: true }],
        };
      case "issues.settings.get":
        return { pollMinutes: 5 };
      case "git.watches.list":
        return { watches: lists.prWatches ?? [PR_WATCH] };
      case "issues.watches.list":
        return { watches: lists.issueWatches ?? [ISSUE_WATCH] };
      case "git.queries.list":
        return { queries: lists.queries ?? [QUERY] };
      case "issues.filters":
        return {
          projects: [{ key: "PROJ", name: "Test Project" }],
          statuses: [
            { id: 1, name: "Open" },
            { id: 2, name: "In Progress" },
          ],
          assignees: [],
        };
      case "git.repositories.list":
        return { repositories: [{ ownerOrProject: "PROJ", repositoryName: "web-app", repositoryId: "11" }] };
      case "git.impact":
        return { prLinks: 0, prWatches: 0 };
      case "issues.impact":
        return { issueLinks: 0 };
    }
    throw new Error(`unexpected ${key}`);
  });
}

const render = (host: ReturnType<typeof scripted>) =>
  mount(createSettingsScreen(host), { workspaceId: "ws-1" });
const calls = (host: ReturnType<typeof scripted>, key: string) =>
  vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);
const sections = (c: HTMLElement) =>
  [...c.querySelectorAll('[data-testid^="backlog-section-"]')].map((el) => el.getAttribute("data-testid"));

beforeEach(() => {
  Element.prototype.scrollIntoView = vi.fn();
});
afterEach(unmount);

describe("Settings sections (FR1, BR1.1, BR1.2)", () => {
  it("stacks the seven sections in order when connected", async () => {
    const c = await render(scripted(connected));
    expect(sections(c)).toEqual(SECTIONS);
    expect(byTestId(c, "backlog-settings")!.className).toContain("gap-8");
  });

  it("shows only the Connection section when not connected", async () => {
    const c = await render(scripted(notConnected));
    expect(sections(c)).toEqual(["backlog-section-connection"]);
  });

  it("keeps the watch and query sections in the member view (BR1.2)", async () => {
    const host = scripted(connected, {
      "connection.connect_api_key": async () => {
        throw actionError(403, {});
      },
    });
    const c = await render(host);
    await act(async () => {
      setValue(byTestId(c, "backlog-space-url") as HTMLInputElement, connected.spaceHost!);
      setValue(byTestId(c, "backlog-api-key") as HTMLInputElement, "k-1");
    });
    await act(async () => byTestId(c, "backlog-connect")!.click());
    await act(async () => byTestId(c, "backlog-replace-dialog-confirm")!.click());
    expect(sections(c)).toEqual([
      "backlog-section-connection",
      "backlog-section-pr-watches",
      "backlog-section-issue-watches",
      "backlog-section-saved-queries",
      "backlog-section-issue-sync",
    ]);
    expect(byTestId(c, "backlog-pr-watches-add")).not.toBeNull();
    expect(byTestId(c, "backlog-poll-save")).toBeNull();
    expect(byTestId(c, "backlog-disconnect")).toBeNull();
  });

  it("chooses the sign-in method with a dropdown (BR1.4)", async () => {
    const c = await render(scripted(notConnected));
    expect(optionValues(c, "backlog-method")).toEqual(["api_key", "oauth"]);
    expect(c.querySelector('label[for="backlog-method"]')!.textContent).toBe(en.signInMethodLabel);
    expect(byTestId(c, "backlog-api-key")).not.toBeNull();
    await act(async () => choose(c, "backlog-method", "oauth"));
    expect(byTestId(c, "backlog-api-key")).toBeNull();
    expect(byTestId(c, "backlog-sign-in-nulab")).not.toBeNull();
  });

  it("points the restore notice at the PR watches section (BR1.5)", async () => {
    const host = scripted(notConnected, {
      "connection.connect_api_key": async () => ({ ...connected, restored: true }),
    });
    const c = await render(host);
    await act(async () => {
      setValue(byTestId(c, "backlog-space-url") as HTMLInputElement, connected.spaceHost!);
      setValue(byTestId(c, "backlog-api-key") as HTMLInputElement, "k-1");
    });
    await act(async () => byTestId(c, "backlog-connect")!.click());
    const review = byTestId(c, "backlog-review-watches")!;
    expect(review.getAttribute("data-variant")).toBe("link");
    await act(async () => review.click());
    expect(Element.prototype.scrollIntoView).toHaveBeenCalled();
    expect(vi.mocked(Element.prototype.scrollIntoView).mock.contexts[0]).toBe(
      document.getElementById("backlog-pr-watches"),
    );
    expect(host.navigate).not.toHaveBeenCalled();
  });
});

describe("PR watches section (FR1.2)", () => {
  it("lists watches in a table with row actions, adds in a dialog and deletes after asking", async () => {
    const host = scripted(connected, {
      "git.watches.delete": async () => ({ ok: true }),
      "git.watches.run": async () => ({ queued: true }),
    });
    const c = await render(host);
    const row = byTestId(c, "backlog-pr-watch-row-w1")!;
    expect(row.tagName).toBe("TR");
    expect(row.textContent).toContain("Reviews");
    expect(row.textContent).toContain("PROJ / web-app");
    expect(row.textContent).toContain(en.watchActive);
    const menu = byTestId(c, "backlog-pr-watch-menu-w1")!;
    expect(menu.getAttribute("aria-label")).toBe("Actions for Reviews");
    expect(menu.getAttribute("data-variant")).toBe("ghost");

    const add = byTestId(c, "backlog-pr-watches-add")!;
    expect(add.getAttribute("data-size")).toBe("sm");
    await act(async () => add.click());
    expect(byTestId(c, "backlog-pr-watch-dialog")!.getAttribute("role")).toBe("dialog");
    expect(byTestId(c, "backlog-watch-name")).not.toBeNull();
    await act(async () => byTestId(c, "backlog-watch-cancel")!.click());
    expect(byTestId(c, "backlog-pr-watch-dialog")).toBeNull();

    await act(async () => byTestId(c, "backlog-pr-watch-run-w1")!.click());
    expect(calls(host, "git.watches.run")[0]![1]).toEqual({ workspaceId: "ws-1", body: { id: "w1" } });
    await act(async () => byTestId(c, "backlog-pr-watch-delete-w1")!.click());
    await act(async () => byTestId(c, "backlog-pr-watch-delete-dialog-confirm")!.click());
    expect(calls(host, "git.watches.delete")).toHaveLength(1);
    expect(byTestId(c, "backlog-pr-watch-row-w1")).toBeNull();
    expect(byTestId(c, "backlog-pr-watches-empty")!.textContent).toContain(en.watchesEmpty);
  });

  it("shows an inline error with Retry when the list fails (BR7.2)", async () => {
    let fail = true;
    const c = await render(
      scripted(connected, {
        "git.watches.list": async () => {
          if (fail) throw actionError(503, { code: "unreachable" });
          return { watches: [PR_WATCH] };
        },
      }),
    );
    expect(byTestId(c, "backlog-pr-watches-error")!.textContent).toContain(en.unreachable);
    expect(sections(c)).toEqual(SECTIONS);
    fail = false;
    await act(async () => byTestId(c, "backlog-pr-watches-retry")!.click());
    expect(byTestId(c, "backlog-pr-watch-row-w1")).not.toBeNull();
  });
});

describe("Issue watches section (FR1.3, FR3)", () => {
  it("lists watches with their last error", async () => {
    const c = await render(scripted(connected));
    const row = byTestId(c, "backlog-issue-watch-row-iw1")!;
    expect(row.textContent).toContain("Open bugs");
    expect(row.textContent).toContain("Test Project");
    expect(row.textContent).toContain("Open");
    expect(row.textContent).toContain(en.lastErrorWorkflowMissing);
  });

  it("adds a watch in a dialog with the default interval 5 and field errors (BR3.1)", async () => {
    let fail = true;
    const save = vi.fn(async (body: Record<string, unknown>) => {
      if (fail) throw actionError(400, { code: "validation", field: "intervalMinutes" });
      return { id: "iw2", state: "active", createdCount: 0, pendingCount: 0, ...body };
    });
    const host = scripted(connected, { "issues.watches.save": save }, { issueWatches: [] });
    const c = await render(host);
    expect(byTestId(c, "backlog-issue-watches-empty")!.textContent).toContain(en.watchesEmpty);
    await act(async () => byTestId(c, "backlog-issue-watches-add")!.click());
    expect(byTestId(c, "backlog-issue-watch-dialog")!.getAttribute("role")).toBe("dialog");
    expect((byTestId(c, "backlog-issue-watch-interval") as HTMLInputElement).value).toBe("5");
    await act(async () => byTestId(c, "backlog-issue-watch-save")!.click());
    expect(byTestId(c, "backlog-issue-watch-name-error")!.textContent).toBe(en.errorName);
    expect(save).not.toHaveBeenCalled();

    await act(async () => setValue(byTestId(c, "backlog-issue-watch-name") as HTMLInputElement, "Bugs"));
    await act(async () => choose(c, "backlog-issue-watch-project", "PROJ"));
    await act(async () => byTestId(c, "backlog-issue-watch-status-1")!.click());
    await act(async () => byTestId(c, "backlog-issue-watch-save")!.click());
    expect(save).toHaveBeenCalledWith({
      name: "Bugs",
      projectKey: "PROJ",
      statusIds: [1],
      assignee: "anyone",
      creator: "anyone",
      intervalMinutes: 5,
      workflowId: "wf-1",
      workflowStepId: "step-1",
    });
    expect(byTestId(c, "backlog-issue-watch-interval-error")!.textContent).toBe(en.errorInterval);
    fail = false;
    await act(async () => byTestId(c, "backlog-issue-watch-save")!.click());
    expect(byTestId(c, "backlog-issue-watch-dialog")).toBeNull();
    expect(byTestId(c, "backlog-issue-watch-row-iw2")).not.toBeNull();
  });
});

describe("Saved PR queries section (WF6, BR2.5)", () => {
  it("renames a query in the dialog and deletes after asking", async () => {
    const save = vi.fn(async (body: Record<string, unknown>) => body);
    const host = scripted(connected, {
      "git.queries.save": save,
      "git.queries.delete": async () => ({ ok: true }),
    });
    const c = await render(host);
    const row = byTestId(c, "backlog-saved-query-row-q1")!;
    expect(row.textContent).toContain("Merged by me");
    expect(row.textContent).toContain("PROJ / web-app");
    await act(async () => byTestId(c, "backlog-saved-query-edit-q1")!.click());
    const name = byTestId(c, "backlog-save-query-name") as HTMLInputElement;
    expect(name.value).toBe("Merged by me");
    await act(async () => setValue(name, "Merged"));
    await act(async () => byTestId(c, "backlog-save-query-save")!.click());
    expect(save).toHaveBeenCalledWith({ ...QUERY, name: "Merged" });
    expect(byTestId(c, "backlog-saved-query-row-q1")!.textContent).toContain("Merged");

    await act(async () => byTestId(c, "backlog-saved-query-delete-q1")!.click());
    await act(async () => byTestId(c, "backlog-saved-query-delete-dialog-confirm")!.click());
    expect(calls(host, "git.queries.delete")[0]![1]).toEqual({ workspaceId: "ws-1", body: { id: "q1" } });
    expect(byTestId(c, "backlog-saved-queries-empty")!.textContent).toContain(en.savedQueriesEmpty);
  });
});

describe("Settings controls (BR5.1, BR5.2, NFR4)", () => {
  it("uses host controls with the GitHub button style, is accessible and has test ids", async () => {
    const c = await render(scripted(connected));
    expect(rawControls(c)).toEqual([]);
    expect(byTestId(c, "backlog-disconnect")!.getAttribute("data-variant")).toBe("destructive");
    expect(byTestId(c, "backlog-test-connection")!.getAttribute("data-variant")).toBe("outline");
    expect(byTestId(c, "backlog-issue-watches-add")!.getAttribute("data-size")).toBe("sm");
    for (const el of c.querySelectorAll('[data-host="Button"]')) {
      expect(el.className, el.outerHTML.slice(0, 80)).toContain("cursor-pointer");
      if (!(el.textContent ?? "").trim())
        expect(el.getAttribute("aria-label"), el.outerHTML.slice(0, 80)).toBeTruthy();
    }
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
    expect(text(c)).not.toContain("undefined");
  });
});
