import { act } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  expectOnlyCatalogueText,
  expectTestIds,
  fakeHost,
  mount,
  pseudoCatalogue,
  setValue,
  text,
  unmount,
} from "../testing/harness";
import { createIssuePanel } from "./issue-panel";
import { createLinksStore } from "./links-store";

beforeEach(() => localStorage.clear());
afterEach(unmount);

const DETAIL = {
  issueKey: "PROJ-118",
  linkState: "active",
  statusUpdatedAt: "2026-10-06T00:00:00Z",
  issue: {
    key: "PROJ-118",
    summary: "Fix login timeout",
    status: "In Progress",
    assignee: "Lan",
    priority: "High",
    dueDate: "2026-10-10T00:00:00Z",
    url: "https://example-space.backlog.com/view/PROJ-118",
  },
  attachments: [
    { name: "spec.pdf", size: 1258291, tooLarge: false },
    { name: "dump.zip", size: 50331648, tooLarge: true },
  ],
};

const COMMENTS = Array.from({ length: 150 }, (_, i) => ({
  id: 150 - i,
  author: "Lan",
  content: `Comment ${150 - i}`,
  created: "2026-10-01T09:00:00Z",
}));

type Handlers = Record<string, (body: Record<string, unknown>) => Promise<unknown>>;

function setup(handlers: Handlers = {}) {
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    if (key === "issues.get") return DETAIL;
    if (key === "issues.comments") {
      const maxId = (body.maxId as number) || 150;
      const page = COMMENTS.filter((c) => c.id <= maxId).slice(0, 21);
      return { comments: page.slice(0, 20), nextMaxId: page.length > 20 ? page[20]!.id : undefined };
    }
    throw new Error(`unexpected ${key}`);
  });
}

const PANEL_PROPS = {
  panelId: "p",
  taskId: "task-17",
  sessionId: null,
  sessionKind: "task",
  presentation: "desktop",
};

const render = (host: ReturnType<typeof setup>, messages = en) =>
  mount(createIssuePanel(host, messages).Component, PANEL_PROPS);

