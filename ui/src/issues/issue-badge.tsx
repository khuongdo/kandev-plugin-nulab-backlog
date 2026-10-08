import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { BUTTON } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { badgeDetail, badgeHover, badgeHref, badgeText } from "./issues-state";
import type { LinksStore } from "./links-store";

/**
 * The slot props: task-card-tags { taskId, workspaceId, workflowStepId },
 * task-row-metadata adds { surface }, chat-top-bar adds { presentation }.
 */
interface CardProps {
  slotProps?: unknown;
}

/**
 * The Backlog issue badge on a Kanban card (task-card-tags), on task rows of
 * Home > Tasks and the sidebar (task-row-metadata, FR1) and in the task top
 * bar (chat-top-bar, FR5). Its text says the key and status, so it never
 * relies on colour. Hover or keyboard focus shows a card with the key,
 * summary and status (FR1.4, NFR4). When the issue can be opened the badge is
 * a link to it in a new tab (FR1.5); otherwise focus or a tap shows the
 * detail. Every badge shares one issues.links.list call per workspace (NFR2).
 */
export function createIssueBadge(
  host: PluginHostApi,
  store: LinksStore,
  messages: Messages = en,
): Component<CardProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const { Button, Tooltip, TooltipTrigger, TooltipContent } = hostUi(host);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;

  return function IssueBadge({ slotProps }: CardProps) {
    const card = (slotProps ?? {}) as { taskId?: string; workspaceId?: string; presentation?: string };
    const ws = card.workspaceId ?? "";
    const taskId = card.taskId ?? "";
    const [, setVersion] = useState(0);
    const [open, setOpen] = useState(false);

    useEffect(() => {
      if (!ws) return;
      const stop = store.subscribe(ws, () => setVersion((v) => v + 1));
      void store.load(ws);
      return stop;
    }, [ws]);

    const link = store.get(ws)?.find((l) => l.taskId === taskId);
    if (!link) return null;
    const label = badgeText(link, messages);
    const detail = badgeDetail(link, relative, messages);
    const href = badgeHref(link);
    const srLabel = link.summary
      ? format(messages.issueBadgeSrSummary, { label, summary: link.summary })
      : format(messages.issueBadgeSr, { label });
    // FR5.5: phones put the top-bar slot in the host menu with 44px touch targets.
    const size = card.presentation === "mobile" ? "min-h-11 min-w-11 px-3" : "h-auto px-2";
    const stop = (e: { stopPropagation(): void }) => e.stopPropagation(); // not the card, not a card drag
    const detailId = `backlog-issue-badge-detail-${taskId}`;
    const hoverCard = (badge: unknown, withDetail: boolean) => (
      <Tooltip>
        <TooltipTrigger asChild>{badge}</TooltipTrigger>
        <TooltipContent data-testid={`backlog-issue-badge-hover-${taskId}`} className="flex flex-col gap-0.5">
          {badgeHover(link).map((line, i) => (
            <span key={i} data-hover-line="" className={i === 0 ? "font-medium" : undefined}>
              {line}
            </span>
          ))}
          {withDetail && detail ? <span className="text-muted-foreground">{detail}</span> : null}
        </TooltipContent>
      </Tooltip>
    );
    if (href) {
      // FR1.5: a link to the issue in a new tab; the detail is also its screen reader description (R-01).
      return (
        <span className="inline-flex">
          {hoverCard(
            <Button asChild variant="outline" size="xs" className={`${BUTTON} ${size} text-xs`}>
              <a
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={srLabel}
                aria-describedby={detail ? detailId : undefined}
                data-testid={`backlog-issue-badge-${taskId}`}
                onClick={stop}
                onPointerDown={stop}
              >
                {label}
              </a>
            </Button>,
            true,
          )}
          {detail ? (
            <span id={detailId} className="sr-only">
              {detail}
            </span>
          ) : null}
        </span>
      );
    }
    // BR2.3: not openable; focus or a tap shows the detail, such as the reconnect hint.
    return (
      <span className="inline-flex flex-col gap-1">
        {hoverCard(
          <Button
            type="button"
            variant="outline"
            size="xs"
            className={`${BUTTON} ${size} text-xs`}
            data-testid={`backlog-issue-badge-${taskId}`}
            aria-label={srLabel}
            aria-expanded={open}
            onFocus={() => setOpen(true)}
            onBlur={() => setOpen(false)}
            onPointerDown={stop}
            onClick={(e: { stopPropagation(): void }) => {
              stop(e);
              setOpen((o) => !o);
            }}
          >
            {label}
          </Button>,
          false,
        )}
        {open && detail ? (
          <span role="status" data-testid={detailId} className="text-xs">
            {detail}
          </span>
        ) : null}
      </span>
    );
  };
}
