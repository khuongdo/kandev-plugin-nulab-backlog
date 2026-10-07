import type { PluginHostApi } from "@kandev/plugin-sdk";

import { en, format, type MessageKey, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";

/** The text of a notice in the given catalogue. */
export function noticeText(n: Notice, messages: Messages = en): string {
  return format(messages[n.key], n.params);
}

const STATE_KEYS: Record<string, MessageKey> = {
  open: "stateOpen",
  closed: "stateClosed",
  merged: "stateMerged",
  draft: "stateDraft", // A3: GitHub and GitLab drafts
  declined: "stateDeclined", // A3: Bitbucket
};

/** The name of a pull request state. */
export function stateText(state: string, messages: Messages = en): string {
  return messages[STATE_KEYS[state] ?? "statusUnknown"];
}

/** The screen-reader text of a PR badge: "Pull request #42, Open, assignee Lan". */
export function prBadgeLabel(
  s: { changeRequestNumber: number | string; state: string; assignee?: string },
  messages: Messages = en,
): string {
  const params = {
    number: s.changeRequestNumber,
    state: stateText(s.state, messages),
    assignee: s.assignee ?? "",
  };
  return format(s.assignee ? messages.prBadgeSr : messages.prBadgeSrNoAssignee, params);
}

const WATCH_KEYS: Record<string, MessageKey> = {
  active: "watchActive",
  paused: "watchPaused",
  not_connected: "watchNotConnected",
};

/** The catalogue key of a watch state. */
export function watchStatusKey(state: string): MessageKey {
  return WATCH_KEYS[state] ?? "watchNotConnected";
}

/** "Created 10/25 tasks, the rest in later cycles" while tasks are pending (AC6.2.2). */
export function watchProgress(w: { createdCount: number; pendingCount: number }): Notice | undefined {
  if (!w.pendingCount) return undefined;
  return {
    key: "watchProgress",
    params: { created: w.createdCount, total: w.createdCount + w.pendingCount },
  };
}

/** The notice for a failed U4 action. retry marks failures a Retry can fix. */
export function gitNotice(error: unknown): Notice {
  const f = readFailure(error);
  switch (f.code) {
    case "reconnect_required":
      return { key: "reconnectRequired" };
    case "rate_limited":
      return { key: "rateLimited", params: { seconds: f.retryAfterSeconds ?? 60 }, retry: true };
    case "unreachable":
      return { key: "unreachable", retry: true };
    case "integration_disabled":
      return { key: "pageOff" };
    case "not_found":
      return { key: "gitNotFound" };
    case "conflict":
      return { key: "gitConflict" };
    case "validation":
      return { key: "errorInput" };
  }
  return { key: "actionFailed", retry: true };
}

/** An Error carrying the notice text, for host flows that show a rejection. */
export function noticeError(n: Notice, messages: Messages = en, cause?: unknown): Error {
  return new Error(noticeText(n, messages), { cause });
}

/** One repository choice of a selected project, valued "PROJ/repo". */
export interface RepoOption {
  value: string;
  label: string;
}

/** Loads the selected projects' repositories for a watch or query form. */
// ponytail: first page only (50 repositories); page with nextCursor if a space has more.
export async function loadRepoOptions(
  host: PluginHostApi,
  workspaceId: string,
  messages: Messages = en,
): Promise<RepoOption[]> {
  const reply = await host.api.invokeAction<{
    repositories?: { ownerOrProject: string; repositoryName: string }[];
  }>("git.repositories.list", { workspaceId, body: {} });
  return (reply?.repositories ?? []).map((r) => ({
    value: `${r.ownerOrProject}/${r.repositoryName}`,
    label: format(messages.watchRepo, { project: r.ownerOrProject, repo: r.repositoryName }),
  }));
}

/** Splits "PROJ/repo" into the project key and repository name. */
export function splitRepo(value: string): { projectKey: string; repoName: string } {
  const i = value.indexOf("/");
  return { projectKey: value.slice(0, i), repoName: value.slice(i + 1) };
}

/**
 * "N PR links and M PR watches will be turned off." for the confirm dialogs
 * (AC1.8.1, AC1.9.1), of projectKeys only when given. Empty when nothing is
 * affected or the counts cannot be loaded: the dialog still works.
 */
export async function loadImpact(
  host: PluginHostApi,
  workspaceId: string,
  projectKeys?: string[],
  messages: Messages = en,
): Promise<string> {
  try {
    const r = await host.api.invokeAction<{ prLinks?: number; prWatches?: number }>("git.impact", {
      workspaceId,
      body: projectKeys ? { projectKeys } : {},
    });
    const links = r?.prLinks ?? 0;
    const watches = r?.prWatches ?? 0;
    return links || watches ? format(messages.impactPR, { links, watches }) : "";
  } catch {
    return "";
  }
}

/** Appends an impact sentence to a dialog body. */
export function withImpact(body: string, impact: string): string {
  return impact ? `${body} ${impact}` : body;
}

/** A source control provider; "backlog" is Backlog Git (FR1.1). */
export type ScmProvider = "github" | "gitlab" | "bitbucket";

const PROVIDER_KEYS: Record<string, MessageKey> = {
  backlog: "providerBacklog",
  github: "providerGithub",
  gitlab: "providerGitlab",
  bitbucket: "providerBitbucket",
};

/** The display name of a provider. */
export function providerName(provider: string, messages: Messages = en): string {
  return messages[PROVIDER_KEYS[provider] ?? "providerBacklog"];
}

/** Backlog project to repositories of one provider (FR3.1). */
export interface Mapping {
  projectKey: string;
  repos: string[];
}

/** One provider as scm.providers.list returns it; never a token (NFR1). */
export interface ProviderView {
  provider: ScmProvider;
  state: string;
  account?: string;
  lastError?: string;
  mappings: Mapping[];
}

/** The notice of a failed scm.* action, in provider words (FR2.4, NFR5). */
export function scmNotice(error: unknown): Notice {
  const f = readFailure(error);
  switch (f.code) {
    case "reconnect_required":
      return { key: "scmTokenRefused" };
    case "rate_limited":
      return { key: "scmRateLimited", params: { seconds: f.retryAfterSeconds ?? 60 }, retry: true };
    case "unreachable":
      return { key: "scmUnreachable", retry: true };
    case "conflict":
      return { key: "scmUnmapped" };
    case "validation":
      if (f.field === "token") return { key: "scmNoToken" };
  }
  return gitNotice(error);
}

/** The providers of a workspace; none when they cannot be read. */
export async function loadProviders(host: PluginHostApi, workspaceId: string): Promise<ProviderView[]> {
  try {
    const r = await host.api.invokeAction<{ providers?: ProviderView[] }>("scm.providers.list", {
      workspaceId,
    });
    return r?.providers ?? [];
  } catch {
    return [];
  }
}

/** The providers whose token works, which can list pull requests. */
export function usableProviders(views: ProviderView[]): ProviderView[] {
  return views.filter((v) => v.state === "connected");
}

/**
 * The mapped repositories of the selected projects as options valued
 * "PROJ:owner/name" (FR3.3); projects no longer selected are hidden (FR6.2).
 */
export function scmRepoOptions(
  view: ProviderView | undefined,
  selected: string[],
  messages: Messages = en,
): RepoOption[] {
  return (view?.mappings ?? [])
    .filter((m) => selected.includes(m.projectKey))
    .flatMap((m) =>
      m.repos.map((repo) => ({
        value: `${m.projectKey}:${repo}`,
        label: format(messages.scmRepoOption, { project: m.projectKey, repo }),
      })),
    );
}

/** Splits "PROJ:owner/name" into the project key and repository. */
export function splitScmRepo(value: string): { projectKey: string; repo: string } {
  const i = value.indexOf(":");
  return { projectKey: value.slice(0, i), repo: value.slice(i + 1) };
}
