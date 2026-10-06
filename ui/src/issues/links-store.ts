import type { PluginHostApi } from "@kandev/plugin-sdk";

import type { LinkView } from "./issues-state";

/** The workspace's issue links, shared by every card badge and the task menu. */
export interface LinksStore {
  /** The loaded links, or undefined before the first load. Synchronous. */
  get(workspaceId: string): readonly LinkView[] | undefined;
  /**
   * Loads when the links are missing or older than LINKS_REFRESH_MS; shares a
   * request in flight and makes none while backing off after a failure.
   */
  load(workspaceId: string): Promise<void>;
  /** Loads again now, for example after an unlink. */
  refresh(workspaceId: string): Promise<void>;
  /** While a workspace has listeners, its links refresh every LINKS_REFRESH_MS and on window focus. */
  subscribe(workspaceId: string, listener: () => void): () => void;
}

/**
 * How often mounted badges reload the links: the server sync's 1-minute tick
 * and minimum poll interval, so a synced status shows within a minute.
 */
export const LINKS_REFRESH_MS = 60_000;
const MAX_BACKOFF_MS = 10 * 60_000;

interface Entry {
  links?: readonly LinkView[];
  pending?: Promise<void>;
  failures: number;
  /** No automatic request before this time (fresh data, or a failure backoff). */
  nextAt: number;
  timer?: ReturnType<typeof setTimeout>;
  listeners: Set<() => void>;
}

export function createLinksStore(host: PluginHostApi): LinksStore {
  const entries = new Map<string, Entry>();
  const entry = (ws: string) => {
    let e = entries.get(ws);
    if (!e) {
      e = { failures: 0, nextAt: 0, listeners: new Set() };
      entries.set(ws, e);
    }
    return e;
  };

  /** Arms the next timed refresh while the workspace has listeners. */
  const schedule = (ws: string, e: Entry) => {
    clearTimeout(e.timer);
    e.timer = undefined;
    if (e.listeners.size > 0 && !e.pending) {
      e.timer = setTimeout(() => void auto(ws, true), Math.max(e.nextAt - Date.now(), 0));
    }
  };

  const fetch = (ws: string): Promise<void> => {
    const e = entry(ws);
    const p = host.api
      .invokeAction<{ links?: LinkView[] }>("issues.links.list", { workspaceId: ws })
      .then((reply) => {
        e.links = reply?.links ?? [];
        e.failures = 0;
        e.nextAt = Date.now() + LINKS_REFRESH_MS;
        e.listeners.forEach((l) => l());
      })
      .catch(() => {
        e.failures++;
        e.nextAt = Date.now() + Math.min(LINKS_REFRESH_MS * 2 ** (e.failures - 1), MAX_BACKOFF_MS);
        console.warn(`nulab-backlog: could not load the issue links of workspace ${ws}`);
      })
      .finally(() => {
        if (e.pending === p) e.pending = undefined;
        schedule(ws, e);
      });
    e.pending = p;
    schedule(ws, e);
    return p;
  };

  /** A request nobody asked for: never two at once, none while backing off; `stale` ignores fresh data. */
  const auto = (ws: string, stale = false): Promise<void> => {
    const e = entry(ws);
    if (e.pending) return e.pending;
    if (Date.now() < e.nextAt && (e.failures > 0 || !stale)) return Promise.resolve();
    return fetch(ws);
  };

  const onFocus = () => {
    if (document.visibilityState === "hidden") return;
    entries.forEach((e, ws) => {
      if (e.listeners.size > 0) void auto(ws, true);
    });
  };
  let watching = false;
  const watch = () => {
    const active = [...entries.values()].some((e) => e.listeners.size > 0);
    if (active === watching) return;
    watching = active;
    if (active) {
      window.addEventListener("focus", onFocus);
      document.addEventListener("visibilitychange", onFocus);
    } else {
      window.removeEventListener("focus", onFocus);
      document.removeEventListener("visibilitychange", onFocus);
    }
  };

  return {
    get: (ws) => entries.get(ws)?.links,
    load: (ws) => auto(ws),
    refresh: (ws) => fetch(ws),
    subscribe(ws, listener) {
      const e = entry(ws);
      e.listeners.add(listener);
      if (!e.timer) schedule(ws, e);
      watch();
      return () => {
        e.listeners.delete(listener);
        if (e.listeners.size === 0) schedule(ws, e); // clears the timer
        watch();
      };
    },
  };
}
