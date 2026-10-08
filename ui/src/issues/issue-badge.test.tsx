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

  it("opens the Backlog issue in a new tab without opening or dragging the card (FR2, BR2.1, BR2.2, R-01, R-08)", async () => {
    const { host } = setup([link("task-1", { stale: true })]);
    const Badge = createIssueBadge(host, createLinksStore(host));
    const onCard = vi.fn();
    const onDrag = vi.fn();
    const inCard = () =>
      React.createElement(
        "div",
        { onClick: onCard, onPointerDown: onDrag },
        React.createElement(Badge, card("task-1")),
      );
    const c = await mount(inCard, {});
    const badge = byTestId(c, "backlog-issue-badge-task-1")!;
    expect(badge.tagName).toBe("A");
    expect(badge.getAttribute("href")).toBe(`https://${HOST}/view/PROJ-120`);
    expect(badge.getAttribute("target")).toBe("_blank");
    expect(badge.getAttribute("rel")).toBe("noopener noreferrer");
    expect(badge.getAttribute("data-host")).toBe("Button"); // host Button asChild: the outline badge style
    expect(badge.getAttribute("data-variant")).toBe("outline");
    expect(badge.textContent).toBe("PROJ-120 · Resolved");
    expect(badge.getAttribute("aria-label")).toBe("Backlog issue PROJ-120 · Resolved");
    // R-01: the detail is the tooltip and the link's description, so keyboard and screen readers get it.
    const detail = "updated at rel:2026-10-06T00:00:00Z · may be out of date";
    expect(badge.hasAttribute("title")).toBe(false); // FR1.4: the hover card replaces the native title
    const described = document.getElementById(badge.getAttribute("aria-describedby")!)!;
    expect(described.textContent).toBe(detail);
    expect(described.classList).toContain("sr-only");
    await act(async () => {
      badge.dispatchEvent(new MouseEvent("pointerdown", { bubbles: true }));
      badge.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    });
    expect(onCard).not.toHaveBeenCalled();
    expect(onDrag).not.toHaveBeenCalled();
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("shows a hover card with the key, summary and status on hover and on keyboard focus (FR1.4, FR1.5, NFR4)", async () => {
    const { host } = setup([
      link("task-1", { issueKey: "PROJ-12", summary: "Fix login", status: "In Progress" }),
    ]);
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card("task-1"));
    const badge = byTestId(c, "backlog-issue-badge-task-1")!;
    expect(badge.getAttribute("href")).toBe(`https://${HOST}/view/PROJ-120`);
    expect(badge.getAttribute("aria-label")).toBe("Backlog issue PROJ-12 · In Progress, Fix login");
    const hover = () => byTestId(c, "backlog-issue-badge-hover-task-1");
    expect(hover()).toBeNull();
    await act(async () => badge.dispatchEvent(new MouseEvent("mouseover", { bubbles: true })));
    expect(hover()!.getAttribute("data-host")).toBe("TooltipContent");
    expect([...hover()!.querySelectorAll("[data-hover-line]")].map((e) => e.textContent)).toEqual([
      "PROJ-12",
      "Fix login",
      "In Progress",
    ]);
    await act(async () => badge.dispatchEvent(new MouseEvent("mouseout", { bubbles: true })));
    expect(hover()).toBeNull();
    await act(async () => badge.focus());
    expect(hover()).not.toBeNull();
    expect(await axeViolations(c)).toEqual([]);
  });

  it("shows the key and status without an empty summary line for an old link (FR1.4)", async () => {
    const { host } = setup([link("task-1", { summary: undefined })]);
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card("task-1"));
    const badge = byTestId(c, "backlog-issue-badge-task-1")!;
    expect(badge.getAttribute("aria-label")).toBe("Backlog issue PROJ-120 · Resolved");
    await act(async () => badge.focus());
    const hover = byTestId(c, "backlog-issue-badge-hover-task-1")!;
    expect([...hover.querySelectorAll("[data-hover-line]")].map((e) => e.textContent)).toEqual([
      "PROJ-120",
      "Resolved",
    ]);
  });

  it("is no link when the issue cannot be opened, and shows the detail on focus or tap (BR2.3, R-01, R-07)", async () => {
    for (const extra of [
      { unavailable: true },
      { url: undefined },
      { url: `http://${HOST}/view/PROJ-120` },
    ]) {
      const { host } = setup([link("task-1", extra)]);
      const onCard = vi.fn();
      const Badge = createIssueBadge(host, createLinksStore(host));
      const inCard = () =>
        React.createElement("div", { onClick: onCard }, React.createElement(Badge, card("task-1")));
      const c = await mount(inCard, {});
      const badge = byTestId(c, "backlog-issue-badge-task-1")!;
      expect(badge.tagName).toBe("BUTTON");
      expect(badge.hasAttribute("href")).toBe(false);
      expect(byTestId(c, "backlog-issue-badge-detail-task-1")).toBeNull();
      await act(async () => badge.focus());
      expect(byTestId(c, "backlog-issue-badge-detail-task-1")!.textContent).toBe(
        "updated at rel:2026-10-06T00:00:00Z",
      );
      await act(async () => badge.click());
      expect(onCard).not.toHaveBeenCalled();
      expect(await axeViolations(c)).toEqual([]);
      unmount();
    }
  });

  it("explains a not-connected link (AC1.8.2, M12)", async () => {
    const { host } = setup([link("task-1", { state: "not_connected" })]);
    const c = await mount(createIssueBadge(host, createLinksStore(host)), card("task-1"));
    const badge = byTestId(c, "backlog-issue-badge-task-1")!;
    expect(badge.textContent).toBe("PROJ-120 – not connected");
    expect(badge.hasAttribute("href")).toBe(false);
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

describe("Issue badge on task rows and the task top bar (FR1.1, FR1.2, FR5.1, FR5.3, FR5.5)", () => {
  const surfaces = [
    { name: "Home > Tasks row", props: { surface: "task-list" } },
    { name: "sidebar row", props: { surface: "sidebar" } },
    { name: "desktop top bar", props: { presentation: "desktop", activeSessionId: null, sessionIds: [] } },
    { name: "phone top bar", props: { presentation: "mobile", activeSessionId: null, sessionIds: [] } },
  ];
  for (const { name, props } of surfaces) {
    it(`shows the badge for a linked task and nothing for an unlinked one on the ${name}`, async () => {
      const { host } = setup([link("task-1")]);
      const store = createLinksStore(host);
      const Badge = createIssueBadge(host, store);
      const slot = (taskId: string) => ({
        slotProps: { taskId, workspaceId: "ws-1", workflowStepId: "s", ...props },
      });
      const both = () =>
        React.createElement(
          "div",
          null,
          React.createElement(Badge, slot("task-1")),
          React.createElement(Badge, slot("task-2")),
        );
      const c = await mount(both, {});
      const badge = byTestId(c, "backlog-issue-badge-task-1")!;
      expect(badge.textContent).toBe("PROJ-120 · Resolved");
      expect(badge.getAttribute("href")).toBe(`https://${HOST}/view/PROJ-120`);
      expect(byTestId(c, "backlog-issue-badge-task-2")).toBeNull();
      // FR5.5: the host's phone menu rules, a 44px touch target.
      expect(badge.classList.contains("min-h-11")).toBe(props.presentation === "mobile");
    });
  }
});
