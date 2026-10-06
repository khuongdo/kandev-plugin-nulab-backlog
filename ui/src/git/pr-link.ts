import type { PluginHostApi, TaskContext } from "@kandev/plugin-sdk";

import { en, format, type Messages } from "../messages/en";
import { readFailure } from "../settings/state";
import { PLUGIN_ID } from "../switch/enabled-events";
import { gitNotice, noticeError } from "./git-state";

/**
 * "Link Backlog pull request" in the task's Link menu (M7, US5.2). Kandev
 * owns the dialog, its validation display, submit state and toast; the
 * plugin supplies the copy and the submit.
 */
export function createPRLinkAction(host: PluginHostApi, messages: Messages = en) {
  return {
    id: `${PLUGIN_ID}-link-pr`,
    label: messages.linkPRLabel,
    placement: "link" as const,
    singleTaskOnly: true,
    run: async (task: TaskContext) => {
      const backlogRepos = task.repositories.filter((r) => r.provider === PLUGIN_ID);
      // The short forms (<n>, <repo>#<n>) need exactly one Backlog repository.
      const repositoryId = backlogRepos.length === 1 ? backlogRepos[0]!.id : undefined;
      const space = backlogRepos[0]?.provider_scope || messages.theSpace;
      host.openTaskLinkDialog({
        title: messages.linkPRTitle,
        description: messages.linkPRDescription,
        inputLabel: messages.linkPRInput,
        placeholder: messages.linkPRPlaceholder,
        emptyError: messages.linkPREmpty,
        failureMessage: messages.linkPRFailed,
        successMessage: messages.linkPRSuccess,
        inputTestId: "backlog-link-pr-input",
        errorTestId: "backlog-link-pr-error",
        submitTestId: "backlog-link-pr-submit",
        onSubmit: async (raw, signal) => {
          const reference = raw.trim();
          try {
            await host.api.invokeAction(
              "git.prs.link",
              {
                workspaceId: task.workspaceId,
                taskId: task.taskId,
                ...(repositoryId ? { repositoryId } : {}),
                body: { reference },
              },
              { signal },
            );
          } catch (error) {
            const f = readFailure(error);
            if (f.code === "not_found") {
              const number = reference.match(/(\d+)\D*$/)?.[1] ?? "";
              throw new Error(format(messages.prNotFound, { number }), { cause: error });
            }
            if (f.code === "validation" && f.field === "reference") {
              throw new Error(format(messages.notAPullRequest, { space }), { cause: error });
            }
            throw noticeError(gitNotice(error), messages, error);
          }
        },
      });
    },
  };
}
