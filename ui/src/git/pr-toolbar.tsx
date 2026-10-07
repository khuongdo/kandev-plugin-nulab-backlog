import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON, ROW } from "../layout";
import { en, format, type Messages } from "../messages/en";

export interface PrToolbarProps {
  title: string;
  /** The result count; hidden while unknown. */
  count?: number;
  loading: boolean;
  lastFetchedAt?: string;
  onRefresh: () => void;
  refreshDisabled?: boolean;
  /** The filter controls under the title row. */
  children?: unknown;
}

/**
 * The Pull requests toolbar, laid out like the host's IntegrationListToolbar
 * (title and count left, last fetched and a ghost refresh right) without
 * its free-text search: Backlog's pull request API has none.
 */
export function createPrToolbar(host: PluginHostApi, messages: Messages = en): Component<PrToolbarProps> {
  const h = host.jsx;
  const { Button } = hostUi(host);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;

  return function PrToolbar({
    title,
    count,
    loading,
    lastFetchedAt,
    onRefresh,
    refreshDisabled,
    children,
  }: PrToolbarProps) {
    return (
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-baseline gap-2">
            <h2 className="text-lg font-semibold">{title}</h2>
            {count !== undefined ? (
              <span className="text-xs text-muted-foreground" data-testid="backlog-prs-count">
                {format(messages.resultCount, { count })}
              </span>
            ) : null}
          </div>
          <div className="flex items-center gap-2">
            {lastFetchedAt && !loading ? (
              <span className="text-xs text-muted-foreground" data-testid="backlog-prs-updated">
                {format(messages.updatedAt, { time: relative(lastFetchedAt) })}
              </span>
            ) : null}
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className={BUTTON}
              data-testid="backlog-prs-refresh"
              aria-label={messages.refresh}
              disabled={loading || refreshDisabled}
              onClick={onRefresh}
            >
              {icon(host, "refresh", loading ? "h-4 w-4 animate-spin" : "h-4 w-4")}
            </Button>
          </div>
        </div>
        <div className={ROW}>{children}</div>
      </div>
    );
  };
}
