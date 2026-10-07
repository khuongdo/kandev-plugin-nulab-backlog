import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { withImpact } from "../git/git-state";
import { loadImpactText } from "../issues/issues-state";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { createConnectedPanel } from "./connected-panel";
import { normalizeHost, startOAuth, takeOAuthReturn } from "./oauth";
import { createProjectPicker } from "./project-picker";
import {
  connectedNotice,
  initialState,
  isOff,
  isSignInAgain,
  reduce,
  type ConnectionView,
  type Field,
  type Notice,
  type ScreenEvent,
  type ScreenState,
} from "./state";
import { subscribeEnabled } from "../switch/enabled-events";

type Props = { workspaceId?: string };
type AnyProps = Record<string, unknown>;

type FormField = Exclude<Field, "projectKeys">;
type Method = "api_key" | "oauth";

const FIELD_IDS: Record<FormField, string> = { spaceUrl: "backlog-space-url", apiKey: "backlog-api-key" };

// BR6.5: one vertical stack with one gap, and a smaller label-to-input gap,
// using the host's utility classes only (the plugin ships no CSS).
const STACK = "flex flex-col gap-4";
const FIELD = "flex flex-col gap-2";
const ROW = "flex gap-2";

/**
 * Builds the M1 settings screen on the host's React and UI kit. The plugin
 * bundles no React; `h` is the host's element factory.
 */
