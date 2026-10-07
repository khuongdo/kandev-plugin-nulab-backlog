import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { BUTTON } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { badgeDetail, badgeHref, badgeText } from "./issues-state";
import type { LinksStore } from "./links-store";

/** The task-card-tags slot props: { taskId, workspaceId, workflowStepId }. */
interface CardProps {
  slotProps?: unknown;
}

/**
 * The Backlog issue badge on a Kanban card (M6, task-card-tags slot). Its
 * text says the key and status, so it never relies on colour. When the issue
 * can be opened the badge is a link to it in a new tab (FR2); otherwise focus
 * or a tap shows the detail. Every card shares one issues.links.list call per
 * workspace.
 */
export function createIssueBadge(
  host: PluginHostApi,
  store: LinksStore,
  messages: Messages = en,
): Component<CardProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const { Button } = hostUi(host);
  const relative = (v: string) => host.utils?.formatRelativeTime?.(v) ?? v;

  return function IssueBadge({ slotProps }: CardProps) {
    const card = (slotProps ?? {}) as { taskId?: string; workspaceId?: string };
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
    const stop = (e: { stopPropagation(): void }) => e.stopPropagation(); // not the card, not a card drag
    const detailId = `backlog-issue-badge-detail-${taskId}`;
    if (href) {
      // BR2.2: a link to the issue in a new tab; the detail is its tooltip and description (R-01).
      return (
        <span className="inline-flex">
          <Button asChild variant="outline" size="xs" className={`${BUTTON} h-auto px-2 text-xs`}>
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              title={detail || undefined}
              aria-label={format(messages.issueBadgeSr, { label })}
              aria-describedby={detail ? detailId : undefined}
              data-testid={`backlog-issue-badge-${taskId}`}
              onClick={stop}
              onPointerDown={stop}
            >
              {label}
            </a>
          </Button>
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
        <Button
          type="button"
          variant="outline"
          size="xs"
          className={`${BUTTON} h-auto px-2 text-xs`}
          data-testid={`backlog-issue-badge-${taskId}`}
          aria-label={format(messages.issueBadgeSr, { label })}
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
        </Button>
        {open && detail ? (
          <span role="status" data-testid={detailId} className="text-xs">
            {detail}
          </span>
        ) : null}
      </span>
    );
  };
}
