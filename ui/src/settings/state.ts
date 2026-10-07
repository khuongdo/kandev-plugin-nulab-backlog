import type { MessageKey } from "../messages/en";
import type { OAuthOutcome } from "./oauth";

/** The ConnectionView fields the M1 screen reads (contract C5). */
export interface ConnectionView {
  connected: boolean;
  /** IntegrationSwitch value; absent means off (opt-in, intent 261007-opt-in-default). */
  enabled?: boolean;
  state: "not_connected" | "connected" | "error" | string;
  spaceHost?: string;
  connectedUserName?: string;
  hasApiKey: boolean;
  authMethod?: string;
  selectedProjects?: string[];
  /** Set on a connect reply that restored an earlier connection (M12). */
  restored?: boolean;
  /** U4: a Git credential is stored for this space (US5.5). */
  hasGitCredential?: boolean;
  /** U4: the Git check of a connection.test reply: ok, invalid or untested (AC5.5.2). */
  gitCheck?: string;
}

export type Field = "spaceUrl" | "apiKey" | "projectKeys";

/** A message to show, with its interpolation params. */
export interface Notice {
  key: MessageKey;
  params?: Record<string, string | number>;
  retry?: boolean;
}

export interface ScreenState {
  load: "loading" | "failed" | "ready";
  view?: ConnectionView;
  isMember: boolean;
  connecting: boolean;
  fieldError?: { field: Field; key: MessageKey };
  notice?: Notice;
  announcement?: Notice;
}

export type ScreenEvent =
  | { type: "loadStarted" }
  | { type: "loaded"; view: ConnectionView }
  | { type: "loadFailed" }
  | { type: "connectStarted" }
  | { type: "connected"; view: ConnectionView }
  | { type: "connectFailed"; error: unknown }
  | { type: "enabledChanged"; enabled: boolean }
  | { type: "oauthReturned"; outcome: OAuthOutcome; restored: boolean }
  | { type: "announce"; notice: Notice }
  | { type: "viewChanged"; view: ConnectionView };

export const initialState: ScreenState = { load: "loading", isMember: false, connecting: false };

/** The text that describes a connected view, or undefined. */
export function connectedNotice(view: ConnectionView | undefined): Notice | undefined {
  if (view?.state !== "connected") return undefined;
  return { key: "connectedAs", params: { name: view.connectedUserName ?? "", host: view.spaceHost ?? "" } };
}

/** True when the OAuth refresh was refused and the user must sign in again (AC1.4.3). */
export function isSignInAgain(view: ConnectionView | undefined): boolean {
  return view?.state === "sign_in_again";
}

/** The M12 notice for a connection that came back. */
export function restoredNotice(view: ConnectionView | undefined): Notice {
  return { key: "restoredNotice", params: { host: view?.spaceHost ?? "" } };
}

const OAUTH_NOTICES: Record<OAuthOutcome, MessageKey> = {
  connected: "oauthConnected",
  cancelled: "oauthCancelled",
  failed: "oauthFailed",
};

/** Pure transition function for the M1 screen. */
export function reduce(state: ScreenState, event: ScreenEvent): ScreenState {
  switch (event.type) {
    case "loadStarted":
      return { ...state, load: "loading" };
    case "loaded":
      return { ...state, load: "ready", view: event.view };
    case "loadFailed":
      return { ...state, load: "failed" };
    case "connectStarted":
      return {
        ...state,
        connecting: true,
        fieldError: undefined,
        notice: undefined,
        announcement: undefined,
      };
    case "connected": {
      const notice = event.view.restored ? restoredNotice(event.view) : undefined;
      return {
        ...state,
        connecting: false,
        view: event.view,
        notice,
        announcement: notice ?? connectedNotice(event.view),
      };
    }
    case "oauthReturned": {
      const notice: Notice =
        event.outcome === "connected" && event.restored
          ? restoredNotice(state.view)
          : { key: OAUTH_NOTICES[event.outcome] };
      return { ...state, notice, announcement: notice };
    }
    case "announce":
      return { ...state, announcement: event.notice };
    case "viewChanged":
      return { ...state, view: event.view };
    case "enabledChanged":
      // The switch was saved: only `enabled` changed, the connection did not (BR7.4).
      return state.view ? { ...state, view: { ...state.view, enabled: event.enabled } } : state;
    case "connectFailed": {
      if (readFailure(event.error).code === "integration_disabled") {
        // Turned off meanwhile: show the Off state, as if loaded with enabled false.
        const view = state.view ? { ...state.view, enabled: false } : undefined;
        return { ...state, connecting: false, view, announcement: { key: "integrationOff" } };
      }
      const failure = classify(event.error);
      return {
        ...state,
        connecting: false,
        ...failure,
        announcement: failure.fieldError ? { key: failure.fieldError.key } : failure.notice,
      };
    }
  }
}

