import { act } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { byTestId, fakeHost, mount, unmount } from "../testing/harness";
import { createIssueBadge } from "./issue-badge";
import type { LinkView } from "./issues-state";
import { createLinksStore, LINKS_REFRESH_MS } from "./links-store";
import { createUnlinkMenuAction } from "./task-menu";

const HOST = "example-space.backlog.com";

function link(status: string): LinkView {
  return {
    taskId: "task-1",
    taskKey: "T-1",
    issueKey: "PROJ-120",
    spaceHost: HOST,
    state: "active",
    status,
    statusUpdatedAt: "2026-10-06T00:00:00Z",
    stale: false,
    unavailable: false,
    url: `https://${HOST}/view/PROJ-120`,
  };
}

/** A host whose issues.links.list answers with the current status, or fails. */
function setup() {
  const state = { status: "In Progress", fail: false };
  const host = fakeHost(async (key) => {
    if (key !== "issues.links.list") throw new Error(`unexpected ${key}`);
    if (state.fail) throw new Error("integration disabled");
    return { links: [link(state.status)] };
  });
  const calls = () =>
    vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "issues.links.list").length;
  return { host, state, calls };
}

const card = { slotProps: { taskId: "task-1", workspaceId: "ws-1", workflowStepId: "s" } };
const advance = (ms: number) => act(async () => void (await vi.advanceTimersByTimeAsync(ms)));

beforeEach(() => {
  vi.useFakeTimers();
  vi.spyOn(console, "warn").mockImplementation(() => {});
});

afterEach(() => {
  unmount();
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe("Links store refresh (review 1, R-03)", () => {
  it("shows a status the server sync changed after one interval of 60 s", async () => {
    const { host, state, calls } = setup();
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card);
    expect(byTestId(c, "backlog-issue-badge-task-1")!.textContent).toBe("PROJ-120 · In Progress");
    state.status = "Closed";
    await advance(LINKS_REFRESH_MS - 1);
    expect(byTestId(c, "backlog-issue-badge-task-1")!.textContent).toBe("PROJ-120 · In Progress");
    await advance(1);
    expect(byTestId(c, "backlog-issue-badge-task-1")!.textContent).toBe("PROJ-120 · Closed");
    expect(calls()).toBe(2);
    expect(LINKS_REFRESH_MS).toBe(60_000);
  });

  it("makes no request once nothing is mounted", async () => {
    const { host, calls } = setup();
    await mount(createIssueBadge(host, createLinksStore(host)), card);
    expect(calls()).toBe(1);
    unmount();
    await advance(10 * LINKS_REFRESH_MS);
    window.dispatchEvent(new Event("focus"));
    await advance(0);
    expect(calls()).toBe(1);
  });

  it("refreshes on focus and on becoming visible, sharing a request in flight", async () => {
    const { host, state, calls } = setup();
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card);
    state.status = "Closed";
    await act(async () => {
      window.dispatchEvent(new Event("focus"));
      document.dispatchEvent(new Event("visibilitychange"));
    });
    expect(calls()).toBe(2);
    expect(byTestId(c, "backlog-issue-badge-task-1")!.textContent).toBe("PROJ-120 · Closed");
    await act(async () => void document.dispatchEvent(new Event("visibilitychange")));
    expect(calls()).toBe(3);
  });

  it("does not retry a failed load on every render, and backs off", async () => {
    const { host, state, calls } = setup();
    state.fail = true;
    const store = createLinksStore(host);
    const menu = createUnlinkMenuAction(host, store);
    const ctx = {
      workspaceId: "ws-1",
      taskId: "task-1",
      taskTitle: "T",
      workflowStepId: null,
      presentation: "desktop" as const,
    };
    await mount(createIssueBadge(host, store), card);
    for (let i = 0; i < 10; i++) expect(menu.visible!(ctx)).toBe(false);
    await act(async () => void window.dispatchEvent(new Event("focus")));
    expect(calls()).toBe(1);
    await advance(LINKS_REFRESH_MS);
    expect(calls()).toBe(2);
    await advance(LINKS_REFRESH_MS);
    expect(calls()).toBe(2); // the second failure waits 2 minutes
    state.fail = false;
    await advance(LINKS_REFRESH_MS);
    expect(calls()).toBe(3);
    expect(menu.visible!(ctx)).toBe(true);
  });
});
