import { describe, expect, it, vi } from "vitest";
import type { PluginHostApi, TaskContext } from "@kandev/plugin-sdk";

import { en } from "../messages/en";
import { actionError, fakeHost, type Invoke } from "../testing/harness";
import { createIssueLinkAction, parseIssueReference } from "./issue-link";
import type { LinkView } from "./issues-state";
import type { LinksStore } from "./links-store";

// Fake keys and the fake space acme.backlog.com only; no real credentials.
const TASK: TaskContext = {
  workspaceId: "ws-1",
  taskId: "task-17",
  repositories: [],
  pathname: "/t/task-17",
  presentation: "desktop",
};

type DialogOptions = Parameters<PluginHostApi["openTaskLinkDialog"]>[0];

function fakeStore(links?: readonly LinkView[]) {
  return {
    get: vi.fn(() => links),
    load: vi.fn(async () => undefined),
    refresh: vi.fn(async () => undefined),
    subscribe: vi.fn(() => () => undefined),
  } satisfies LinksStore;
}

async function open(invoke: Invoke, store = fakeStore([])) {
  const host = fakeHost(invoke);
  const action = createIssueLinkAction(host, store);
  await action.run(TASK);
  const options = vi.mocked(host.openTaskLinkDialog).mock.calls[0]![0] as DialogOptions;
  return { host, action, options, store };
}

const signal = () => new AbortController().signal;

describe("Link Backlog issue in the task's Link menu (FR1, FR3)", () => {
  it("is a single-task link action that opens Kandev's link dialog with catalogue copy (FR1.1, FR1.2)", async () => {
    const { action, options } = await open(async () => ({}));
    expect(action).toMatchObject({
      id: "nulab-backlog-link-issue",
      label: en.linkIssueLabel,
      placement: "link",
      singleTaskOnly: true,
    });
    expect(action.label).toBe("Link Backlog issue");
    expect(options).toMatchObject({
      title: en.linkIssueTitle,
      description: en.linkIssueDescription,
      inputLabel: en.linkIssueInput,
      placeholder: en.linkIssuePlaceholder,
      emptyError: en.linkIssueEmpty,
      failureMessage: en.linkIssueFailed,
      successMessage: en.linkIssueSuccess,
      inputTestId: "backlog-link-issue-input",
      errorTestId: "backlog-link-issue-error",
      submitTestId: "backlog-link-issue-submit",
    });
  });

  it("links the task to the key in a pasted link, then refreshes the links (FR1.3, FR1.4, FR2.2)", async () => {
    const { host, options, store } = await open(async () => ({ taskId: "task-17" }));
    const s = signal();
    await options.onSubmit("https://acme.backlog.com/view/PROJ-123", s);
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "issues.link",
      { workspaceId: "ws-1", taskId: "task-17", body: { issueKey: "PROJ-123" } },
      { signal: s },
    );
    expect(store.refresh).toHaveBeenCalledWith("ws-1");
  });

  it("still succeeds when the links refresh fails after the link (R-04)", async () => {
    const store = fakeStore([]);
    store.refresh.mockRejectedValue(new Error("offline"));
    const { options } = await open(async () => ({}), store);
    await expect(options.onSubmit("PROJ-1", signal())).resolves.toBeUndefined();
  });

  it("rejects text that is not an issue key or link without calling Backlog (FR2.3)", async () => {
    const { host, options } = await open(async () => ({}));
    await expect(options.onSubmit("https://example.com/view/PROJ-123", signal())).rejects.toThrow(
      en.notAnIssueReference,
    );
    expect(host.api.invokeAction).not.toHaveBeenCalled();
  });

  it.each([
    [{ code: "not_found" }, 404, "Issue PROJ-999 was not found."],
    [{ code: "conflict" }, 409, en.issueLinkConflict],
    [
      { code: "validation", field: "issueKey" },
      400,
      "PROJ is not one of the projects selected for this workspace.",
    ],
    [{ code: "reconnect_required" }, 401, en.reconnectRequired],
    [{ code: "rate_limited", retryAfterSeconds: 30 }, 429, "Backlog is limiting requests. Try again in 30 s"],
  ])("shows %j inline as a specific message (FR1.5)", async (error, status, message) => {
    const { options, store } = await open(async () => {
      throw actionError(status, { ...error, detail: "secret-token body" });
    });
    const failure = options.onSubmit("PROJ-999", signal());
    await expect(failure).rejects.toThrow(message);
    await expect(failure).rejects.not.toThrow(/secret/);
    expect(store.refresh).not.toHaveBeenCalled();
  });

  it("is hidden and starts the load before the links are loaded (FR3.1)", () => {
    const store = fakeStore(undefined);
    const action = createIssueLinkAction(
      fakeHost(async () => ({})),
      store,
    );
    expect(action.visible(TASK)).toBe(false);
    expect(store.load).toHaveBeenCalledWith("ws-1");
  });

  it("is hidden for a linked task and shown for an unlinked one (FR3.1, FR3.2)", () => {
    const linked = fakeStore([{ taskId: "task-17", issueKey: "PROJ-1", state: "active" } as LinkView]);
    expect(
      createIssueLinkAction(
        fakeHost(async () => ({})),
        linked,
      ).visible(TASK),
    ).toBe(false);
    const other = fakeStore([{ taskId: "task-18", issueKey: "PROJ-1", state: "active" } as LinkView]);
    expect(
      createIssueLinkAction(
        fakeHost(async () => ({})),
        other,
      ).visible(TASK),
    ).toBe(true);
  });
});

describe("parseIssueReference (FR2)", () => {
  it.each([
    ["PROJ-123", "PROJ-123"],
    ["proj-123", "PROJ-123"],
    ["  PROJ_2-7 \n", "PROJ_2-7"],
    ["https://acme.backlog.com/view/PROJ-123", "PROJ-123"],
    ["https://acme.backlog.jp/view/PROJ-123", "PROJ-123"],
    ["https://acme.backlogtool.com/view/PROJ-123", "PROJ-123"],
    ["https://acme.backlog.com/view/PROJ-123#comment-1", "PROJ-123"],
    ["https://acme.backlog.com/view/PROJ-123?x=1", "PROJ-123"],
  ])("reads %j as %s", (raw, key) => {
    expect(parseIssueReference(raw)).toBe(key);
  });

  it.each([
    "",
    "   ",
    "123",
    "PROJ-0",
    "PROJ 123",
    "hello",
    "http://acme.backlog.com/view/PROJ-123",
    "https://example.com/view/PROJ-123",
    "https://backlog.com.example.com/view/PROJ-123",
    "https://acme.backlog.com/projects/PROJ",
    "https://acme.backlog.com/view/PROJ-123/extra",
    "javascript:alert(1)",
  ])("rejects %j", (raw) => {
    expect(parseIssueReference(raw)).toBeUndefined();
  });
});
