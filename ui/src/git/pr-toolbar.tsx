import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON, ROW, TOOLBAR } from "../layout";
import { en, format, type Messages } from "../messages/en";

export interface PrToolbarProps {
  title: string;
  /** The result count; hidden while unknown. */
  count?: number;
  loading: boolean;
  lastFetchedAt?: string;
  onRefresh: () => void;
  refreshDisabled?: boolean;
  /** The filter controls between the title and the refresh button. */
  children?: unknown;
  /** The test id prefix: backlog-prs (default) or backlog-issues. */
  idPrefix?: string;
  /** The refresh button's name; Refresh by default. */
  refreshLabel?: string;
  /** A ref for the title, which pagination focuses; the title is then focusable. */
  headingRef?: unknown;
}

/**
 * A list toolbar laid out like the host's IntegrationListToolbar (FR5.2):
 * one bordered row with the title and count on the left, the filters in the
 * middle, and last fetched with a ghost refresh at the right end. The Pull
 * requests list has no free-text search: Backlog's pull request API has none.
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
    idPrefix = "backlog-prs",
    refreshLabel = messages.refresh,
    headingRef,
  }: PrToolbarProps) {
    return (
      <div className={TOOLBAR} data-testid={`${idPrefix}-toolbar`}>
        <div className="flex items-baseline gap-2">
          <h2
            ref={headingRef}
            tabIndex={headingRef ? -1 : undefined}
            data-testid={`${idPrefix}-list`}
            className="text-sm font-semibold"
          >
            {title}
          </h2>
          {count !== undefined ? (
            <span className="text-xs text-muted-foreground" data-testid={`${idPrefix}-count`}>
              {format(messages.resultCount, { count })}
            </span>
          ) : null}
        </div>
        <div className={`${ROW} md:flex-1`}>{children}</div>
        <div className="flex items-center gap-2 md:ml-auto">
          {lastFetchedAt && !loading ? (
            <span className="text-xs text-muted-foreground" data-testid={`${idPrefix}-updated`}>
              {format(messages.updatedAt, { time: relative(lastFetchedAt) })}
            </span>
          ) : null}
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className={BUTTON}
            data-testid={`${idPrefix}-refresh`}
            aria-label={refreshLabel}
            disabled={loading || refreshDisabled}
            onClick={onRefresh}
          >
            {icon(host, "refresh", loading ? "h-4 w-4 animate-spin" : "h-4 w-4")}
          </Button>
        </div>
      </div>
    );
  };
}
