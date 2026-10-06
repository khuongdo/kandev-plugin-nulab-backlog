import type { KandevPlugin, PluginHostApi } from "@kandev/plugin-sdk";

import { PLUGIN_ICON } from "./brand/backlog-logo";
import { en } from "./messages/en";
import { createBacklogPage } from "./page/BacklogPage";
import { createSettingsScreen } from "./settings/SettingsScreen";
import type { ConnectionView } from "./settings/state";
import { PLUGIN_ID, publishEnabled } from "./switch/enabled-events";
import { createIntegrationSwitch } from "./switch/integration-switch";

declare global {
  interface Window {
    registerKandevPlugin(id: string, plugin: KandevPlugin): void;
  }
}

/**
 * Loads each workspace's switch and publishes it to Kandev, only after a
 * successful load (BR7.5). A failed load leaves that workspace unpublished.
 */
function publishSwitches(host: PluginHostApi, workspaceIds: readonly string[]): void {
  for (const workspaceId of workspaceIds) {
    host.api
      .invokeAction<ConnectionView>("connection.get", { workspaceId })
      .then((view) => publishEnabled(host, workspaceId, view.enabled !== false))
      .catch(() => console.warn(`nulab-backlog: could not load the switch for workspace ${workspaceId}`));
  }
}

let unsubscribeWorkspaces: (() => void) | undefined;

const plugin: KandevPlugin = {
  initialize(registry, host) {
    const icon = PLUGIN_ICON(host);
    // These registrations never depend on the switch, so Backlog can always
    // be turned back on (BR5.4, BR7.6, BR7.8).
    registry.registerIntegrationSettings({
      id: PLUGIN_ID,
      label: en.integrationLabel,
      description: en.integrationDescription,
      icon,
      Component: createSettingsScreen(host),
      action: createIntegrationSwitch(host),
    });
    registry.registerNavItem({
      id: "backlog",
      label: en.integrationLabel,
      path: "/backlog",
      section: "integrations",
      icon,
    });
    registry.registerRoute("/backlog", createBacklogPage(host), {
      topbar: { title: en.integrationLabel, icon },
    });

    publishSwitches(host, host.context.getWorkspaceIds());
    unsubscribeWorkspaces = host.context.subscribeWorkspaces((ids) => publishSwitches(host, ids));
  },
  destroy() {
    unsubscribeWorkspaces?.();
    unsubscribeWorkspaces = undefined;
  },
};

window.registerKandevPlugin(PLUGIN_ID, plugin);