describe("Backlog issue panel in the task (M8, US3.2, US3.5)", () => {
  it("registers as a task panel with a catalogue title", () => {
    const reg = createIssuePanel(setup());
    expect(reg.id).toBe("backlog-issue");
    expect(reg.title).toBe(en.issuePanelTitle);
    expect(reg.mobileEnabled).toBe(true);
  });

  it("is offered only for a linked task", async () => {
    const host = setup({
      "issues.links.list": async () => ({
        links: [{ taskId: "task-17", issueKey: "PROJ-118", state: "active" }],
      }),
    });
    const store = createLinksStore(host);
    const reg = createIssuePanel(host, en, store);
    const ctx = { taskId: "task-17", sessionId: null, sessionKind: "task", presentation: "desktop" } as const;
    expect(reg.visible!(ctx as never)).toBe(false);
    await store.load("ws-1");
    expect(reg.visible!(ctx as never)).toBe(true);
    expect(reg.visible!({ ...ctx, taskId: "task-99" } as never)).toBe(false);
  });

  it("shows the issue read-only with Open in Backlog (AC3.2.1)", async () => {
    const host = setup();
    const c = await render(host);
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenCalledWith("issues.get", {
      workspaceId: "ws-1",
      taskId: "task-17",
    });
    const panel = byTestId(c, "backlog-issue-panel")!;
    for (const v of ["PROJ-118", "Fix login timeout", "In Progress", "Lan", "High"])
      expect(panel.textContent).toContain(v);
    const due = byTestId(c, "backlog-issue-due")!.querySelector("time")!;
    expect(due.getAttribute("datetime")).toBe("2026-10-10T00:00:00Z");
    expect(due.textContent).toBe(
      new Date("2026-10-10T00:00:00Z").toLocaleDateString("en", { dateStyle: "long" }),
    );
    expect(byTestId(c, "backlog-issue-updated")!.textContent).toBe("updated at rel:2026-10-06T00:00:00Z");
    const open = byTestId(c, "backlog-issue-open")!;
    expect(open.getAttribute("href")).toBe(DETAIL.issue.url);
    expect(open.getAttribute("target")).toBe("_blank");
    expect(c.querySelectorAll("input, textarea, select")).toHaveLength(0);
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("refetches when reopened (AC3.2.2)", async () => {
    const host = setup();
    await render(host);
    unmount();
    await render(host);
    expect(vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "issues.get")).toHaveLength(2);
  });

  it("has its own unavailable, not-connected and unlinked states", async () => {
    let next: () => Promise<unknown> = async () => {
      throw actionError(404, { code: "not_found" });
    };
    const host = setup({ "issues.get": () => next() });
    let c = await render(host);
    expect(byTestId(c, "backlog-issue-state")!.textContent).toBe(en.issueUnavailableLong);
    unmount();
    next = async () => ({ issueKey: "PROJ-118", linkState: "not_connected", spaceHost: "old.backlog.jp" });
    c = await render(host);
    expect(byTestId(c, "backlog-issue-state")!.textContent).toBe(
      "Reconnect old.backlog.jp to restore this link",
    );
    unmount();
    next = async () => {
      throw actionError(503, { code: "unreachable" });
    };
    c = await render(host);
    expect(text(c)).toContain(en.unreachable);
    expect(byTestId(c, "backlog-issue-retry")).not.toBeNull();
  });

  it("shows 150 comments newest first with Load more (AC3.5.1)", async () => {
    const host = setup();
    const c = await render(host);
    const list = () => byTestId(c, "backlog-issue-comments-list")!.querySelectorAll("li");
    expect(list()).toHaveLength(20);
    expect(list()[0]!.textContent).toContain("Comment 150");
    for (let i = 0; i < 7; i++) await act(async () => byTestId(c, "backlog-issue-comments-more")!.click());
    expect(list()).toHaveLength(150);
    expect(list()[149]!.textContent).toContain("Comment 1");
    expect(byTestId(c, "backlog-issue-comments-more")).toBeNull();
    const calls = vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "issues.comments");
    expect(calls[1]![1]).toEqual({ workspaceId: "ws-1", taskId: "task-17", body: { maxId: 130 } });
  });

  it("lists attachments with size and the preview limit (AC3.5.2)", async () => {
    const c = await render(setup());
    expect(byTestId(c, "backlog-issue-attachment-0")!.textContent).toContain("spec.pdf");
    expect(byTestId(c, "backlog-issue-attachment-0")!.textContent).toContain("1.2 MB");
    const big = byTestId(c, "backlog-issue-attachment-1")!;
    expect(big.textContent).toContain("48.0 MB");
    expect(big.textContent).toContain(en.tooLargeToPreview);
    expect(byTestId(c, "backlog-issue-attachment-link-1")!.getAttribute("href")).toBe(DETAIL.issue.url);
  });

  it("keeps the other sections when comments fail, with Retry (AC3.5.3)", async () => {
    let fail = true;
    const host = setup({
      "issues.comments": async () => {
        if (fail) throw actionError(429, { code: "rate_limited", retryAfterSeconds: 5 });
        return { comments: [{ id: 1, author: "Lan", content: "Hi", created: "" }] };
      },
    });
    const c = await render(host);
    expect(byTestId(c, "backlog-issue-comments-error")!.textContent).toContain("Try again in 5 s");
    expect(byTestId(c, "backlog-issue-attachment-0")).not.toBeNull();
    fail = false;
    await act(async () => byTestId(c, "backlog-issue-comments-retry")!.click());
    expect(byTestId(c, "backlog-issue-comments-list")!.textContent).toContain("Hi");
  });

  it("toggles sections with aria-expanded and remembers the choice", async () => {
    const host = setup();
    let c = await render(host);
    const toggle = byTestId(c, "backlog-issue-comments-toggle")!;
    expect(toggle.tagName).toBe("BUTTON");
    expect(toggle.getAttribute("aria-expanded")).toBe("true");
    await act(async () => toggle.click());
    expect(toggle.getAttribute("aria-expanded")).toBe("false");
    expect(byTestId(c, "backlog-issue-comments-list")).toBeNull();
    unmount();
    c = await render(host);
    expect(byTestId(c, "backlog-issue-comments-toggle")!.getAttribute("aria-expanded")).toBe("false");
    expect(localStorage.getItem("nulab-backlog:panel:comments")).toBe("false");
  });

  it("uses only catalogue text", async () => {
    const host = setup({
      "issues.get": async () => ({
        issueKey: "⟦PROJ-1⟧",
        linkState: "active",
        issue: { key: "⟦PROJ-1⟧", summary: "⟦S⟧", status: "⟦Open⟧", url: "u" },
        attachments: [],
      }),
      "issues.comments": async () => ({ comments: [] }),
    });
    const c = await render(host, pseudoCatalogue(en));
    expectOnlyCatalogueText(c, (ok, msg) => expect(ok, msg).toBe(true));
  });
});

