import * as React from "react";
import { beforeAll, describe, expect, it, vi } from "vitest";
import type { KandevPlugin, PluginHostApi, PluginRegistry } from "@kandev/plugin-sdk";

import { actionError } from "./testing/harness";

let plugin: KandevPlugin;

beforeAll(async () => {
  const register = vi.fn<(id: string, plugin: KandevPlugin) => void>();
  (window as unknown as { registerKandevPlugin: typeof register }).registerKandevPlugin = register;
  await import("./index");
  expect(register).toHaveBeenCalledTimes(1);
  expect(register.mock.calls[0]![0]).toBe("nulab-backlog");
  plugin = register.mock.calls[0]![1];
});

const flush = () => new Promise((r) => setTimeout(r, 0));

function setup(views: Record<string, { enabled?: boolean } | Error | "hang">) {
  const registry = {
    registerIntegrationSettings: vi.fn(),
    registerNavItem: vi.fn(),
    registerRoute: vi.fn(),
    // U4
    registerRepositoryProvider: vi.fn(),
    registerTaskAction: vi.fn(),
    registerReviewProvider: vi.fn(),
    // U3
    registerTranslations: vi.fn(),
    registerComponent: vi.fn(),
    registerTaskMenuAction: vi.fn(),
    registerTaskPanel: vi.fn(),
  };
  let listener: ((ids: readonly string[]) => void) | undefined;
  const unsubscribe = vi.fn();
  const host = {
    React,
    jsx: React.createElement,
    ui: {},
    api: {
      invokeAction: vi.fn(async (_key: string, input: { workspaceId: string }) => {
        const v = views[input.workspaceId];
        if (v instanceof Error) throw v;
        if (v === "hang") return new Promise(() => undefined);
        return { connected: false, state: "not_connected", hasApiKey: false, ...v };
      }),
    },
    context: {
      getWorkspaceIds: () => Object.keys(views).slice(0, 2),
      subscribeWorkspaces: (l: (ids: readonly string[]) => void) => {
        listener = l;
        return unsubscribe;
      },
      getActiveWorkspaceId: () => "ws-1",
      subscribeActiveWorkspace: () => () => undefined,
    },
    setIntegrationEnabled: vi.fn(),
    i18n: { locale: "en", t: (_key: string, o?: { defaultValue?: string }) => o?.defaultValue ?? "" },
  };
  return {
    registry,
    host,
    unsubscribe,
    changeWorkspaces: (ids: string[]) => listener?.(ids),
    init: () => plugin.initialize(registry as unknown as PluginRegistry, host as unknown as PluginHostApi),
  };
}