export function createSettingsScreen(host: PluginHostApi, messages: Messages = en): Component<Props> {
  const hostApi = host; // onSubmit shadows `host` with the typed space host
  const h = host.jsx; // the JSX factory (tsconfig jsxFactory "h")
  const { useCallback, useEffect, useRef, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const Input = host.ui.Input as Component<AnyProps>;
  const Label = host.ui.Label as Component<AnyProps>;
  const t = (notice: Notice) => format(messages[notice.key], notice.params);
  const ConfirmDialog = createConfirmDialog(host, messages);
  const ConnectedPanel = createConnectedPanel(host, messages);
  const ProjectPicker = createProjectPicker(host, messages);

  return function BacklogSettings({ workspaceId: routedWorkspaceId }: Props) {
    const workspaceId = routedWorkspaceId ?? host.context.getActiveWorkspaceId();
    const [state, setState] = useState<ScreenState>(initialState);
    const [spaceUrl, setSpaceUrl] = useState("");
    const [apiKey, setApiKey] = useState("");
    const busy = useRef(false);
    const lastInput = useRef<{ spaceUrl: string; apiKey: string } | undefined>(undefined);
    const [method, setMethod] = useState<Method>("api_key");
    const [confirm, setConfirm] = useState<
      { kind: "replace" | "changeSpace"; host: string; impact?: string } | undefined
    >(undefined);
    const oauthChecked = useRef(false);
    const dispatch = useCallback((e: ScreenEvent) => setState((s) => reduce(s, e)), []);
    const announce = useCallback((notice: Notice) => dispatch({ type: "announce", notice }), []);
    const onView = useCallback((view: ConnectionView) => dispatch({ type: "viewChanged", view }), []);

    const load = useCallback(async () => {
      dispatch({ type: "loadStarted" });
      try {
        if (!workspaceId) throw new Error("no workspace");
        const view = await host.api.invokeAction<ConnectionView>("connection.get", { workspaceId });
        dispatch({ type: "loaded", view });
      } catch {
        dispatch({ type: "loadFailed" });
        return;
      }
      if (oauthChecked.current) return;
      oauthChecked.current = true; // the ?oauth= result is shown once (US1.3)
      const back = takeOAuthReturn();
      if (back) dispatch({ type: "oauthReturned", ...back });
    }, [workspaceId]);

    useEffect(() => {
      void load();
    }, [load]);

    // Follow the card switch for this workspace after it is saved (WF6 step 5).
    useEffect(
      () =>
        subscribeEnabled((ws, enabled) => {
          if (ws === workspaceId) dispatch({ type: "enabledChanged", enabled });
        }),
      [workspaceId],
    );

    const connect = async (input: { spaceUrl: string; apiKey: string }) => {
      if (busy.current) return; // BR6.1: ignore clicks while a Connect runs
      busy.current = true;
      lastInput.current = input;
      setApiKey(""); // BR6.3: the key input is empty after every Connect
      dispatch({ type: "connectStarted" });
      try {
        const view = await host.api.invokeAction<ConnectionView>("connection.connect_api_key", {
          workspaceId,
          body: input,
        });
        lastInput.current = undefined;
        dispatch({ type: "connected", view });
      } catch (error) {
        dispatch({ type: "connectFailed", error });
      } finally {
        busy.current = false;
      }
    };

    // Starts OAuth and leaves the page; only a failure comes back here.
    const signIn = async (url: string) => {
      if (busy.current || !workspaceId) return;
      busy.current = true;
      dispatch({ type: "connectStarted" });
      try {
        await startOAuth(host, workspaceId, url);
      } catch (error) {
        busy.current = false;
        dispatch({ type: "connectFailed", error });
      }
    };

    const run = () => (method === "oauth" ? signIn(spaceUrl) : connect({ spaceUrl, apiKey }));

    const onSubmit = (e: { preventDefault(): void }) => {
      e.preventDefault();
      const view = state.view;
      if (view && (view.state === "connected" || isSignInAgain(view))) {
        // A replacement asks first; another host is a space change (US1.6, US1.8).
        const host = normalizeHost(spaceUrl);
        const kind = host === view.spaceHost ? "replace" : "changeSpace";
        setConfirm({ kind, host });
        // U3/U4: a space change turns off the old space's issue links, PR links and watches (AC1.8.1).
        if (kind === "changeSpace" && workspaceId) {
          void loadImpactText(hostApi, workspaceId, undefined, messages).then((impact) =>
            setConfirm((c) => (c && c.host === host ? { ...c, impact } : c)),
          );
        }
        return;
      }
      void run();
    };

    const field = (
      name: FormField,
      label: string,
      value: string,
      set: (v: string) => void,
      extra: AnyProps,
    ) => {
      const id = FIELD_IDS[name];
      const error = state.fieldError?.field === name ? state.fieldError : undefined;
      return (
        <div className={FIELD}>
          <Label htmlFor={id}>{label}</Label>
          <Input
            id={id}
            data-testid={id}
            value={value}
            onChange={(e: { target: { value: string } }) => set(e.target.value)}
            aria-invalid={error ? "true" : undefined}
            aria-describedby={error ? `${id}-error` : undefined}
            {...extra}
          />
          {error ? (
            <p id={`${id}-error`} data-testid={`${id}-error`}>
              {messages[error.key]}
            </p>
          ) : null}
        </div>
      );
    };

    const announcement = (
      <div role="status" aria-live="polite" data-testid="backlog-announcement">
        {state.announcement ? t(state.announcement) : ""}
      </div>
    );

    if (state.load === "loading") {
      return (
        <div data-testid="backlog-settings" className={STACK}>
          <p>{messages.loading}</p>
          {announcement}
        </div>
      );
    }
    if (state.load === "failed") {
      return (
        <div data-testid="backlog-settings" className={STACK}>
          <p>{messages.loadFailed}</p>
          <Button type="button" data-testid="backlog-load-retry" onClick={() => void load()}>
            {messages.retry}
          </Button>
          {announcement}
        </div>
      );
    }

    const status = connectedNotice(state.view);
    const off = isOff(state.view);
    const showForm = !state.isMember && !off;
    const signInAgain = isSignInAgain(state.view) && !off;
    const replacing = Boolean(status) || signInAgain;

    const methodChoice = (
      <div role="radiogroup" aria-labelledby="backlog-method-label" className={FIELD}>
        <span id="backlog-method-label">{messages.signInMethodLabel}</span>
        {(["api_key", "oauth"] as const).map((m) => (
          <label key={m} className={ROW}>
            <input
              type="radio"
              name="backlog-method"
              value={m}
              data-testid={m === "oauth" ? "backlog-method-oauth" : "backlog-method-api-key"}
              checked={method === m}
              onChange={() => setMethod(m)}
            />
            {m === "oauth" ? messages.methodOAuth : messages.methodApiKey}
          </label>
        ))}
      </div>
    );
    return (
      <div data-testid="backlog-settings" className={STACK}>
        {status ? <p data-testid="backlog-status">{t(status)}</p> : null}
        {state.view?.state === "error" ? <p data-testid="backlog-incomplete">{messages.incomplete}</p> : null}
        {off ? <p data-testid="backlog-off">{messages.integrationOff}</p> : null}
        {state.isMember && !status && !off ? (
          <p data-testid="backlog-member">{messages.notConnectedMember}</p>
        ) : null}
        {status && !off && !state.isMember && workspaceId ? (
          <ConnectedPanel workspaceId={workspaceId} onView={onView} announce={announce} view={state.view} />
        ) : null}
        {status && !off && !state.isMember && workspaceId ? (
          // A new space (a change or a restore) or a new account on the same space
          // remounts the picker so it reloads and keeps no list or checkbox of the old one
          // (R-02, R-12). Not the epoch: a project save bumps it. The view carries no user
          // id, so the account is told apart by its display name.
          <ProjectPicker
            key={`${state.view?.spaceHost ?? ""}:${state.view?.connectedUserName ?? ""}`}
            workspaceId={workspaceId}
            onView={onView}
            announce={announce}
          />
        ) : null}
        {signInAgain ? (
          <div className={STACK}>
            <p data-testid="backlog-sign-in-again-notice">{messages.signInAgainNotice}</p>
            <Button
              type="button"
              data-testid="backlog-sign-in-again"
              disabled={state.connecting}
              onClick={() => void signIn(state.view?.spaceHost ?? "")}
            >
              {messages.signInAgain}
            </Button>
          </div>
        ) : null}
        {showForm ? (
          <form data-testid="backlog-connect-form" className={STACK} onSubmit={onSubmit} noValidate>
            {status ? <h3>{messages.replaceHeading}</h3> : null}
            {methodChoice}
            {field("spaceUrl", messages.spaceUrlLabel, spaceUrl, setSpaceUrl, {
              placeholder: messages.spaceUrlPlaceholder,
              autoComplete: "url",
            })}
            {method === "api_key"
              ? field("apiKey", messages.apiKeyLabel, apiKey, setApiKey, {
                  type: "password",
                  autoComplete: "off",
                })
              : null}
            {method === "api_key" ? (
              <Button type="submit" data-testid="backlog-connect" disabled={state.connecting}>
                {state.connecting
                  ? messages.connecting
                  : replacing
                    ? messages.replaceCredentials
                    : messages.connect}
              </Button>
            ) : (
              <Button type="submit" data-testid="backlog-sign-in-nulab" disabled={state.connecting}>
                {state.connecting ? messages.connecting : messages.signInWithNulab}
              </Button>
            )}
          </form>
        ) : null}
        {confirm ? (
          <ConfirmDialog
            testId={confirm.kind === "replace" ? "backlog-replace-dialog" : "backlog-change-space-dialog"}
            title={confirm.kind === "replace" ? messages.replaceTitle : messages.changeSpaceTitle}
            body={
              confirm.kind === "replace"
                ? messages.replaceBody
                : withImpact(format(messages.changeSpaceBody, { host: confirm.host }), confirm.impact ?? "")
            }
            confirmLabel={messages.confirm}
            onConfirm={run}
            onClose={() => setConfirm(undefined)}
          />
        ) : null}
        {state.notice ? (
          <div data-testid="backlog-notice" className={STACK}>
            <p>{t(state.notice)}</p>
            {state.notice.key === "restoredNotice" ? (
              <div className={STACK}>
                <p>{messages.restoredWatches}</p>
                <Button
                  type="button"
                  data-testid="backlog-review-watches"
                  onClick={() => host.navigate("/backlog/watches")}
                >
                  {messages.reviewWatches}
                </Button>
              </div>
            ) : null}
            {state.notice.retry && lastInput.current ? (
              <Button
                type="button"
                data-testid="backlog-connect-retry"
                onClick={() => lastInput.current && void connect(lastInput.current)}
              >
                {messages.retry}
              </Button>
            ) : null}
          </div>
        ) : null}
        {announcement}
      </div>
    );
  };
}
