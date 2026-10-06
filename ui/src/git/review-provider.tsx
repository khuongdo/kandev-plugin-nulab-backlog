import type { Component, PluginHostApi, ReviewSummary, ReviewTaskAssociation } from "@kandev/plugin-sdk";

import { en, type Messages } from "../messages/en";
import { PLUGIN_ID } from "../switch/enabled-events";
import { gitNotice, noticeText } from "./git-state";

/** A git.prs.status summary: Kandev's ReviewSummary plus what the panel shows. */
export interface BacklogSummary extends ReviewSummary {
  base?: string;
  branch?: string;
  assignee?: string;
}

interface PanelProps {
  panelId: string;
  presentation: "desktop" | "mobile";
  workspaceId: string;
  taskId: string;
  reviewKey: string;
  connectionScope: string;
  repositoryId: string;
  changeRequestNumber: string | number;
}

const EMPTY: readonly never[] = Object.freeze([]);

/** A keyed snapshot store with listeners, for Kandev's useSyncExternalStore-style reads. */
function store<T>() {
  const values = new Map<string, readonly T[]>();
  const listeners = new Map<string, Set<() => void>>();
  return {
    get: (key: string): readonly T[] => values.get(key) ?? EMPTY,
    set(key: string, next: readonly T[]) {
      values.set(key, next);
      listeners.get(key)?.forEach((l) => l());
    },
    subscribe(key: string, listener: () => void) {
      const set = listeners.get(key) ?? new Set();
      set.add(listener);
      listeners.set(key, set);
      return () => {
        set.delete(listener);
      };
    },
  };
}

/**
 * The PR badge, status and unlink, published through Kandev's review provider
 * (M6, US5.4, US5.2). Kandev renders the indicators and refreshes them every
 * 90 s, so the plugin runs no status poller.
 */
export function createReviewProvider(host: PluginHostApi, messages: Messages = en) {
  const h = host.jsx;
  const summaries = store<BacklogSummary>();
  const associations = store<ReviewTaskAssociation>();
  const Detail = host.ui.ChangeRequestDetail as Component<Record<string, unknown>>;
  const workspace = () => host.context.getActiveWorkspaceId() ?? "";

  const { useEffect, useState } = host.React;

  const ReviewPanel = ({ presentation, taskId, reviewKey }: PanelProps) => {
    const [, setVersion] = useState(0);
    useEffect(() => summaries.subscribe(taskId, () => setVersion((v) => v + 1)), [taskId]);
    const s = summaries.get(taskId).find((x) => x.reviewKey === reviewKey);
    const detail = s
      ? {
          providerId: PLUGIN_ID,
          reviewKey: s.reviewKey,
          number: s.changeRequestNumber,
          title: s.title,
          url: s.url,
          state: s.state,
          author: { name: s.assignee ?? "" },
          sourceBranch: s.branch ?? "",
          targetBranch: s.base ?? "",
          additions: 0,
          deletions: 0,
          reviews: [],
          requestedReviewers: [],
          checks: [],
          comments: [],
        }
      : null;
    return <Detail detail={detail} presentation={presentation} error={s ? null : messages.statusUnknown} />;
  };

  return {
    id: PLUGIN_ID,
    label: messages.integrationLabel,
    changeRequestNoun: messages.reviewNoun,
    order: 50,
    getSnapshot: (taskId: string) => summaries.get(taskId),
    subscribe: (taskId: string, listener: () => void) => summaries.subscribe(taskId, listener),
    refresh: async (taskId: string, signal: AbortSignal) => {
      try {
        const reply = await host.api.invokeAction<{ summaries?: BacklogSummary[] }>(
          "git.prs.status",
          { workspaceId: workspace(), taskId, body: {} },
          { signal },
        );
        summaries.set(taskId, reply?.summaries ?? []);
      } catch (error) {
        if (signal.aborted) return;
        const reason = noticeText(gitNotice(error), messages);
        summaries.set(
          taskId,
          summaries.get(taskId).map((s) => ({
            ...s,
            statusBadge: { label: messages.statusUnknown },
            taskStatus: s.taskStatus && { ...s.taskStatus, error: reason },
          })),
        );
      }
    },
    getAssociationSnapshot: (workspaceId: string) => associations.get(workspaceId),
    subscribeAssociations: (workspaceId: string, listener: () => void) =>
      associations.subscribe(workspaceId, listener),
    refreshAssociations: async (workspaceId: string, signal: AbortSignal) => {
      const reply = await host.api.invokeAction<{ associations?: ReviewTaskAssociation[] }>(
        "git.links.list",
        { workspaceId },
        { signal },
      );
      associations.set(workspaceId, reply?.associations ?? []);
    },
    unlink: async (ctx: { workspaceId: string; taskId: string; reviewKey: string; signal: AbortSignal }) => {
      await host.api.invokeAction(
        "git.prs.unlink",
        { workspaceId: ctx.workspaceId, taskId: ctx.taskId, body: { reviewKey: ctx.reviewKey } },
        { signal: ctx.signal },
      );
      summaries.set(
        ctx.taskId,
        summaries.get(ctx.taskId).filter((s) => s.reviewKey !== ctx.reviewKey),
      );
      associations.set(
        ctx.workspaceId,
        associations
          .get(ctx.workspaceId)
          .filter((a) => !(a.taskId === ctx.taskId && a.reviewKey === ctx.reviewKey)),
      );
    },
    ReviewPanel,
  };
}
