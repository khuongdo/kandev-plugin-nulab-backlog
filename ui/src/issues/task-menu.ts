import type { PluginHostApi, TaskMenuActionRegistration } from "@kandev/plugin-sdk";

import { en, type Messages } from "../messages/en";
import { noticeText } from "../git/git-state";
import { issueNotice } from "./issues-state";
import type { LinksStore } from "./links-store";

/**
 * "Unlink Backlog issue" in the task menu (US3.3). It shows only for a
 * linked task, decided synchronously from the shared links store; before the
 * store has the workspace it starts the load and stays hidden.
 */
export function createUnlinkMenuAction(
  host: PluginHostApi,
  store: LinksStore,
  messages: Messages = en,
): TaskMenuActionRegistration {
  return {
    id: "backlog-unlink-issue",
    label: messages.unlinkIssue,
    group: "primary",
    visible({ workspaceId, taskId }) {
      const links = store.get(workspaceId);
      if (!links) {
        void store.load(workspaceId);
        return false;
      }
      return links.some((l) => l.taskId === taskId);
    },
    async run({ workspaceId, taskId }) {
      try {
        await host.api.invokeAction("issues.unlink", { workspaceId, taskId });
      } catch (error) {
        host.toast.error(noticeText(issueNotice(error), messages));
        return;
      }
      host.toast.success(messages.unlinked);
      await store.refresh(workspaceId);
    },
  };
}