describe("Pull requests on the issue panel (FR5)", () => {
  const LINKS = {
    links: [
      {
        provider: "github",
        repo: "acme/web",
        number: 42,
        taskId: "task-17",
        title: "Add login",
        state: "merged",
        url: "https://github.com/acme/web/pull/42",
      },
      {
        provider: "gitlab",
        repo: "g/p",
        number: 7,
        issueKey: "PROJ-118",
        auto: true,
        title: "PROJ-118 timeout",
        state: "draft",
        url: "https://gitlab.com/g/p/-/merge_requests/7",
      },
    ],
  };
  const KANDEV = {
    pullRequests: [
      {
        taskId: "task-2",
        number: 5,
        url: "https://github.com/acme/web/pull/5",
        title: "Fix",
        state: "open",
        provider: "github",
        headBranch: "fix",
        baseBranch: "main",
      },
    ],
  };
  const prHost = (handlers: Handlers = {}) =>
    setup({
      "scm.links.list": async () => LINKS,
      "scm.task_prs.list": async () => KANDEV,
      ...handlers,
    });
  const prCalls = (host: ReturnType<typeof setup>, key: string) =>
    vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);

  it("shows linked and auto-linked pull requests and Kandev's own (FR5.1-FR5.4)", async () => {
    const host = prHost();
    const c = await render(host);
    expect(prCalls(host, "scm.links.list")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { taskId: "task-17", issueKey: "PROJ-118" },
    });
    expect(prCalls(host, "scm.task_prs.list")[0]![1]).toEqual({ workspaceId: "ws-1", taskId: "task-17" });
    const manual = byTestId(c, "backlog-issue-pr-github|acme/web|42")!;
    expect(byTestId(manual, "backlog-issue-pr-link-github|acme/web|42")!.getAttribute("href")).toBe(
      "https://github.com/acme/web/pull/42",
    );
    expect(text(manual)).toContain("GitHub");
    expect(text(manual)).toContain(en.stateMerged);
    expect(byTestId(manual, "backlog-issue-pr-auto-github|acme/web|42")).toBeNull();
    const auto = byTestId(c, "backlog-issue-pr-gitlab|g/p|7")!;
    expect(byTestId(auto, "backlog-issue-pr-auto-gitlab|g/p|7")!.textContent).toBe(en.autoLinked);
    expect(text(auto)).toContain(en.stateDraft);
    const kandev = byTestId(c, "backlog-issue-kandev-pr-task-2-5")!;
    expect(kandev.querySelector("a")!.getAttribute("href")).toBe("https://github.com/acme/web/pull/5");
    expect(text(kandev)).toContain(en.fromKandev);
  });

  it("removes an auto-link with its issue and a manual link without (FR5.3)", async () => {
    const host = prHost({ "scm.prs.unlink": async () => ({ ok: true }) });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-issue-pr-remove-gitlab|g/p|7")!.click());
    expect(prCalls(host, "scm.prs.unlink")[0]![1]).toEqual({
      workspaceId: "ws-1",
      taskId: "task-17",
      body: { key: "gitlab|g/p|7", issueKey: "PROJ-118" },
    });
    expect(byTestId(c, "backlog-issue-pr-gitlab|g/p|7")).toBeNull();
    await act(async () => byTestId(c, "backlog-issue-pr-remove-github|acme/web|42")!.click());
    expect(prCalls(host, "scm.prs.unlink")[1]![1]).toMatchObject({ body: { key: "github|acme/web|42" } });
  });

  it("links a pasted pull request URL to the task (FR5.1)", async () => {
    const linked = {
      provider: "bitbucket",
      repo: "ws/r",
      number: 3,
      taskId: "task-17",
      title: "New",
      state: "open",
      url: "https://bitbucket.org/ws/r/pull-requests/3",
    };
    let fail = true;
    const host = prHost({
      "scm.prs.link": async () => {
        if (fail) throw actionError(400, { code: "validation", field: "url" });
        return linked;
      },
    });
    const c = await render(host);
    expect(byTestId(c, "backlog-issue-pr-url"), "the panel opens read-only").toBeNull();
    await act(async () => byTestId(c, "backlog-issue-pr-add")!.click());
    await act(async () => setValue(byTestId(c, "backlog-issue-pr-url") as HTMLInputElement, "https://x"));
    await act(async () => byTestId(c, "backlog-issue-pr-link")!.click());
    expect(byTestId(c, "backlog-issue-pr-notice")!.textContent).toBe(en.errorPrUrl);
    fail = false;
    await act(async () => byTestId(c, "backlog-issue-pr-link")!.click());
    expect(prCalls(host, "scm.prs.link").at(-1)![1]).toEqual({
      workspaceId: "ws-1",
      taskId: "task-17",
      body: { url: "https://x" },
    });
    expect(byTestId(c, "backlog-issue-pr-bitbucket|ws/r|3")).not.toBeNull();
    expect((byTestId(c, "backlog-issue-pr-url") as HTMLInputElement).value).toBe("");
  });

  it("says when there is none, and a failed read never hides the issue", async () => {
    const host = prHost({
      "scm.links.list": async () => ({ links: [] }),
      "scm.task_prs.list": async () => {
        throw actionError(409, { code: "integration_disabled" });
      },
    });
    const c = await render(host);
    expect(byTestId(c, "backlog-issue-prs-empty")!.textContent).toBe(en.noPullRequests);
    expect(byTestId(c, "backlog-issue-prs-error")).not.toBeNull();
    expect(byTestId(c, "backlog-issue-open")).not.toBeNull();
  });

  it("keeps catalogue text, test ids and axe", async () => {
    const c = await render(prHost(), pseudoCatalogue(en));
    await act(async () => byTestId(c, "backlog-issue-pr-add")!.click());
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
    expect(await axeViolations(c)).toEqual([]);
  });
});
