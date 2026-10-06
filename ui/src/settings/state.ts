import type { MessageKey } from "../messages/en";

/** The ConnectionView fields the M1 screen reads (contract C5). */
export interface ConnectionView {
  connected: boolean;
  /** IntegrationSwitch value; absent means on (BR7.1). */
  enabled?: boolean;
  state: "not_connected" | "connected" | "error" | string;
  spaceHost?: string;
  connectedUserName?: string;
  hasApiKey: boolean;
}

export type Field = "spaceUrl" | "apiKey";

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
  | { type: "enabledChanged"; enabled: boolean };

export const initialState: ScreenState = { load: "loading", isMember: false, connecting: false };

/** The text that describes a connected view, or undefined. */
export function connectedNotice(view: ConnectionView | undefined): Notice | undefined {
  if (view?.state !== "connected") return undefined;
  return { key: "connectedAs", params: { name: view.connectedUserName ?? "", host: view.spaceHost ?? "" } };
}

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
    case "connected":
      return { ...state, connecting: false, view: event.view, announcement: connectedNotice(event.view) };
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
  };
}

/** True while Backlog is off for the workspace. */
export function isOff(view: ConnectionView | undefined): boolean {
  return view?.enabled === false;
}

function classify(error: unknown): Pick<ScreenState, "fieldError" | "notice" | "isMember"> {
  const f = readFailure(error);
  if (f.status === 403) return { isMember: true };
  switch (f.code) {
    case "validation":
      if (f.field === "spaceUrl")
        return { isMember: false, fieldError: { field: "spaceUrl", key: "errorSpaceUrl" } };
      if (f.field === "apiKey")
        return { isMember: false, fieldError: { field: "apiKey", key: "errorApiKey" } };
      return { isMember: false, notice: { key: "errorInput" } };
    case "unreachable":
      return { isMember: false, notice: { key: "unreachable", retry: true } };
    case "rate_limited":
      return {
        isMember: false,
        notice: { key: "rateLimited", params: { seconds: f.retryAfterSeconds ?? 60 } },
      };
    case "conflict":
      return { isMember: false, notice: { key: "busy" } };
    default:
      return { isMember: false, notice: { key: "connectFailed" } };
  }
}
