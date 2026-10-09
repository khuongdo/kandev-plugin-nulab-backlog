import { act } from "react";
import axe from "axe-core";
import { afterEach, describe, expect, it } from "vitest";

import { createBacklogPage, pageState } from "./BacklogPage";
import { en, type Messages } from "../messages/en";
import {
  actionError,
  byTestId,
  connected,
  deferred,
  expectErrorAlert,
  expectOnlyCatalogueText,
  fakeHost,
  incomplete,
  mount,
  notConnected,
  text,
  unmount,
  type Invoke,
  type View,
} from "../testing/harness";

afterEach(unmount);

const SETTINGS_HREF = "/settings/workspaces/ws-1/integrations/nulab-backlog";
const off: View = { ...connected, enabled: false };
const noEnabledField: View = { ...connected };
delete noEnabledField.enabled;

/** U3: once connected, /backlog lists the issues; these tests answer with an empty list. */
const withIssues =
  (invoke: Invoke): Invoke =>
  async (key, input) => {
    if (key === "issues.filters") return { projects: [], statuses: [], assignees: [] };
    if (key === "issues.list") return { items: [], total: 0, page: 1, pageSize: 20 };
    return invoke(key, input);
  };

/** `workspace: null` means no active workspace. */
async function renderPage(invoke: Invoke, workspace: string | null = "ws-1", messages: Messages = en) {
  const host = fakeHost(withIssues(invoke), {
    context: {
      getActiveWorkspaceId: () => workspace ?? undefined,
      subscribeActiveWorkspace: () => () => undefined,
    },
  });
  const c = await mount(createBacklogPage(host, messages), {});
  return { c, host };
}

/** The states that show the alert instead of the lists (BR2.3). */
const alertStates: [string, View, string][] = [
  ["Off", off, en.pageOff],
  ["Off when the view has no enabled field (opt-in)", noEnabledField, en.pageOff],
  ["Not connected", notConnected, en.pageNotConnected],
  ["Incomplete", incomplete, en.incomplete],
];
const viewStates: [string, View, string][] = [...alertStates, ["Connected", connected, ""]];

describe("pageState (WF7)", () => {
  it.each([
    ["no workspace", { workspaceId: undefined, load: "ready" as const, view: connected }, "no_workspace"],
    ["loading", { workspaceId: "ws-1", load: "loading" as const }, "loading"],
    ["failed", { workspaceId: "ws-1", load: "failed" as const }, "failed"],
    ["off wins over the connection state", { workspaceId: "ws-1", load: "ready" as const, view: off }, "off"],
    [
      "no enabled field is off (opt-in)",
      { workspaceId: "ws-1", load: "ready" as const, view: noEnabledField },
      "off",
    ],
    ["not connected", { workspaceId: "ws-1", load: "ready" as const, view: notConnected }, "not_connected"],
    ["connected", { workspaceId: "ws-1", load: "ready" as const, view: connected }, "connected"],
    ["incomplete", { workspaceId: "ws-1", load: "ready" as const, view: incomplete }, "incomplete"],
  ])("%s", (_, input, kind) => {
    expect(pageState(input).kind).toBe(kind);
  });
});

