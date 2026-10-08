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

// FR4.1: Projects right after Connection, then the rest in their order.
const SECTIONS = [
  "backlog-section-connection",
  "backlog-section-projects",
  "backlog-section-pr-watches",
  "backlog-section-issue-watches",
  "backlog-section-saved-queries",
  "backlog-section-quick-actions",
  "backlog-section-issue-sync",
  "backlog-section-source-control",
];

const DEFAULT_ACTIONS = {
  issue: [
    {
      id: "implement",
      label: "Implement",
      hint: "Build and open a PR",
      icon: "code",
      promptTemplate: "Implement {{url}}",
    },
    {
      id: "investigate",
      label: "Investigate",
      hint: "Find the root cause",
      icon: "search",
      promptTemplate: "Investigate {{url}}",
    },
    {
      id: "reproduce",
      label: "Reproduce",
      hint: "Document repro steps",
      icon: "bug",
      promptTemplate: "Reproduce {{url}}",
    },
  ],
  pr: [
    { id: "review", label: "Review", hint: "Read the diff", icon: "eye", promptTemplate: "Review {{url}}" },
  ],
};

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

const ISSUE_QUERY = {
  id: "iq1",
  name: "My bugs",
  projectKey: "PROJ",
  statusIds: [1],
  assignee: "me",
  keyword: "login",
};

