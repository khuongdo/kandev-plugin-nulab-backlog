import * as React from "react";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  axeViolations,
  byTestId,
  expectOnlyCatalogueText,
  expectTestIds,
  fakeHost,
  mount,
  pseudoCatalogue,
  unmount,
} from "../testing/harness";
import { createIssueBadge } from "./issue-badge";
import type { LinkView } from "./issues-state";
import { createLinksStore } from "./links-store";

afterEach(unmount);

const HOST = "example-space.backlog.com";

function link(taskId: string, extra: Partial<LinkView> = {}): LinkView {
  return {
    taskId,
    taskKey: "T-1",
    issueKey: "PROJ-120",
    spaceHost: HOST,
    state: "active",
    status: "Resolved",
    statusUpdatedAt: "2026-10-06T00:00:00Z",
    stale: false,
    unavailable: false,
    url: `https://${HOST}/view/PROJ-120`,
    ...extra,
  };
}

function setup(links: LinkView[]) {
  let current = links;
  const host = fakeHost(async (key) => {
    if (key === "issues.links.list") return { links: current };
    throw new Error(`unexpected ${key}`);
  });
  return { host, set: (l: LinkView[]) => (current = l) };
}

const card = (taskId: string) => ({ slotProps: { taskId, workspaceId: "ws-1", workflowStepId: "s" } });

describe("Issue badge on the card (M6, task-card-tags)", () => {
  it("loads the links once per workspace for every card", async () => {
    const { host } = setup([link("task-1"), link("task-2", { issueKey: "PROJ-7" })]);
    const store = createLinksStore(host);
    const Badge = createIssueBadge(host, store);
    const two = () =>
      React.createElement(
        "div",
        null,
        React.createElement(Badge, card("task-1")),
        React.createElement(Badge, card("task-2")),
        React.createElement(Badge, card("task-3")),
      );
    const c = await mount(two, {});
    expect(
      vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "issues.links.list"),
    ).toHaveLength(1);
    expect(byTestId(c, "backlog-issue-badge-task-1")!.textContent).toBe("PROJ-120 · Resolved");
    expect(byTestId(c, "backlog-issue-badge-task-2")!.textContent).toBe("PROJ-7 · Resolved");
    expect(byTestId(c, "backlog-issue-badge-task-3")).toBeNull();
  });

  it("shows the update time on focus or tap, and may be out of date (AC4.1.4)", async () => {
    const { host } = setup([link("task-1", { stale: true })]);
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card("task-1"));
    const badge = byTestId(c, "backlog-issue-badge-task-1")!;
    expect(badge.getAttribute("aria-label")).toBe("Backlog issue PROJ-120 · Resolved");
    expect(byTestId(c, "backlog-issue-badge-detail-task-1")).toBeNull();
    await act(async () => badge.focus());
    expect(byTestId(c, "backlog-issue-badge-detail-task-1")!.textContent).toBe(
      "updated at rel:2026-10-06T00:00:00Z · may be out of date",
    );
    await act(async () => badge.blur());
    await act(async () => badge.click());
    expect(byTestId(c, "backlog-issue-badge-detail-task-1")).not.toBeNull();
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("explains a not-connected link (AC1.8.2, M12)", async () => {
    const { host } = setup([link("task-1", { state: "not_connected" })]);
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card("task-1"));
    const badge = byTestId(c, "backlog-issue-badge-task-1")!;
    expect(badge.textContent).toBe("PROJ-120 – not connected");
    await act(async () => badge.click());
    expect(byTestId(c, "backlog-issue-badge-detail-task-1")!.textContent).toBe(
      `Reconnect ${HOST} to restore this link`,
    );
  });

  it("says Issue unavailable (AC4.1.2)", async () => {
    const { host } = setup([link("task-1", { unavailable: true })]);
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card("task-1"));
    expect(byTestId(c, "backlog-issue-badge-task-1")!.textContent).toBe("PROJ-120 · Issue unavailable");
  });

  it("disappears after an unlink (AC3.3.5)", async () => {
    const { host, set } = setup([link("task-1")]);
    const store = createLinksStore(host);
    const c = await mount(createIssueBadge(host, store), card("task-1"));
    expect(byTestId(c, "backlog-issue-badge-task-1")).not.toBeNull();
    set([]);
    await act(async () => store.refresh("ws-1"));
    expect(byTestId(c, "backlog-issue-badge-task-1")).toBeNull();
  });

  it("renders nothing without slot props, and only catalogue text", async () => {
    const { host } = setup([link("task-1", { issueKey: "⟦PROJ-120⟧", status: "⟦Resolved⟧" })]);
    const div = document.createElement("div");
    document.body.appendChild(div);
    const root = createRoot(div);
    await act(async () =>
      root.render(React.createElement(createIssueBadge(host, createLinksStore(host)), {})),
    );
    expect(div.innerHTML).toBe("");
    act(() => root.unmount());
    div.remove();
    const c = await mount(
      createIssueBadge(host, createLinksStore(host), pseudoCatalogue(en)),
      card("task-1"),
    );
    expectOnlyCatalogueText(c, (ok, msg) => expect(ok, msg).toBe(true));
  });
});
