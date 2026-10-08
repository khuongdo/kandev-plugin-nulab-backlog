import type { KandevPlugin, PluginHostApi } from "@kandev/plugin-sdk";

import { PLUGIN_ICON } from "./brand/backlog-logo";
import { createPRLinkAction } from "./git/pr-link";
import { createRepositoryProvider } from "./git/repository-provider";
import { createReviewProvider } from "./git/review-provider";
import { createIssueBadge } from "./issues/issue-badge";
import { createIssueLinkAction } from "./issues/issue-link";
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
 * NFR3: how long initialize waits for the ON/OFF state before it shows the
 * Integrations entry anyway (FR3.4). Kandev gives up on initialize after 10 s.
 */
export const ENTRY_CHECK_TIMEOUT_MS = 3_000;

/**
 * Loads each workspace's switch and publishes it to Kandev, only after a
 * successful load (BR7.5). A failed load leaves that workspace unpublished.
 * Resolves with each state: true or false, undefined when it could not be read.
 */
function publishSwitches(
  host: PluginHostApi,
  workspaceIds: readonly string[],
): Promise<(boolean | undefined)[]> {
  return Promise.all(
    workspaceIds.map((workspaceId) =>
      host.api
        .invokeAction<ConnectionView>("connection.get", { workspaceId })
        .then((view) => {
          const enabled = view.enabled === true;
          publishEnabled(host, workspaceId, enabled);
          return enabled;
        })
        .catch(() => {
          console.warn(`nulab-backlog: could not load the switch for workspace ${workspaceId}`);
          return undefined;
        }),
    ),
  );
}

/**
 * FR3.1, FR3.4: show the Integrations entry unless every workspace reads OFF.
 * No workspace yet, a failed read or a timeout counts as unknown: show it.
 */
async function showEntry(states: Promise<(boolean | undefined)[]>): Promise<boolean> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  const timeout = new Promise<undefined>((resolve) => {
    timer = setTimeout(() => resolve(undefined), ENTRY_CHECK_TIMEOUT_MS);
  });
  const read = await Promise.race([states, timeout]);
  clearTimeout(timer);
  return !read || read.length === 0 || read.some((on) => on !== false);
}

let unsubscribeWorkspaces: (() => void) | undefined;

const plugin: KandevPlugin = {
  async initialize(registry, host) {
    // U3 (US8.5): the catalogue first, then every screen reads Kandev's language.
    registerMessages(registry);
    const messages = messagesFor(host);
    const icon = PLUGIN_ICON(host);
    const links = createLinksStore(host);
    // The settings card never depends on the switch, so Backlog can always be
    // turned back on (FR3.3).
    registry.registerIntegrationSettings({
      id: PLUGIN_ID,
      label: messages.integrationLabel,
      description: messages.integrationDescription,
      icon,
      Component: createSettingsScreen(host, messages),
      action: createIntegrationSwitch(host, messages),
    });
    // U4: Backlog Git repositories, PR links and the PR badge. PR watches and
    // saved queries live in the settings, the PR list on /backlog (FR1, FR2).
    registry.registerRepositoryProvider(createRepositoryProvider(host, messages));
    registry.registerTaskAction(createPRLinkAction(host, messages));
    // FR1, FR3: Link Backlog issue in the same Link menu, hidden for a linked task.
    registry.registerTaskAction(createIssueLinkAction(host, links, messages));
    registry.registerReviewProvider(createReviewProvider(host, messages));
    // U3: the issue badge on cards (M6), task rows of Home > Tasks and the
    // sidebar (FR1) and the task top bar (FR5); one component, one links store
    // (NFR2). Then Unlink in the task menu and the issue panel (M8).
    const badge = createIssueBadge(host, links, messages);
    for (const slot of ["task-card-tags", "task-row-metadata", "chat-top-bar"]) {
      registry.registerComponent(slot, badge);
    }
    registry.registerTaskMenuAction(createUnlinkMenuAction(host, links, messages));
    registry.registerTaskPanel(createIssuePanel(host, messages, links));

    const states = publishSwitches(host, host.context.getWorkspaceIds());
    unsubscribeWorkspaces = host.context.subscribeWorkspaces((ids) => void publishSwitches(host, ids));

    // FR3 (supersedes BR5.4, BR7.6, BR7.8): Kandev cannot hide a menu entry
    // after load, so the one Integrations entry and its /backlog route (the
    // Issues and Pull requests lists, BR2.1) are added only when Backlog is ON
    // somewhere at load; a toggle shows in the menu after a reload (FR3.2).
    if (await showEntry(states)) {
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
    }
  },
  destroy() {
    unsubscribeWorkspaces?.();
    unsubscribeWorkspaces = undefined;
  },
};

window.registerKandevPlugin(PLUGIN_ID, plugin);