const SCM_PROVIDERS = [
  {
    provider: "github",
    state: "connected",
    account: "Lan",
    mappings: [{ projectKey: "PROJ", repos: ["acme/web"] }],
  },
  { provider: "gitlab", state: "not_configured", mappings: [] },
  { provider: "bitbucket", state: "not_configured", mappings: [] },
];
const SCM_WATCH = {
  id: "s1",
  name: "GitHub reviews",
  provider: "github",
  projectKey: "PROJ",
  repo: "acme/web",
  statuses: ["open"],
  author: "anyone",
  intervalMinutes: 5,
  state: "active",
  createdCount: 1,
  unmapped: true,
};
const SCM_QUERY = {
  id: "sq1",
  name: "GitHub open",
  provider: "github",
  projectKey: "PROJ",
  repo: "acme/web",
  statuses: ["open"],
  author: "me",
  isDefault: true,
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
      case "issues.quick_actions.get":
        return DEFAULT_ACTIONS;
      case "issues.queries.list":
        return { queries: lists.issueQueries ?? [ISSUE_QUERY] };
      case "scm.providers.list":
        return { providers: SCM_PROVIDERS, active: "" }; // pending: every provider, as before (FR3.3)
      case "scm.watches.list":
        return { watches: lists.scmWatches ?? [] };
      case "scm.queries.list":
        return { queries: lists.scmQueries ?? [] };
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
  it("stacks the eight sections in order when connected", async () => {
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
      "backlog-section-quick-actions",
      "backlog-section-issue-sync",
      "backlog-section-source-control",
    ]); // FR2.6: members see the source control settings read-only
    expect(byTestId(c, "backlog-scm-github-save")).toBeNull();
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

describe("Source control watches and queries (FR3.4, FR4.2, FR4.3)", () => {
  it("lists provider watches with their provider and the unmapped state, acting through scm.watches", async () => {
    const host = scripted(
      connected,
      { "scm.watches.pause": async () => ({ ...SCM_WATCH, state: "paused" }) },
      {
        scmWatches: [SCM_WATCH],
      },
    );
    const c = await render(host);
    const row = byTestId(c, "backlog-pr-watch-row-s1")!;
    expect(text(row)).toContain("GitHub · PROJ · acme/web");
    expect(byTestId(row, "backlog-pr-watch-unmapped-s1")!.textContent).toBe(en.scmUnmapped);
    expect(byTestId(c, "backlog-pr-watch-row-w1"), "Backlog Git watches stay").not.toBeNull();
    await act(async () => byTestId(c, "backlog-pr-watch-pause-s1")!.click());
    expect(calls(host, "scm.watches.pause")[0]![1]).toEqual({ workspaceId: "ws-1", body: { id: "s1" } });
  });

  it("lists provider saved queries in their own table with the provider (FR4.2)", async () => {
    const host = scripted(connected, {}, { scmQueries: [SCM_QUERY] });
    const c = await render(host);
    const row = byTestId(c, "backlog-saved-scm-query-row-sq1")!;
    expect(text(row)).toContain("GitHub open");
    expect(text(row)).toContain("GitHub · PROJ · acme/web · Open · author Me");
    expect(text(row)).toContain(en.defaultQuery);
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

describe("Empty watch lists (FR4.2, FR4.3)", () => {
  it("says No issue watches yet and has only the header Add watch button", async () => {
    const c = await render(scripted(connected, {}, { issueWatches: [] }));
    expect(byTestId(c, "backlog-issue-watches-empty")!.textContent).toContain("No issue watches yet");
    expect(byTestId(c, "backlog-issue-watches-empty")!.textContent).not.toContain(en.watchesEmpty);
    const section = byTestId(c, "backlog-section-issue-watches")!;
    expect([...section.querySelectorAll("button")].filter((b) => b.textContent === en.addWatch)).toHaveLength(
      1,
    );
    expect(byTestId(c, "backlog-issue-watches-add")).not.toBeNull();
    expect(byTestId(c, "backlog-issue-watches-empty-add")).toBeNull();
  });

  it("says No PR watches yet and has only the header Add watch button", async () => {
    const c = await render(scripted(connected, {}, { prWatches: [] }));
    expect(byTestId(c, "backlog-pr-watches-empty")!.textContent).toContain(en.watchesEmpty);
    const section = byTestId(c, "backlog-section-pr-watches")!;
    expect([...section.querySelectorAll("button")].filter((b) => b.textContent === en.addWatch)).toHaveLength(
      1,
    );
    expect(byTestId(c, "backlog-pr-watches-add")).not.toBeNull();
    expect(byTestId(c, "backlog-pr-watches-empty-add")).toBeNull();
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
    expect(byTestId(c, "backlog-issue-watches-empty")!.textContent).toContain(en.issueWatchesEmpty);
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

describe("Saved issue queries in Settings (FR4.3, FR4.4, FR3.3)", () => {
  it("lists issue queries next to PR queries, renames, stars and deletes them", async () => {
    const save = vi.fn(async (body: Record<string, unknown>) => body);
    const host = scripted(connected, {
      "issues.queries.save": save,
      "issues.queries.delete": async () => ({ ok: true }),
      "issues.queries.set_default": async (b) => ({ queries: [{ ...ISSUE_QUERY, isDefault: b.isDefault }] }),
      "git.queries.set_default": async (b) => ({ queries: [{ ...QUERY, isDefault: b.isDefault }] }),
    });
    const c = await render(host);
    const row = byTestId(c, "backlog-saved-issue-query-row-iq1")!;
    expect(row.textContent).toContain("My bugs");
    expect(row.textContent).toContain("PROJ");
    expect(row.textContent).toContain(en.whoMe);

    await act(async () => byTestId(c, "backlog-saved-issue-query-star-iq1")!.click());
    expect(calls(host, "issues.queries.set_default")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { id: "iq1", isDefault: true },
    });
    expect(byTestId(c, "backlog-saved-issue-query-row-iq1")!.textContent).toContain(en.defaultQuery);
    await act(async () => byTestId(c, "backlog-saved-query-star-q1")!.click());
    expect(calls(host, "git.queries.set_default")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { id: "q1", isDefault: true },
    });

    await act(async () => byTestId(c, "backlog-saved-issue-query-edit-iq1")!.click());
    await act(async () => setValue(byTestId(c, "backlog-save-query-name") as HTMLInputElement, "Bugs"));
    await act(async () => byTestId(c, "backlog-save-query-save")!.click());
    expect(save).toHaveBeenCalledWith({ ...ISSUE_QUERY, isDefault: true, name: "Bugs" });
    expect(byTestId(c, "backlog-saved-issue-query-row-iq1")!.textContent).toContain("Bugs");

    await act(async () => byTestId(c, "backlog-saved-issue-query-delete-iq1")!.click());
    await act(async () => byTestId(c, "backlog-saved-query-delete-dialog-confirm")!.click());
    expect(calls(host, "issues.queries.delete")[0]![1]).toEqual({ workspaceId: "ws-1", body: { id: "iq1" } });
    expect(byTestId(c, "backlog-saved-issue-queries-empty")).not.toBeNull();
  });
});

describe("Quick actions section (FR2.3, FR2.4)", () => {
  const labels = (c: HTMLElement) =>
    [...c.querySelectorAll<HTMLInputElement>('[data-testid^="backlog-quick-action-label-"]')].map(
      (i) => i.value,
    );

  it("shows the Issues tab with the defaults and saves edits, additions and deletions", async () => {
    const save = vi.fn(async (body: Record<string, unknown>) => ({
      ...DEFAULT_ACTIONS,
      issue: body.actions,
    }));
    const c = await render(scripted(connected, { "issues.quick_actions.save": save }));
    expect(byTestId(c, "backlog-quick-actions-tab-issue")!.getAttribute("aria-selected")).toBe("true");
    expect(labels(c)).toEqual(["Implement", "Investigate", "Reproduce"]);
    expect((byTestId(c, "backlog-quick-action-prompt-1") as HTMLTextAreaElement).value).toBe(
      "Investigate {{url}}",
    );

    await act(async () => setValue(byTestId(c, "backlog-quick-action-label-0") as HTMLInputElement, "Build"));
    await act(async () => byTestId(c, "backlog-quick-action-delete-2")!.click());
    await act(async () => byTestId(c, "backlog-quick-actions-add")!.click());
    expect(labels(c)).toEqual(["Build", "Investigate", en.newQuickAction]);
    await act(async () => choose(c, "backlog-quick-action-icon-2", "check"));
    await act(async () => byTestId(c, "backlog-quick-actions-save")!.click());
    expect(save).toHaveBeenCalledWith({
      kind: "issue",
      actions: [
        { ...DEFAULT_ACTIONS.issue[0], label: "Build" },
        DEFAULT_ACTIONS.issue[1],
        { id: "", label: en.newQuickAction, hint: "", icon: "check", promptTemplate: "" },
      ],
    });
    expect(byTestId(c, "backlog-quick-actions-status")!.textContent).toBe(en.quickActionsSaved);

    await act(async () => byTestId(c, "backlog-quick-actions-tab-pr")!.click());
    expect(labels(c)).toEqual(["Review"]);
  });

  it("resets a kind by saving an empty list and shows the defaults again", async () => {
    const save = vi.fn(async () => DEFAULT_ACTIONS);
    const c = await render(scripted(connected, { "issues.quick_actions.save": save }));
    await act(async () => byTestId(c, "backlog-quick-action-delete-0")!.click());
    await act(async () => byTestId(c, "backlog-quick-actions-reset")!.click());
    expect(save).toHaveBeenCalledWith({ kind: "issue", actions: [] });
    expect(labels(c)).toEqual(["Implement", "Investigate", "Reproduce"]);
  });

  it("refuses an empty or over-long label with a message and saves nothing", async () => {
    const save = vi.fn(async () => DEFAULT_ACTIONS);
    const c = await render(scripted(connected, { "issues.quick_actions.save": save }));
    await act(async () =>
      setValue(byTestId(c, "backlog-quick-action-label-1") as HTMLInputElement, "x".repeat(101)),
    );
    await act(async () => byTestId(c, "backlog-quick-actions-save")!.click());
    expect(byTestId(c, "backlog-quick-action-label-error-1")!.textContent).toBe(en.errorQuickActionLabel);
    await act(async () => setValue(byTestId(c, "backlog-quick-action-label-1") as HTMLInputElement, " "));
    await act(async () => byTestId(c, "backlog-quick-actions-save")!.click());
    expect(byTestId(c, "backlog-quick-action-label-error-1")).not.toBeNull();
    expect(save).not.toHaveBeenCalled();
  });

  it("shows a server refusal without saving and an error when loading fails", async () => {
    const c = await render(
      scripted(connected, {
        "issues.quick_actions.save": async () => {
          throw actionError(400, { code: "validation", field: "icon" });
        },
      }),
    );
    await act(async () => byTestId(c, "backlog-quick-actions-save")!.click());
    expect(byTestId(c, "backlog-quick-actions-notice")!.textContent).toBe(en.errorInput);
    unmount();
    const failing = await render(
      scripted(connected, {
        "issues.quick_actions.get": async () => {
          throw actionError(503, { code: "unreachable" });
        },
      }),
    );
    expect(byTestId(failing, "backlog-quick-actions-error")!.textContent).toContain(en.unreachable);
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
