import type { PluginHostApi } from "@kandev/plugin-sdk";

import { en, format, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";

/** One task linked to an issue row. */
export interface TaskLink {
  taskId: string;
  taskKey?: string;
}

/** One row of issues.list (C5 IssuePage item). */
export interface IssueItem {
  issueKey: string;
  summary: string;
  status: string;
  statusId: number;
  assignee?: string;
  updatedAt: string;
  url: string;
  linkedTasks: TaskLink[];
}

/** The issues.list reply. */
export interface IssuePage {
  items?: IssueItem[];
  total?: number;
  page?: number;
  pageSize?: number;
  refreshedAt?: string;
}

/** One link of issues.links.list. */
export interface LinkView {
  taskId: string;
  taskKey?: string;
  issueKey: string;
  spaceHost?: string;
  state: string;
  status?: string;
  statusUpdatedAt?: string;
  stale?: boolean;
  unavailable?: boolean;
  url?: string;
}

/** What the issue list shows for a failed load. */
export type IssuesFailure =
  | { kind: "not_connected" | "no_project" | "sign_in_again" }
  | { kind: "rate_limited"; retryAfterSeconds: number; notice: Notice }
  | { kind: "error"; notice: Notice };

/** Maps a failed issues.list to its state; a server body is never shown. */
export function issuesFailure(error: unknown): IssuesFailure {
  const f = readFailure(error);
  switch (f.code) {
    case "reconnect_required":
      return { kind: "sign_in_again" };
    case "integration_disabled":
      return { kind: "not_connected" };
    case "validation":
      if (f.field === "projectKeys") return { kind: "no_project" };
      return { kind: "error", notice: { key: "errorInput" } };
    case "rate_limited": {
      const seconds = f.retryAfterSeconds ?? 60;
      return {
        kind: "rate_limited",
        retryAfterSeconds: seconds,
        notice: { key: "rateLimitedRetrying", params: { seconds } },
      };
    }
    case "unreachable":
      return { kind: "error", notice: { key: "unreachable", retry: true } };
  }
  return { kind: "error", notice: { key: "issuesLoadFailed", retry: true } };
}

/** The notice for a failed issue action (create, link, refresh, unlink). */
export function issueNotice(error: unknown): Notice {
  const f = readFailure(error);
  switch (f.code) {
    case "reconnect_required":
      return { key: "reconnectRequired" };
    case "rate_limited":
      return { key: "rateLimited", params: { seconds: f.retryAfterSeconds ?? 60 } };
    case "unreachable":
      return { key: "unreachable" };
    case "integration_disabled":
      return { key: "pageOff" };
    case "not_found":
      return { key: "gitNotFound" };
    case "conflict":
      return { key: "gitConflict" };
    case "validation":
      return { key: f.field === "projectKeys" ? "issuesNoProject" : "errorInput" };
  }
  return { key: f.status === 403 ? "adminOnly" : "actionFailed" };
}

/** "Showing 41-57 of 57"; empty when there is nothing to show. */
export function showingText(
  p: { page: number; pageSize: number; total: number },
  messages: Messages = en,
): string {
  const from = (p.page - 1) * p.pageSize + 1;
  if (p.total === 0 || from > p.total) return "";
  return format(messages.issuesShowing, { from, to: Math.min(p.page * p.pageSize, p.total), total: p.total });
}

/** The badge text: "PROJ-120 · Resolved", never colour alone (M6). */
export function badgeText(l: LinkView, messages: Messages = en): string {
  if (l.state === "not_connected") return format(messages.issueBadgeNotConnected, { key: l.issueKey });
  if (l.unavailable) return format(messages.issueBadgeUnavailable, { key: l.issueKey });
  return format(messages.issueBadge, { key: l.issueKey, status: l.status ?? "" });
}

/** The badge detail shown on focus or tap. */
export function badgeDetail(l: LinkView, relative: (v: string) => string, messages: Messages = en): string {
  if (l.state === "not_connected") return format(messages.reconnectToRestore, { host: l.spaceHost ?? "" });
  const parts = [];
  if (l.statusUpdatedAt) parts.push(format(messages.issueUpdatedAt, { time: relative(l.statusUpdatedAt) }));
  if (l.stale) parts.push(messages.mayBeOutOfDate);
  return parts.join(" · ");
}

/** Joins words with ", " and a final " and ". */
function joinList(words: string[], messages: Messages): string {
  if (words.length < 2) return words.join("");
  return words.slice(0, -1).join(messages.listComma) + messages.listAnd + words[words.length - 1];
}

/** The screen-reader text of a row. */
export function rowLabel(item: IssueItem, messages: Messages = en): string {
  const parts = [format(messages.rowSr, { key: item.issueKey, title: item.summary, status: item.status })];
  if (item.assignee) parts.push(format(messages.rowSrAssignee, { name: item.assignee }));
  if (item.linkedTasks.length > 0) {
    const tasks = joinList(
      item.linkedTasks.map((t) => t.taskKey ?? t.taskId),
      messages,
    );
    parts.push(format(messages.rowSrLinked, { tasks }));
  }
  return parts.join(messages.listComma);
}

/**
 * "N issue links, M PR links and K PR watches will be turned off." for the
 * U2 confirm dialogs (AC1.8.1, AC1.9.1), of projectKeys only when given. A
 * count that cannot be loaded counts as 0; empty when nothing is affected.
 */
export async function loadImpactText(
  host: PluginHostApi,
  workspaceId: string,
  projectKeys?: string[],
  messages: Messages = en,
): Promise<string> {
  const input = { workspaceId, body: projectKeys ? { projectKeys } : {} };
  const [issues, git] = await Promise.all([
    host.api.invokeAction<{ issueLinks?: number }>("issues.impact", input).catch(() => undefined),
    host.api
      .invokeAction<{ prLinks?: number; prWatches?: number }>("git.impact", input)
      .catch(() => undefined),
  ]);
  const counts = { issues: issues?.issueLinks ?? 0, links: git?.prLinks ?? 0, watches: git?.prWatches ?? 0 };
  return counts.issues || counts.links || counts.watches ? format(messages.impactAll, counts) : "";
}

/** Kandev's page of a task. */
export const taskHref = (taskId: string) => `/t/${encodeURIComponent(taskId)}`;
