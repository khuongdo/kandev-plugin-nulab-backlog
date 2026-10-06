import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { en, format, type Messages } from "../messages/en";
import {
  connectedNotice,
  initialState,
  isOff,
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

const FIELD_IDS: Record<Field, string> = { spaceUrl: "backlog-space-url", apiKey: "backlog-api-key" };

// BR6.5: one vertical stack with one gap, and a smaller label-to-input gap,
// using the host's utility classes only (the plugin ships no CSS).
const STACK = "flex flex-col gap-4";
const FIELD = "flex flex-col gap-2";

/**
 * Builds the M1 settings screen on the host's React and UI kit. The plugin
 * bundles no React; `h` is the host's element factory.
 */
export function createSettingsScreen(host: PluginHostApi, messages: Messages = en): Component<Props> {
  const h = host.jsx; // the JSX factory (tsconfig jsxFactory "h")
  const { useCallback, useEffect, useRef, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const Input = host.ui.Input as Component<AnyProps>;
  const Label = host.ui.Label as Component<AnyProps>;
  const t = (notice: Notice) => format(messages[notice.key], notice.params);

  return function BacklogSettings({ workspaceId: routedWorkspaceId }: Props) {
    const workspaceId = routedWorkspaceId ?? host.context.getActiveWorkspaceId();
    const [state, setState] = useState<ScreenState>(initialState);
    const [spaceUrl, setSpaceUrl] = useState("");
    const [apiKey, setApiKey] = useState("");
    const busy = useRef(false);
    const lastInput = useRef<{ spaceUrl: string; apiKey: string } | undefined>(undefined);
    const dispatch = useCallback((e: ScreenEvent) => setState((s) => reduce(s, e)), []);

    const load = useCallback(async () => {
      dispatch({ type: "loadStarted" });
      try {
        if (!workspaceId) throw new Error("no workspace");
        const view = await host.api.invokeAction<ConnectionView>("connection.get", { workspaceId });
        dispatch({ type: "loaded", view });
      } catch {
        dispatch({ type: "loadFailed" });
      }
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

    const onSubmit = (e: { preventDefault(): void }) => {
      e.preventDefault();
      void connect({ spaceUrl, apiKey });
    };

    const field = (name: Field, label: string, value: string, set: (v: string) => void, extra: AnyProps) => {
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
    return (
      <div data-testid="backlog-settings" className={STACK}>
        {status ? <p data-testid="backlog-status">{t(status)}</p> : null}
        {state.view?.state === "error" ? <p data-testid="backlog-incomplete">{messages.incomplete}</p> : null}
        {off ? <p data-testid="backlog-off">{messages.integrationOff}</p> : null}
        {state.isMember && !status && !off ? (
          <p data-testid="backlog-member">{messages.notConnectedMember}</p>
        ) : null}
        {showForm ? (
          <form data-testid="backlog-connect-form" className={STACK} onSubmit={onSubmit} noValidate>
            {status ? <h3>{messages.replaceHeading}</h3> : null}
            {field("spaceUrl", messages.spaceUrlLabel, spaceUrl, setSpaceUrl, {
              placeholder: messages.spaceUrlPlaceholder,
              autoComplete: "url",
            })}
            {field("apiKey", messages.apiKeyLabel, apiKey, setApiKey, {
              type: "password",
              autoComplete: "off",
            })}
            <Button type="submit" data-testid="backlog-connect" disabled={state.connecting}>
              {state.connecting ? messages.connecting : messages.connect}
            </Button>
          </form>
        ) : null}
        {state.notice ? (
          <div data-testid="backlog-notice" className={STACK}>
            <p>{t(state.notice)}</p>
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
