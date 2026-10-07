import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { createPrList } from "../git/pr-list";
import { hostUi } from "../host-ui";
import { createIssuesPage } from "../issues/issues-page";
import type { LinksStore } from "../issues/links-store";
import { BUTTON, STACK } from "../layout";
import { en, type MessageKey, type Messages } from "../messages/en";
import type { ConnectionView } from "../settings/state";
import { PLUGIN_ID } from "../switch/enabled-events";

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

type Scope = "issues" | "prs";

/** The scope the URL asks for (?scope=prs), Issues by default (BR2.2). */
function scopeFromUrl(): Scope {
  try {
    return new URLSearchParams(window.location.search).get("scope") === "prs" ? "prs" : "issues";
  } catch {
    return "issues";
  }
}

/**
 * The Backlog page at /backlog (WF3, WF4, WF7): the one Integrations entry.
 * Connected, it shows Issues and Pull requests scopes like the GitHub
 * integration; otherwise an alert with the settings link (BR2.3).
 */
export function createBacklogPage(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): Component {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const { Alert, AlertDescription, AlertTitle, Button, Tabs, TabsContent, TabsList, TabsTrigger } =
    hostUi(host);
  const IssuesPage = createIssuesPage(host, messages, store);
  const PrList = createPrList(host, messages);

  return function BacklogPage() {
    const [workspaceId, setWorkspaceId] = useState(host.context.getActiveWorkspaceId());
    const [load, setLoad] = useState<PageInput["load"]>("loading");
    const [view, setView] = useState<ConnectionView | undefined>(undefined);
    const [scope, setScope] = useState<Scope>(scopeFromUrl);
    // A scope stays mounted once opened, so each keeps its filters (BR2.2).
    const [opened, setOpened] = useState<Record<Scope, boolean>>(() => ({
      issues: false,
      prs: false,
      [scope]: true,
    }));

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

    const switchScope = (next: Scope) => {
      setScope(next);
      setOpened((o) => ({ ...o, [next]: true }));
      host.navigate(next === "prs" ? "/backlog?scope=prs" : "/backlog", { replace: true });
    };

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
              <Button
                type="button"
                variant="outline"
                className={BUTTON}
                data-testid="backlog-page-retry"
                onClick={() => void reload()}
              >
                {messages.retry}
              </Button>
            </div>,
          ];
        case "connected":
          return (
            <Tabs
              value={scope}
              onValueChange={switchScope}
              className={STACK}
              data-testid="backlog-scope-tabs"
            >
              <TabsList aria-label={messages.scopeLabel}>
                <TabsTrigger value="issues" className={BUTTON} data-testid="backlog-scope-issues">
                  {messages.scopeIssues}
                </TabsTrigger>
                <TabsTrigger value="prs" className={BUTTON} data-testid="backlog-scope-prs">
                  {messages.scopePRs}
                </TabsTrigger>
              </TabsList>
              <TabsContent value="issues" forceMount hidden={scope !== "issues"}>
                {opened.issues ? <IssuesPage workspaceId={state.workspaceId} /> : null}
              </TabsContent>
              <TabsContent value="prs" forceMount hidden={scope !== "prs"}>
                {opened.prs ? (
                  <PrList workspaceId={state.workspaceId} selectedProjects={view?.selectedProjects ?? []} />
                ) : null}
              </TabsContent>
            </Tabs>
          );
        default: {
          const href = settingsHref(state.workspaceId);
          return (
            <Alert data-testid="backlog-page-alert">
              <AlertTitle>{messages.pageAlertTitle}</AlertTitle>
              <AlertDescription data-testid="backlog-page-status">
                {messages[STATUS_KEY[state.kind]]}
              </AlertDescription>
              <div>
                <Button
                  type="button"
                  variant="link"
                  className={`${BUTTON} px-0`}
                  data-testid="backlog-page-settings-link"
                  onClick={() => host.navigate(href)}
                >
                  {messages.openSettings}
                </Button>
              </div>
            </Alert>
          );
        }
      }
    })();

    return (
      <div data-testid="backlog-page" className={STACK}>
        {body}
      </div>
    );
  };
}
