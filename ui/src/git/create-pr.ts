import type { PluginHostApi, RepositoryProviderRegistration } from "@kandev/plugin-sdk";

import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "../settings/confirm-dialog";
import { readFailure } from "../settings/state";
import { PLUGIN_ID } from "../switch/enabled-events";
import { gitNotice, noticeError } from "./git-state";

type CreateFn = NonNullable<RepositoryProviderRegistration["createChangeRequest"]>;

interface LinkReply {
  spaceHost: string;
  projectKey: string;
  repoName: string;
  number: number;
}

/**
 * The provider's createChangeRequest (M11, US5.3). Kandev keeps its Create
 * PR dialog and pushes first; the plugin forwards only the task selectors,
 * so the backend takes the head branch from the verified context.
 */
export function createChangeRequest(host: PluginHostApi, messages: Messages = en): CreateFn {
  const ConfirmDialog = createConfirmDialog(host, messages);
  const h = host.jsx;

  /** Asks "Pull request #N is already open for this branch. Link it?" (AC5.3.4). */
  const confirmLink = (number: number) =>
    new Promise<boolean>((resolve) => {
      const text = format(messages.prExists, { number });
      const Content = () =>
        h(ConfirmDialog, {
          testId: "backlog-pr-exists-dialog",
          title: messages.prExistsTitle,
          body: text,
          confirmLabel: messages.linkIt,
          onConfirm: async () => resolve(true),
          onClose: () => {
            resolve(false);
            modal.close();
          },
        });
      const modal = host.openModal({ title: messages.prExistsTitle, content: Content, dismissible: false });
    });

  return async ({ workspaceId, taskId, sessionId, repositoryId, title, body, baseBranch, signal }) => {
    const selectors = { workspaceId, taskId, sessionId, repositoryId };
    try {
      const reply = await host.api.invokeAction<{ url: string; linked: boolean }>(
        "git.prs.create",
        { ...selectors, body: { title, body, baseBranch } },
        { signal },
      );
      return { url: reply.url, provider: PLUGIN_ID, linked: reply.linked };
    } catch (error) {
      const f = readFailure(error);
      if (f.code === "conflict" && f.pullRequestNumber) {
        if (!(await confirmLink(f.pullRequestNumber)))
          throw new Error(messages.prNotCreated, { cause: error });
        const link = await host.api.invokeAction<LinkReply>(
          "git.prs.link",
          { workspaceId, taskId, repositoryId, body: { reference: String(f.pullRequestNumber) } },
          { signal },
        );
        const url = `https://${link.spaceHost}/git/${link.projectKey}/${encodeURIComponent(link.repoName)}/pullRequests/${link.number}`;
        return { url, provider: PLUGIN_ID, linked: true };
      }
      if (f.code === "validation" && f.field === "title")
        throw new Error(messages.titleRequired, { cause: error });
      if (f.code === "validation" && f.field === "branch")
        throw new Error(messages.pushFirst, { cause: error });
      throw noticeError(gitNotice(error), messages, error);
    }
  };
}
