import type { PluginHostApi, TaskContext } from "@kandev/plugin-sdk";

import { noticeError } from "../git/git-state";
import { en, format, type Messages } from "../messages/en";
import { readFailure } from "../settings/state";
import { PLUGIN_ID } from "../switch/enabled-events";
import { issueNotice } from "./issues-state";
import type { LinksStore } from "./links-store";

const ISSUE_KEY = /^[A-Z][A-Z0-9_]*-[1-9][0-9]*$/;
/** The only hosts a pasted link may use (project Mandated rule, NFR1). */
const BACKLOG_DOMAINS = ["backlog.com", "backlog.jp", "backlogtool.com"];

/**
 * The issue key in what the user typed (FR2): an issue key in any case, or an
 * `https` Backlog issue link `https://<space>/view/<KEY>` under backlog.com,
 * backlog.jp or backlogtool.com (FR2.3). Anything else is undefined.
 */
export function parseIssueReference(raw: string): string | undefined {
  const text = raw.trim();
  let candidate = text;
  if (/^[a-z][a-z0-9+.-]*:/i.test(text)) {
    let url: URL;
    try {
      url = new URL(text);
    } catch {
      return undefined;
    }
    const host = url.hostname.toLowerCase();
    const isBacklog = BACKLOG_DOMAINS.some((d) => host.endsWith(`.${d}`));
    const path = url.pathname.match(/^\/view\/([^/]+)$/);
    if (url.protocol !== "https:" || !isBacklog || !path) return undefined;
    candidate = path[1]!; // still percent-encoded: "%" fails the key pattern
  }
  const key = candidate.toUpperCase();
  return ISSUE_KEY.test(key) ? key : undefined;
}

/** The inline message for a failed issues.link; never the Backlog response (NFR1). */
function linkError(error: unknown, key: string, messages: Messages): Error {
  const f = readFailure(error);
  if (f.code === "not_found") return new Error(format(messages.issueNotFound, { key }), { cause: error });
  if (f.code === "conflict") return new Error(messages.issueLinkConflict, { cause: error });
  if (f.code === "validation" && f.field === "issueKey") {
    const project = key.slice(0, key.lastIndexOf("-"));
    return new Error(format(messages.issueProjectNotSelected, { project }), { cause: error });
  }
  return noticeError(issueNotice(error), messages, error);
}

/**
 * "Link Backlog issue" in the task's Link menu, next to GitHub Issue (FR1).
 * Kandev owns the dialog, its empty check, submit state, inline error and
 * toast; the plugin supplies the copy and the submit, which takes an issue
 * key or an https Backlog issue link (FR2). Hidden while the task is linked,
 * decided from the shared links store like Unlink (FR3).
 */
export function createIssueLinkAction(host: PluginHostApi, store: LinksStore, messages: Messages = en) {
  return {
    id: `${PLUGIN_ID}-link-issue`,
    label: messages.linkIssueLabel,
    placement: "link" as const,
    singleTaskOnly: true,
    visible({ workspaceId, taskId }: TaskContext): boolean {
      const links = store.get(workspaceId);
      if (!links) {
        void store.load(workspaceId);
        return false;
      }
      return !links.some((l) => l.taskId === taskId);
    },
    run: async ({ workspaceId, taskId }: TaskContext) => {
      host.openTaskLinkDialog({
        title: messages.linkIssueTitle,
        description: messages.linkIssueDescription,
        inputLabel: messages.linkIssueInput,
        placeholder: messages.linkIssuePlaceholder,
        emptyError: messages.linkIssueEmpty,
        failureMessage: messages.linkIssueFailed,
        successMessage: messages.linkIssueSuccess,
        inputTestId: "backlog-link-issue-input",
        errorTestId: "backlog-link-issue-error",
        submitTestId: "backlog-link-issue-submit",
        onSubmit: async (raw, signal) => {
          const issueKey = parseIssueReference(raw);
          if (!issueKey) throw new Error(messages.notAnIssueReference);
          try {
            await host.api.invokeAction(
              "issues.link",
              { workspaceId, taskId, body: { issueKey } },
              { signal },
            );
          } catch (error) {
            throw linkError(error, issueKey, messages);
          }
          // The link is saved; a failed refresh only delays the badge (FR1.4).
          await store.refresh(workspaceId).catch(() => undefined);
        },
      });
    },
  };
}
