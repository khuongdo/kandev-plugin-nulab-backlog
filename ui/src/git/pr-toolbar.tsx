import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON, TOOLBAR } from "../layout";
import { en, format, type Messages } from "../messages/en";

export interface PrToolbarProps {
  title: string;
  /** The result count; hidden while unknown. */
  count?: number;
  loading: boolean;
  lastFetchedAt?: string;
  onRefresh: () => void;
  refreshDisabled?: boolean;
  /** The filter slot between the title and the refresh button. */
  children?: unknown;
  /** The test id prefix: backlog-prs (default) or backlog-scm-prs. */
  idPrefix?: string;
}

/**
 * The Pull requests toolbar, laid out like the host's IntegrationListToolbar
 * (BR5.1) without its query box: Backlog's pull request API has no search.
 * Desktop: one bordered row with the title and count, the filters, then last
 * fetched and a ghost refresh at the right end. Phones: the title row, the
 * filters at full width, then a row with the count on the left and last
 * fetched and refresh on the right (BR4.8).
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
  }: PrToolbarProps) {
    return (
      <div className={TOOLBAR} data-testid={`${idPrefix}-toolbar`}>
        <div className="flex min-w-0 items-baseline gap-2">
          <h2 data-testid={`${idPrefix}-list`} className="truncate text-sm font-semibold">
            {title}
          </h2>
          {count !== undefined ? (
            <span
              className="hidden text-xs tabular-nums text-muted-foreground md:inline"
              data-testid={`${idPrefix}-count`}
            >
              {count}
            </span>
          ) : null}
        </div>
        {children}
        <div className="flex min-w-0 items-center justify-end gap-2 md:ml-auto">
          {count !== undefined ? (
            <span
              className="mr-auto shrink-0 whitespace-nowrap text-xs text-muted-foreground md:hidden"
              data-testid={`${idPrefix}-count-mobile`}
            >
              {format(messages.resultCount, { count })}
            </span>
          ) : null}
          {lastFetchedAt && !loading ? (
            <span className="truncate text-xs text-muted-foreground" data-testid={`${idPrefix}-updated`}>
              {format(messages.updatedAt, { time: relative(lastFetchedAt) })}
            </span>
          ) : null}
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className={BUTTON}
            data-testid={`${idPrefix}-refresh`}
            aria-label={messages.refresh}
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
