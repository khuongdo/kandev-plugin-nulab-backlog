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

function setup(views: Record<string, { enabled: boolean } | Error>) {
  const registry = {
    registerIntegrationSettings: vi.fn(),
    registerNavItem: vi.fn(),
    registerRoute: vi.fn(),
    // U4
    registerRepositoryProvider: vi.fn(),
    registerTaskAction: vi.fn(),
    registerReviewProvider: vi.fn(),
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

  it("registers the entry and the route even when Backlog is off everywhere", async () => {
    const s = setup({ "ws-1": { enabled: false }, "ws-2": { enabled: false } });
    await s.init();
    await flush();
    expect(s.registry.registerNavItem).toHaveBeenCalledTimes(3); // U4 adds watches and the dashboard
    expect(s.registry.registerRoute).toHaveBeenCalledTimes(3);
  });

  it("registers the U4 provider, task action, review provider and pages", async () => {
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
      expect(s.registry.registerRoute).toHaveBeenCalledWith(path, expect.any(Function), expect.anything());
      expect(s.registry.registerNavItem).toHaveBeenCalledWith(
        expect.objectContaining({ path, section: "integrations" }),
      );
    }
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