describe("Backlog page (/backlog)", () => {
  it("asks for a workspace when none is active, without loading", async () => {
    const { c, host } = await renderPage(async () => connected, null);
    expect(text(c)).toContain(en.pageNoWorkspace);
    expect(host.api.invokeAction).not.toHaveBeenCalled();
  });

  it("shows a loading indicator while connection.get runs", async () => {
    const pending = deferred<unknown>();
    const { c } = await renderPage(() => pending.promise);
    expect(text(c)).toContain(en.pageLoading);
  });

  it("shows load failed with Retry, and Retry reloads", async () => {
    let fail = true;
    const { c, host } = await renderPage(async () => {
      if (fail) throw actionError(500, { code: "internal" });
      return connected;
    });
    expect(text(c)).toContain(en.pageLoadFailed);
    const alert = expectErrorAlert(byTestId(c, "backlog-page-error")); // intent 261009, FR2.2
    expect(alert.contains(byTestId(c, "backlog-page-retry"))).toBe(true);
    fail = false;
    await act(async () => byTestId(c, "backlog-page-retry")!.click());
    expect(byTestId(c, "backlog-scope-bar")).not.toBeNull();
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.get", { workspaceId: "ws-1" });
  });

  it.each(alertStates)(
    "shows %s in an alert with a link to that workspace's Backlog settings",
    async (_, view, message) => {
      const { c, host } = await renderPage(async () => view);
      expect(byTestId(c, "backlog-page-status")!.textContent).toBe(message);
      const link = byTestId(c, "backlog-page-settings-link")!;
      expect(link.getAttribute("data-variant")).toBe("link");
      expect(link.textContent).toBe(en.openSettings);
      await act(async () => link.click());
      expect(host.navigate).toHaveBeenCalledWith(SETTINGS_HREF);
    },
  );

  it("lists the issues once connected, and not otherwise (U3)", async () => {
    const { c, host } = await renderPage(async () => connected);
    expect(byTestId(c, "backlog-issues")).not.toBeNull();
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "issues.list",
      expect.objectContaining({ workspaceId: "ws-1" }),
    );
    unmount();
    const other = await renderPage(async () => notConnected);
    expect(byTestId(other.c, "backlog-issues")).toBeNull();
    expect(other.host.api.invokeAction).not.toHaveBeenCalledWith("issues.list", expect.anything());
  });

  it("puts a data-testid on every interactive element", async () => {
    for (const invoke of [async () => connected, async () => Promise.reject(actionError(500, {}))]) {
      const { c } = await renderPage(invoke);
      const interactive = c.querySelectorAll("button, input, select, textarea, a");
      expect(interactive.length).toBeGreaterThan(0);
      interactive.forEach((el) => expect(el.getAttribute("data-testid")).toBeTruthy());
      unmount();
    }
  });

  it("has no axe violations in each state", async () => {
    const invokes: [string, Invoke][] = [
      ...viewStates.map(([name, view]) => [name, async () => view] as [string, Invoke]),
      ["failed", async () => Promise.reject(actionError(500, {}))],
    ];
    for (const [name, invoke] of invokes) {
      const { c } = await renderPage(invoke);
      const result = await axe.run(c, { rules: { "color-contrast": { enabled: false } } });
      expect(result.violations.map((v) => `${name}: ${v.id}`)).toEqual([]);
      unmount();
    }
  });

  it("renders no literal text outside the message catalogue", async () => {
    const pseudo = Object.fromEntries(Object.entries(en).map(([k, v]) => [k, `⟦${v}⟧`])) as Messages;
    const invokes: Invoke[] = [
      ...viewStates.map(
        ([, view]) =>
          async () =>
            view,
      ),
      async () => Promise.reject(actionError(500, {})),
    ];
    for (const invoke of invokes) {
      const { c } = await renderPage(invoke, "ws-1", pseudo);
      expectOnlyCatalogueText(c, (ok, msg) => expect(ok, msg).toBe(true));
      unmount();
    }
    const { c } = await renderPage(async () => connected, null, pseudo);
    expectOnlyCatalogueText(c, (ok, msg) => expect(ok, msg).toBe(true));
  });
});

/** Opens the page on ws-1, switches to ws-2, and leaves both connection.get replies pending. */
async function overlappingLoads() {
  const replies: Record<string, ReturnType<typeof deferred<unknown>>> = {
    "ws-1": deferred<unknown>(),
    "ws-2": deferred<unknown>(),
  };
  let switchWorkspace: (id: string | undefined) => void = () => undefined;
  const host = fakeHost(async (_key, input) => replies[input!.workspaceId!].promise, {
    context: {
      getActiveWorkspaceId: () => "ws-1",
      subscribeActiveWorkspace: (listener: (id: string | undefined) => void) => {
        switchWorkspace = listener;
        return () => undefined;
      },
    },
  });
  const c = await mount(createBacklogPage(host, en), {});
  await act(async () => switchWorkspace("ws-2"));
  return { c, replies };
}

describe("BacklogPage workspace changes (R-03)", () => {
  it("ignores a late reply for the previous workspace when loads overlap out of order", async () => {
    const { c, replies } = await overlappingLoads();

    await act(async () => replies["ws-2"].resolve(connected));
    await act(async () => replies["ws-1"].resolve(off));
    expect(byTestId(c, "backlog-scope-bar")).not.toBeNull();
    expect(byTestId(c, "backlog-page-alert")).toBeNull();
  });

  it("ignores a late failure for the previous workspace", async () => {
    const { c, replies } = await overlappingLoads();

    await act(async () => replies["ws-2"].resolve(notConnected));
    await act(async () => replies["ws-1"].reject(actionError(500, { code: "internal" })));
    expect(text(byTestId(c, "backlog-page-status")!)).toBe(en.pageNotConnected);
  });
});
