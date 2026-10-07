import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { gitNotice, noticeText } from "../git/git-state";
import { createPrList } from "../git/pr-list";
import type { Query } from "../git/save-query-dialog";
import { hostUi } from "../host-ui";
import { createIssuesPage } from "../issues/issues-page";
import type { IssueQuery } from "../issues/issues-state";
import type { LinksStore } from "../issues/links-store";
import { BUTTON, STACK } from "../layout";
import { en, type MessageKey, type Messages } from "../messages/en";
import type { ConnectionView, Notice } from "../settings/state";
import { PLUGIN_ID } from "../switch/enabled-events";
import { loadQuickActions, NO_ACTIONS, type QuickActions } from "./quick-actions";

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

/** The built-in preset of each kind (FR3.1, FR4.1); not stored. */
const PRESET: Record<Scope, string> = { issues: "assigned-open", prs: "open-assigned" };

interface SavedQueries {
  issues: IssueQuery[];
  prs: Query[];
}

/** A scope bar pick; n grows on every pick so picking the same query again re-applies it. */
interface Pick {
  source: "preset" | "saved";
  id: string;
  n: number;
}

const SAVED_ACTIONS: Record<Scope, string> = { issues: "issues.queries", prs: "git.queries" };

async function listQueries<T>(host: PluginHostApi, workspaceId: string, action: string): Promise<T[]> {
  try {
    const r = await host.api.invokeAction<{ queries?: T[] }>(`${action}.list`, { workspaceId });
    return r?.queries ?? [];
  } catch {
    return []; // the presets still work without saved queries
  }
}

/**
 * The Backlog page at /backlog (WF3, WF4, WF7; FR3-FR5): the one Integrations
 * entry. Connected, it shows the host scope bar (Issues / Pull requests, the
 * built-in preset, the Saved menu with a default star) above the list, which
 * opens on the starred saved query or the preset; otherwise an alert with the
 * settings link (BR2.3).
 */