describe("plugin entry", () => {
  it("registers the settings card with the logo icon and the switch action", async () => {
    const s = setup({});
    await s.init();
    expect(s.registry.registerIntegrationSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "nulab-backlog",
        label: "Backlog",
        icon: expect.any(Function),
        Component: expect.any(Function),
        action: expect.any(Function),
      }),
    );
  });

  it("registers the home Integrations entry and the /backlog route with the logo (BR7.6)", async () => {
    const s = setup({});
    await s.init();
    const icon = s.registry.registerIntegrationSettings.mock.calls[0]![0].icon;
    expect(s.registry.registerNavItem).toHaveBeenCalledWith({
      id: "backlog",
      label: "Backlog",
      path: "/backlog",
      section: "integrations",
      icon,
    });
    expect(s.registry.registerRoute).toHaveBeenCalledWith("/backlog", expect.any(Function), {
      topbar: { title: "Backlog", icon },
    });
  });

  it("adds no Integrations entry or route when Backlog is OFF in every workspace, but keeps the settings card (FR3.1, FR3.3, FR3.5)", async () => {
    const s = setup({ "ws-1": { enabled: false }, "ws-2": { enabled: false } });
    await s.init();
    expect(s.registry.registerNavItem).not.toHaveBeenCalled();
    expect(s.registry.registerRoute).not.toHaveBeenCalled();
    expect(s.registry.registerIntegrationSettings).toHaveBeenCalledTimes(1);
    expect(s.host.setIntegrationEnabled).toHaveBeenCalledWith("nulab-backlog", "ws-1", false);
  });

  it("adds the entry and the route when Backlog is ON in one workspace (FR3.1)", async () => {
    const s = setup({ "ws-1": { enabled: false }, "ws-2": { enabled: true } });
    await s.init();
    expect(s.registry.registerNavItem).toHaveBeenCalledTimes(1);
    expect(s.registry.registerRoute).toHaveBeenCalledTimes(1);
  });

  it("adds the entry when the ON/OFF state cannot be read (fail open, FR3.4)", async () => {
    const s = setup({ "ws-1": { enabled: false }, "ws-2": actionError(500, { code: "internal" }) });
    await s.init();
    expect(s.registry.registerNavItem).toHaveBeenCalledTimes(1);
    expect(s.registry.registerRoute).toHaveBeenCalledTimes(1);
  });

  it("stops waiting for the ON/OFF state well within the host's 10 s initialize limit and fails open (NFR3, FR3.4)", async () => {
    const { ENTRY_CHECK_TIMEOUT_MS } = await import("./index");
    expect(ENTRY_CHECK_TIMEOUT_MS).toBeLessThan(10_000);
    vi.useFakeTimers();
    try {
      const s = setup({ "ws-1": { enabled: false }, "ws-2": "hang" });
      const done = s.init();
      await vi.advanceTimersByTimeAsync(ENTRY_CHECK_TIMEOUT_MS);
      await done;
      expect(s.registry.registerNavItem).toHaveBeenCalledTimes(1);
      expect(s.registry.registerRoute).toHaveBeenCalledTimes(1);
    } finally {
      vi.useRealTimers();
    }
  });

  it("has exactly one Integrations entry and one route, /backlog (BR2.1, FR2.1, FR2.2)", async () => {
    const s = setup({});
    await s.init();
    expect(s.registry.registerNavItem.mock.calls.map(([item]) => item.path)).toEqual(["/backlog"]);
    expect(s.registry.registerRoute.mock.calls.map(([path]) => path)).toEqual(["/backlog"]);
  });

  it("registers the U4 provider, task action and review provider", async () => {
    const s = setup({});
    await s.init();
    expect(s.registry.registerRepositoryProvider).toHaveBeenCalledWith(
      expect.objectContaining({ id: "nulab-backlog", supportsDraft: false }),
    );
    expect(s.registry.registerTaskAction).toHaveBeenCalledWith(
      expect.objectContaining({ placement: "link" }),
    );
    expect(s.registry.registerReviewProvider).toHaveBeenCalledWith(
      expect.objectContaining({ id: "nulab-backlog" }),
    );
    for (const path of ["/backlog/watches", "/backlog/dashboard"]) {
      expect(s.registry.registerRoute).not.toHaveBeenCalledWith(path, expect.anything(), expect.anything());
    }
  });

  it("registers the U3 catalogue first, the card badge, the task menu and the task panel", async () => {
    const s = setup({});
    await s.init();
    expect(s.registry.registerTranslations).toHaveBeenCalledWith({
      en: expect.objectContaining({ linkToTask: "Link to task" }),
    });
    expect(s.registry.registerTranslations.mock.invocationCallOrder[0]).toBeLessThan(
      s.registry.registerIntegrationSettings.mock.invocationCallOrder[0]!,
    );
    expect(s.registry.registerComponent).toHaveBeenCalledWith("task-card-tags", expect.any(Function));
    expect(s.registry.registerTaskMenuAction).toHaveBeenCalledWith(
      expect.objectContaining({ id: "backlog-unlink-issue", group: "primary" }),
    );
    expect(s.registry.registerTaskPanel).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "backlog-issue",
        title: "Backlog issue",
        Component: expect.any(Function),
      }),
    );
  });

  it("registers one shared issue badge for Kanban cards, task rows and the task top bar (FR1.1-FR1.3, FR5.1, NFR2)", async () => {
    const s = setup({});
    await s.init();
    const slots = ["task-card-tags", "task-row-metadata", "chat-top-bar"];
    const calls = s.registry.registerComponent.mock.calls.filter(([slot]) => slots.includes(slot as string));
    expect(calls.map(([slot]) => slot)).toEqual(slots);
    // One component (so one links store): every slot shares one issues.links.list per workspace.
    expect(new Set(calls.map(([, c]) => c)).size).toBe(1);
  });

  it("publishes each workspace's switch at start, only after a successful load (BR7.5)", async () => {
    const s = setup({ "ws-1": { enabled: true }, "ws-2": actionError(500, { code: "internal" }) });
    await s.init();
    await flush();
    expect(s.host.api.invokeAction).toHaveBeenCalledWith("connection.get", { workspaceId: "ws-1" });
    expect(s.host.api.invokeAction).toHaveBeenCalledWith("connection.get", { workspaceId: "ws-2" });
    expect(s.host.setIntegrationEnabled).toHaveBeenCalledTimes(1);
    expect(s.host.setIntegrationEnabled).toHaveBeenCalledWith("nulab-backlog", "ws-1", true);
  });

  it("publishes off for a workspace whose view has no enabled field (opt-in)", async () => {
    const s = setup({ "ws-1": {} });
    await s.init();
    await flush();
    expect(s.host.setIntegrationEnabled).toHaveBeenCalledWith("nulab-backlog", "ws-1", false);
  });

  it("loads and publishes again when the workspace list changes", async () => {
    const s = setup({ "ws-1": { enabled: true }, "ws-2": { enabled: true }, "ws-3": { enabled: false } });
    await s.init();
    await flush();
    s.host.setIntegrationEnabled.mockClear();
    s.changeWorkspaces(["ws-1", "ws-3"]);
    await flush();
    expect(s.host.setIntegrationEnabled).toHaveBeenCalledWith("nulab-backlog", "ws-3", false);
    expect(s.host.setIntegrationEnabled).toHaveBeenCalledWith("nulab-backlog", "ws-1", true);
  });

  it("stops listening for workspace changes on destroy", async () => {
    const s = setup({});
    await s.init();
    await plugin.destroy?.();
    expect(s.unsubscribe).toHaveBeenCalledTimes(1);
  });
});
