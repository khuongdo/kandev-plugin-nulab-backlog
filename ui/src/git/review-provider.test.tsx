import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import { actionError, byTestId, fakeHost, mount, unmount, type Invoke } from "../testing/harness";
import { createReviewProvider } from "./review-provider";

afterEach(unmount);

const SUMMARY = {
  providerId: "nulab-backlog",
  reviewKey: "example-space.backlog.com|11|42",
  title: "Add login page",
  url: "https://example-space.backlog.com/git/PROJ/web-app/pullRequests/42",
  connectionScope: "example-space.backlog.com",
  repositoryId: "11",
  changeRequestNumber: 42,
  state: "open",
  statusBadge: { label: "Open – Lan" },
  taskStatus: { number: 42, state: "open", pipelineState: "neutral", checks: [] },
  base: "main",
  branch: "feature/login",
  assignee: "Lan",
};

const ASSOCIATION = {
  providerId: "nulab-backlog",
  taskId: "task-17",
  reviewKey: SUMMARY.reviewKey,
  connectionScope: SUMMARY.connectionScope,
  repositoryId: "11",
  changeRequestNumber: 42,
};

function setup(invoke: Invoke) {
  const host = fakeHost(invoke);
  return { host, provider: createReviewProvider(host) };
}

const signal = () => new AbortController().signal;

describe("Review provider (M6, US5.4)", () => {
  it("names the provider and noun", () => {
    const { provider } = setup(async () => ({}));
    expect(provider).toMatchObject({
      id: "nulab-backlog",
      label: "Backlog",
      changeRequestNoun: en.reviewNoun,
    });
  });

  it("loads the associations from git.links.list", async () => {
    const { host, provider } = setup(async () => ({ associations: [ASSOCIATION] }));
    const listener = vi.fn();
    provider.subscribeAssociations!("ws-1", listener);
    expect(provider.getAssociationSnapshot!("ws-1")).toEqual([]);
    await provider.refreshAssociations!("ws-1", signal());
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "git.links.list",
      { workspaceId: "ws-1" },
      expect.anything(),
    );
    expect(provider.getAssociationSnapshot!("ws-1")).toEqual([ASSOCIATION]);
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("refreshes a task's summaries with git.prs.status (AC5.4.1)", async () => {
    const { host, provider } = setup(async () => ({ summaries: [SUMMARY] }));
    const listener = vi.fn();
    const stop = provider.subscribe("task-17", listener);
    const empty = provider.getSnapshot("task-17");
    expect(provider.getSnapshot("task-17")).toBe(empty);
    await provider.refresh("task-17", signal());
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "git.prs.status",
      { workspaceId: "ws-1", taskId: "task-17", body: {} },
      expect.anything(),
    );
    expect(provider.getSnapshot("task-17")).toEqual([SUMMARY]);
    expect(listener).toHaveBeenCalledTimes(1);
    stop();
    await provider.refresh("task-17", signal());
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("marks summaries Status unknown when the refresh fails (AC5.4.3)", async () => {
    let fail = false;
    const { provider } = setup(async () => {
      if (fail) throw actionError(503, { code: "unreachable" });
      return { summaries: [SUMMARY] };
    });
    await provider.refresh("task-17", signal());
    fail = true;
    await provider.refresh("task-17", signal());
    const [s] = provider.getSnapshot("task-17");
    expect(s!.statusBadge).toEqual({ label: en.statusUnknown });
    expect(s!.taskStatus?.error).toBe(en.unreachable);
  });

  it("unlinks with git.prs.unlink and drops the summary (AC5.2.3)", async () => {
    const { host, provider } = setup(async (key) =>
      key === "git.prs.status"
        ? { summaries: [SUMMARY] }
        : key === "git.links.list"
          ? { associations: [ASSOCIATION] }
          : {},
    );
    await provider.refresh("task-17", signal());
    await provider.refreshAssociations!("ws-1", signal());
    await provider.unlink!({ ...ASSOCIATION, workspaceId: "ws-1", signal: signal() });
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "git.prs.unlink",
      { workspaceId: "ws-1", taskId: "task-17", body: { reviewKey: SUMMARY.reviewKey } },
      expect.anything(),
    );
    expect(provider.getSnapshot("task-17")).toEqual([]);
    expect(provider.getAssociationSnapshot!("ws-1")).toEqual([]);
  });

  it("renders the review panel with Kandev's ChangeRequestDetail", async () => {
    const { provider } = setup(async () => ({ summaries: [SUMMARY] }));
    await provider.refresh("task-17", signal());
    const c = await mount(provider.ReviewPanel, {
      panelId: "p1",
      presentation: "desktop",
      workspaceId: "ws-1",
      taskId: "task-17",
      reviewKey: SUMMARY.reviewKey,
      connectionScope: SUMMARY.connectionScope,
      repositoryId: "11",
      changeRequestNumber: 42,
    });
    expect(byTestId(c, "fake-change-request-detail")!.textContent).toBe(
      "Add login page|open|feature/login|main",
    );
  });
});
