import type { PluginHostApi } from "@kandev/plugin-sdk";

import { gitNotice } from "../git/git-state";
import type { Notice } from "./state";

export interface ListState<T> {
  items: T[];
  loading: boolean;
  /** Set when the last load failed (BR7.2). */
  error?: Notice;
  reload: () => void;
  setItems: (fn: (items: T[]) => T[]) => void;
}

/**
 * Loads one list action (`{ [field]: T[] }`) for a settings section. Each
 * section owns its list, so one failing list never hides the others.
 */
export function useActionList<T>(
  host: PluginHostApi,
  workspaceId: string,
  action: string,
  field: string,
): ListState<T> {
  const { useCallback, useEffect, useState } = host.React;
  const [items, setItems] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Notice | undefined>(undefined);

  const reload = useCallback(() => {
    setLoading(true);
    setError(undefined);
    host.api
      .invokeAction<Record<string, T[] | undefined>>(action, { workspaceId })
      .then((r) => setItems(r?.[field] ?? []))
      .catch((e: unknown) => setError(gitNotice(e)))
      .finally(() => setLoading(false));
  }, [workspaceId, action, field]);

  useEffect(reload, [reload]);
  return { items, loading, error, reload, setItems };
}
