import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { hostUi } from "../host-ui";
import { icon } from "../icons";
import { BUTTON } from "../layout";
import { en, type Messages } from "../messages/en";
import type { Notice } from "./state";

export interface MenuItem {
  testId: string;
  label: string;
  onSelect: () => void;
  disabled?: boolean;
}

export interface SectionParts {
  /** An inline error with Retry (BR7.2). */
  ListError: Component<{ testId: string; notice: Notice; onRetry: () => void }>;
  /** An empty state with an optional next step (BR7.1). */
  ListEmpty: Component<{ testId: string; title: string; children?: unknown }>;
  /** The row action menu: an icon-only ghost trigger with a name (NFR4). */
  RowMenu: Component<{ testId: string; label: string; items: MenuItem[] }>;
}

/** The pieces the watch and query sections share, in the GitHub style. */
export function createSectionParts(host: PluginHostApi, messages: Messages = en): SectionParts {
  const h = host.jsx;
  const {
    Alert,
    AlertDescription,
    Button,
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    Empty,
    EmptyContent,
    EmptyHeader,
    EmptyTitle,
  } = hostUi(host);

  return {
    ListError: ({ testId, notice, onRetry }) => (
      <Alert data-testid={`${testId}-error`} variant="destructive">
        <AlertDescription>{noticeText(notice, messages)}</AlertDescription>
        <Button
          type="button"
          variant="outline"
          size="sm"
          className={BUTTON}
          data-testid={`${testId}-retry`}
          onClick={onRetry}
        >
          {messages.retry}
        </Button>
      </Alert>
    ),
    ListEmpty: ({ testId, title, children }) => (
      <Empty data-testid={`${testId}-empty`}>
        <EmptyHeader>
          <EmptyTitle>{title}</EmptyTitle>
        </EmptyHeader>
        {children ? <EmptyContent>{children}</EmptyContent> : null}
      </Empty>
    ),
    RowMenu: ({ testId, label, items }) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className={BUTTON}
            data-testid={testId}
            aria-label={label}
          >
            {icon(host, "more")}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {items.map((item) => (
            <DropdownMenuItem
              key={item.testId}
              className={BUTTON}
              data-testid={item.testId}
              disabled={item.disabled}
              onSelect={item.onSelect}
            >
              {item.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    ),
  };
}
