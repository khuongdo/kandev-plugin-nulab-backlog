import type { KandevPlugin, PluginHostApi } from "@kandev/plugin-sdk";

import { PLUGIN_ICON } from "./brand/backlog-logo";
import { createDashboardPage } from "./git/dashboard-page";
import { createPRLinkAction } from "./git/pr-link";
import { createRepositoryProvider } from "./git/repository-provider";
import { createReviewProvider } from "./git/review-provider";
import { createWatchesPage } from "./git/watches-page";
import { createIssueBadge } from "./issues/issue-badge";
import { createIssuePanel } from "./issues/issue-panel";
import { messagesFor, registerMessages } from "./issues/i18n";
import { createLinksStore } from "./issues/links-store";
import { createUnlinkMenuAction } from "./issues/task-menu";
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
    // U3 (US8.5): the catalogue first, then every screen reads Kandev's language.
    registerMessages(registry);
    const messages = messagesFor(host);
    const icon = PLUGIN_ICON(host);
    const links = createLinksStore(host);
    // These registrations never depend on the switch, so Backlog can always
    // be turned back on (BR5.4, BR7.6, BR7.8).
    registry.registerIntegrationSettings({
      id: PLUGIN_ID,
      label: messages.integrationLabel,
      description: messages.integrationDescription,
      icon,
      Component: createSettingsScreen(host, messages),
      action: createIntegrationSwitch(host, messages),
    });
    registry.registerNavItem({
      id: "backlog",
      label: messages.integrationLabel,
      path: "/backlog",
      section: "integrations",
      icon,
    });
    registry.registerRoute("/backlog", createBacklogPage(host, messages, links), {
      topbar: { title: messages.integrationLabel, icon },
    });
    // U4: Backlog Git repositories, PR links, the PR badge, PR watches and the dashboard.
    registry.registerRepositoryProvider(createRepositoryProvider(host, messages));
    registry.registerTaskAction(createPRLinkAction(host, messages));
    registry.registerReviewProvider(createReviewProvider(host, messages));
    for (const [id, path, label, page] of [
      ["backlog-watches", "/backlog/watches", messages.watchesTitle, createWatchesPage(host, messages)],
      [
        "backlog-dashboard",
        "/backlog/dashboard",
        messages.dashboardTitle,
        createDashboardPage(host, messages),
      ],
    ] as const) {
      registry.registerNavItem({ id, label, path, section: "integrations", icon });
      registry.registerRoute(path, page, { topbar: { title: label, icon } });
    }
    // U3: the issue badge on cards (M6), Unlink in the task menu, the issue panel (M8).
    registry.registerComponent("task-card-tags", createIssueBadge(host, links, messages));
    registry.registerTaskMenuAction(createUnlinkMenuAction(host, links, messages));
    registry.registerTaskPanel(createIssuePanel(host, messages, links));

    publishSwitches(host, host.context.getWorkspaceIds());
    unsubscribeWorkspaces = host.context.subscribeWorkspaces((ids) => publishSwitches(host, ids));
  },
  destroy() {
    unsubscribeWorkspaces?.();
    unsubscribeWorkspaces = undefined;
  },
};

window.registerKandevPlugin(PLUGIN_ID, plugin);