export function createBacklogPage(
  host: PluginHostApi,
  messages: Messages = en,
  store?: LinksStore,
): Component {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const { Alert, AlertDescription, AlertTitle, Button, IntegrationScopeBar } = hostUi(host);
  const IssuesPage = createIssuesPage(host, messages, store);
  const PrList = createPrList(host, messages);
  const presets: Record<Scope, { value: string; label: string; group: "inbox" }[]> = {
    issues: [{ value: PRESET.issues, label: messages.presetAssignedOpen, group: "inbox" }],
    prs: [{ value: PRESET.prs, label: messages.presetOpenAssigned, group: "inbox" }],
  };

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
    const [quickActions, setQuickActions] = useState<QuickActions>(NO_ACTIONS);
    const [saved, setSaved] = useState<SavedQueries | undefined>(undefined);
    const [picks, setPicks] = useState<Record<Scope, Pick> | undefined>(undefined);
    const [saveRequest, setSaveRequest] = useState<Record<Scope, number>>({ issues: 0, prs: 0 });
    const [pending, setPending] = useState<string | null>(null);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);

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
    const connectedWs = state.kind === "connected" ? state.workspaceId : undefined;
    const selectedProjects = view?.selectedProjects ?? [];

    // A saved query is usable when its project is still selected (BR2.4).
    const usable = (kind: Scope, id: string, list: SavedQueries) => {
      const q = (list[kind] as { id?: string; projectKey?: string }[]).find((x) => x.id === id);
      if (!q) return false;
      return kind === "issues" && !q.projectKey ? true : selectedProjects.includes(q.projectKey ?? "");
    };

    // Quick actions and the saved queries are read once per page and workspace (NFR4);
    // the lists mount after, on the starred query or the preset (FR3.2, FR4.5).
    useEffect(() => {
      if (!connectedWs) return;
      let live = true;
      setSaved(undefined);
      setPicks(undefined);
      void loadQuickActions(host, connectedWs).then((a) => live && setQuickActions(a));
      void Promise.all([
        listQueries<IssueQuery>(host, connectedWs, SAVED_ACTIONS.issues),
        listQueries<Query>(host, connectedWs, SAVED_ACTIONS.prs),
      ]).then(([issues, prs]) => {
        if (!live) return;
        const list = { issues, prs };
        const first = (kind: Scope): Pick => {
          const star = (list[kind] as { id?: string; isDefault?: boolean }[]).find(
            (q) => q.isDefault && q.id && usable(kind, q.id, list),
          );
          return star?.id
            ? { source: "saved", id: star.id, n: 0 }
            : { source: "preset", id: PRESET[kind], n: 0 };
        };
        setSaved(list);
        setPicks({ issues: first("issues"), prs: first("prs") });
      });
      return () => {
        live = false;
      };
    }, [connectedWs]);

    const switchScope = (next: Scope) => {
      setScope(next);
      setOpened((o) => ({ ...o, [next]: true }));
      host.navigate(next === "prs" ? "/backlog?scope=prs" : "/backlog", { replace: true });
    };

    const pick = (kind: Scope, source: Pick["source"], id: string) =>
      setPicks((p) => (p ? { ...p, [kind]: { source, id, n: p[kind].n + 1 } } : p));

    const onSelect = (sel: { kind: Scope; source: Pick["source"]; id: string }) => {
      if (sel.source === "saved" && saved && !usable(sel.kind, sel.id, saved)) return;
      pick(sel.kind, sel.source, sel.id);
    };

    const toggleDefault = async (id: string) => {
      const q = (saved?.[scope] as { id?: string; isDefault?: boolean }[] | undefined)?.find(
        (x) => x.id === id,
      );
      if (!q || !connectedWs) return;
      setPending(id);
      setNotice(undefined);
      try {
        const r = await host.api.invokeAction<{ queries?: unknown[] }>(
          `${SAVED_ACTIONS[scope]}.set_default`,
          {
            workspaceId: connectedWs,
            body: { id, isDefault: !q.isDefault },
          },
        );
        setSaved((s) => (s ? { ...s, [scope]: r?.queries ?? s[scope] } : s));
      } catch (e) {
        setNotice(gitNotice(e));
      }
      setPending(null);
    };

    const deleteSaved = async (id: string) => {
      if (!connectedWs) return;
      setNotice(undefined);
      try {
        await host.api.invokeAction(`${SAVED_ACTIONS[scope]}.delete`, {
          workspaceId: connectedWs,
          body: { id },
        });
        setSaved((s) =>
          s ? { ...s, [scope]: (s[scope] as { id?: string }[]).filter((q) => q.id !== id) } : s,
        );
        if (picks?.[scope].source === "saved" && picks[scope].id === id) pick(scope, "preset", PRESET[scope]);
      } catch (e) {
        setNotice(gitNotice(e));
      }
    };

    const addSaved = (kind: Scope) => (q: IssueQuery | Query) =>
      setSaved((s) => (s ? { ...s, [kind]: [...s[kind], q] } : s));

    const selection = <T extends { id?: string }>(kind: Scope, list: T[]) => {
      const p = picks![kind];
      return {
        key: `${p.source}:${p.id}:${p.n}`,
        query: p.source === "saved" ? list.find((q) => q.id === p.id) : undefined,
      };
    };

    const savedPresets = saved
      ? [
          ...saved.issues.map((q) => ({
            id: q.id ?? "",
            kind: "issues" as Scope,
            label: q.name,
            isDefault: q.isDefault,
          })),
          ...saved.prs.map((q) => ({
            id: q.id ?? "",
            kind: "prs" as Scope,
            label: selectedProjects.includes(q.projectKey)
              ? q.name
              : `${q.name} (${messages.projectNotSelected})`,
            isDefault: q.isDefault,
          })),
        ]
      : [];

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
          if (!saved || !picks) return <p data-testid="backlog-page-status">{messages.pageLoading}</p>;
          return null;
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

    if (state.kind !== "connected" || !saved || !picks) {
      return (
        <div data-testid="backlog-page" className={`${STACK} px-4 py-4 sm:px-6`}>
          {body}
        </div>
      );
    }

    return (
      <div data-testid="backlog-page" className="flex min-w-0 flex-1 flex-col">
        <IntegrationScopeBar
          testId="backlog-scope-bar"
          savedMenuTestId="backlog-saved-menu"
          kinds={[
            { value: "issues", label: messages.scopeIssues },
            { value: "prs", label: messages.scopePRs },
          ]}
          selected={{ kind: scope, source: picks[scope].source, id: picks[scope].id }}
          onSelect={onSelect}
          onKindChange={switchScope}
          presetsByKind={(kind: Scope) => presets[kind]}
          savedPresets={savedPresets}
          onDeleteSaved={(id: string) => void deleteSaved(id)}
          canSaveCurrent
          onSaveCurrent={() => setSaveRequest((r) => ({ ...r, [scope]: r[scope] + 1 }))}
          onToggleSavedDefault={(id: string) => void toggleDefault(id)}
          defaultMutationPendingId={pending}
        />
        {notice ? (
          <p role="alert" className="px-4 sm:px-6" data-testid="backlog-scope-notice">
            {noticeText(notice, messages)}
          </p>
        ) : null}
        <div hidden={scope !== "issues"} data-testid="backlog-scope-issues-panel">
          {opened.issues ? (
            <IssuesPage
              workspaceId={state.workspaceId}
              quickActions={quickActions.issue}
              selection={selection("issues", saved.issues)}
              onSavedQuery={addSaved("issues")}
              saveRequest={saveRequest.issues}
            />
          ) : null}
        </div>
        <div hidden={scope !== "prs"} data-testid="backlog-scope-prs-panel">
          {opened.prs ? (
            <PrList
              workspaceId={state.workspaceId}
              quickActions={quickActions.pr}
              selection={selection("prs", saved.prs)}
              onSavedQuery={addSaved("prs")}
              saveRequest={saveRequest.prs}
            />
          ) : null}
        </div>
      </div>
    );
  };
}