interface ActionFailure {
  status?: number;
  code?: string;
  field?: string;
  retryAfterSeconds?: number;
  /** U4: the open pull request of a create conflict (AC5.3.4). */
  pullRequestNumber?: number;
}

/** Reads the host ApiError shape: { status, body: { error: { code, field, retryAfterSeconds } } }. */
export function readFailure(error: unknown): ActionFailure {
  if (typeof error !== "object" || error === null) return {};
  const { status, body } = error as { status?: unknown; body?: unknown };
  const inner = (body as { error?: unknown } | null | undefined)?.error;
  const e = typeof inner === "object" && inner !== null ? (inner as Record<string, unknown>) : {};
  return {
    status: typeof status === "number" ? status : undefined,
    code: typeof e.code === "string" ? e.code : undefined,
    field: typeof e.field === "string" ? e.field : undefined,
    retryAfterSeconds: typeof e.retryAfterSeconds === "number" ? e.retryAfterSeconds : undefined,
    pullRequestNumber: typeof e.pullRequestNumber === "number" ? e.pullRequestNumber : undefined,
  };
}

/** True while Backlog is off for a loaded view; opt-in, so only an explicit on is on. */
export function isOff(view: ConnectionView | undefined): boolean {
  return view !== undefined && view.enabled !== true;
}

function classify(error: unknown): Pick<ScreenState, "fieldError" | "notice" | "isMember"> {
  if (readFailure(error).status === 403) return { isMember: true };
  const f = failureNotice(error, { fallback: "connectFailed" });
  return { isMember: false, fieldError: f.fieldError, notice: f.notice };
}

/** What to show for a failed action. */
export interface Failure {
  notice?: Notice;
  fieldError?: { field: Field; key: MessageKey };
  /** Set for rate_limited: the seconds to wait before a retry. */
  retryAfterSeconds?: number;
}

const FIELD_ERRORS: Record<Field, MessageKey> = {
  spaceUrl: "errorSpaceUrl",
  apiKey: "errorApiKey",
  projectKeys: "errorProjectKeys",
};

/**
 * Maps an ActionError to its message. `retrying` picks the countdown text a
 * read shows while it waits to retry (AC8.4.4); writes never retry by themselves.
 */
export function failureNotice(
  error: unknown,
  opts: { retrying?: boolean; fallback?: MessageKey } = {},
): Failure {
  const f = readFailure(error);
  switch (f.code) {
    case "validation":
      if (f.field === "oauth") return { notice: { key: "oauthNotConfigured" } };
      if (f.field && f.field in FIELD_ERRORS) {
        const field = f.field as Field;
        return { fieldError: { field, key: FIELD_ERRORS[field] } };
      }
      return { notice: { key: "errorInput" } };
    case "reconnect_required":
      return { notice: { key: "reconnectRequired" } };
    case "unreachable":
      return { notice: { key: "unreachable", retry: true } };
    case "rate_limited": {
      const seconds = f.retryAfterSeconds ?? 60;
      const key = opts.retrying ? "rateLimitedRetrying" : "rateLimited";
      return { notice: { key, params: { seconds } }, retryAfterSeconds: seconds };
    }
    case "conflict":
      return { notice: { key: "busy" } };
    case "integration_disabled":
      return { notice: { key: "integrationOff" } };
  }
  if (f.status === 403) return { notice: { key: "adminOnly" } };
  return { notice: { key: opts.fallback ?? "actionFailed" } };
}
