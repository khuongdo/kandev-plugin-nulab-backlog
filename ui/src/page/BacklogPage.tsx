import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { createIssuesPage } from "../issues/issues-page";
import type { LinksStore } from "../issues/links-store";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import type { ConnectionView } from "../settings/state";
import { PLUGIN_ID } from "../switch/enabled-events";

type AnyProps = Record<string, unknown>;

export interface PageInput {
  workspaceId?: string;
  load: "loading" | "failed" | "ready";
  view?: ConnectionView;
}

export type PageState =
  | { kind: "no_workspace" | "loading" | "failed" }
  | { kind: "off" | "not_connected" | "incomplete"; workspaceId: string }
  | { kind: "connected"; workspaceId: string; name: string; host: string };

/** Pure WF7 state function: which P1 state the page shows. */
export function pageState({ workspaceId, load, view }: PageInput): PageState {
  if (!workspaceId) return { kind: "no_workspace" };
  if (load !== "ready" || !view) return { kind: load === "failed" ? "failed" : "loading" };
  if (view.enabled === false) return { kind: "off", workspaceId };
  if (view.state === "connected")
    return { kind: "connected", workspaceId, name: view.connectedUserName ?? "", host: view.spaceHost ?? "" };
  if (view.state === "error") return { kind: "incomplete", workspaceId };
  return { kind: "not_connected", workspaceId };
}

/** Kandev's integration settings page for the plugin in one workspace. */
export function settingsHref(workspaceId: string): string {
  return `/settings/workspaces/${encodeURIComponent(workspaceId)}/integrations/${PLUGIN_ID}`;
}

const STATUS_KEY: Record<"off" | "not_connected" | "incomplete", MessageKey> = {
  off: "pageOff",
  not_connected: "pageNotConnected",
  incomplete: "incomplete",
};

/** The Backlog page at /backlog (WF7, BR7.7); once connected it lists the issues (U3, M2). */
export function createBacklogPage(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): Component {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const IssuesPage = createIssuesPage(host, messages, store);

  return function BacklogPage() {
    const [workspaceId, setWorkspaceId] = useState(host.context.getActiveWorkspaceId());
    const [load, setLoad] = useState<PageInput["load"]>("loading");
    const [view, setView] = useState<ConnectionView | undefined>(undefined);

    // Only the latest load may render: a late reply for an earlier load, or for
    // the previous workspace, is dropped (like the switch's live check).
    const latest = useRef(0);

    useEffect(() => host.context.subscribeActiveWorkspace(setWorkspaceId), []);

    const reload = useCallback(async () => {
      if (!workspaceId) return;
      const seq = ++latest.current;
      setLoad("loading");
      try {
        const next = await host.api.invokeAction<ConnectionView>("connection.get", { workspaceId });
        if (seq !== latest.current) return;
        setView(next);
        setLoad("ready");
      } catch {
        if (seq === latest.current) setLoad("failed");
      }
    }, [workspaceId]);

    useEffect(() => {
      void reload();
      return () => {
        latest.current++; // the workspace changed or the page closed
      };
    }, [reload]);

    const state = pageState({ workspaceId, load, view });
    const body = (() => {
      switch (state.kind) {
        case "no_workspace":
          return <p data-testid="backlog-page-status">{messages.pageNoWorkspace}</p>;
        case "loading":
          return <p data-testid="backlog-page-status">{messages.pageLoading}</p>;
        case "failed":
          return [
            <p key="s" data-testid="backlog-page-status">
              {messages.pageLoadFailed}
            </p>,
            <div key="r">
              <Button type="button" data-testid="backlog-page-retry" onClick={() => void reload()}>
                {messages.retry}
              </Button>
            </div>,
          ];
        default: {
          const href = settingsHref(state.workspaceId);
          const status =
            state.kind === "connected"
              ? format(messages.connectedAs, { name: state.name, host: state.host })
              : messages[STATUS_KEY[state.kind]];
          return [
            <p key="s" data-testid="backlog-page-status">
              {status}
            </p>,
            <a
              key="l"
              href={href}
              data-testid="backlog-page-settings-link"
              onClick={(e: { preventDefault(): void }) => {
                e.preventDefault();
                host.navigate(href);
              }}
            >
              {messages.openSettings}
            </a>,
            state.kind === "connected" ? (
              <div key="i">
                <IssuesPage workspaceId={state.workspaceId} />
              </div>
            ) : null,
          ];
        }
      }
    })();

    return (
      <div data-testid="backlog-page" className="flex flex-col gap-4">
        {body}
      </div>
    );
  };
}
