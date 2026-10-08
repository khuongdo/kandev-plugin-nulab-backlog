import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { BUTTON, FILTER_TRIGGER } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";

const STATUSES = ["open", "closed", "merged"] as const;
const STATUS_KEYS: Record<(typeof STATUSES)[number], MessageKey> = {
  open: "stateOpen",
  closed: "stateClosed",
  merged: "stateMerged",
};

export interface StatusMultiFilterProps {
  /** The picked statuses; never empty. */
  value: string[];
  onToggle: (status: string) => void;
  /** The test id prefix: backlog-prs (default) or backlog-scm-prs. */
  idPrefix?: string;
}

/**
 * The pull request status filter (BR5.3): one compact "Status (n)" trigger
 * like the other dropdowns, whose popover lists Open, Closed and Merged as
 * checkboxes. The last picked status stays focusable but is aria-disabled and
 * ignores the toggle, so at least one stays picked (R-05).
 */
export function createStatusMultiFilter(
  host: PluginHostApi,
  messages: Messages = en,
): Component<StatusMultiFilterProps> {
  const h = host.jsx;
  const { Button, Checkbox, Label, Popover, PopoverContent, PopoverTrigger } = hostUi(host);

  return function StatusMultiFilter({ value, onToggle, idPrefix = "backlog-prs" }: StatusMultiFilterProps) {
    return (
      <Popover>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className={`${BUTTON} ${FILTER_TRIGGER} justify-start font-normal`}
            data-testid={`${idPrefix}-status`}
            aria-haspopup="dialog"
          >
            {format(messages.prsStatusCount, { count: value.length })}
          </Button>
        </PopoverTrigger>
        <PopoverContent align="start" className="w-48" aria-label={messages.watchStatusLegend}>
          <div className="flex flex-col gap-2">
            {STATUSES.map((s) => {
              const checked = value.includes(s);
              const last = checked && value.length === 1;
              return (
                <div key={s} className="flex items-center gap-2">
                  <Checkbox
                    id={`${idPrefix}-status-${s}`}
                    data-testid={`${idPrefix}-status-${s}`}
                    checked={checked}
                    aria-disabled={last || undefined}
                    onCheckedChange={() => {
                      if (!last) onToggle(s);
                    }}
                  />
                  <Label htmlFor={`${idPrefix}-status-${s}`}>{messages[STATUS_KEYS[s]]}</Label>
                </div>
              );
            })}
          </div>
        </PopoverContent>
      </Popover>
    );
  };
}
