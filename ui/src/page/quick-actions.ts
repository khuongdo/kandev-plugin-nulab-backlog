import type { PluginHostApi } from "@kandev/plugin-sdk";

/** One "+ Task" menu entry (issues.quick_actions.*, FR2). */
export interface QuickAction {
  id: string;
  label: string;
  hint: string;
  icon: string;
  promptTemplate: string;
}

/** The workspace's quick actions per kind, defaults already filled in by the server. */
export interface QuickActions {
  issue: QuickAction[];
  pr: QuickAction[];
}

export type QuickActionKind = keyof QuickActions;

export const NO_ACTIONS: QuickActions = { issue: [], pr: [] };

/** The icons Kandev's start-task menu can draw (FR2.4). */
export const QUICK_ACTION_ICONS = ["eye", "message", "tool", "code", "search", "bug", "sparkle", "check"];

/** Kandev's task title limit (TASK_TITLE_MAX_LENGTH). */
export const TITLE_MAX = 60;

/**
 * Fills {{url}} and {{title}} (and the legacy {url}/{title}) in one pass, like
 * Kandev's interpolatePromptTemplate; unknown placeholders stay as written.
 */
export function interpolate(template: string, values: { url: string; title: string }): string {
  return template.replace(/\{\{?(url|title)\}\}?/g, (_m, key: "url" | "title") => values[key]);
}

/** "<label>: <title>", cut to Kandev's limit by characters with an ellipsis. */
export function taskTitle(label: string, title: string): string {
  const chars = Array.from(`${label}: ${title}`);
  return chars.length <= TITLE_MAX ? chars.join("") : `${chars.slice(0, TITLE_MAX - 1).join("")}…`;
}

/** Loads the quick actions once per page (NFR4); no actions when the call fails. */
export async function loadQuickActions(host: PluginHostApi, workspaceId: string): Promise<QuickActions> {
  try {
    const r = await host.api.invokeAction<Partial<QuickActions>>("issues.quick_actions.get", { workspaceId });
    return { issue: r?.issue ?? [], pr: r?.pr ?? [] };
  } catch {
    return NO_ACTIONS;
  }
}
