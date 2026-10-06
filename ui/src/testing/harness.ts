// Shared helpers for the UI tests. Not part of the bundle.
import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { vi } from "vitest";
import type { PluginHostApi } from "@kandev/plugin-sdk";

declare global {
  var IS_REACT_ACT_ENVIRONMENT: boolean;
}
globalThis.IS_REACT_ACT_ENVIRONMENT = true;

export const HOST = "example-space.backlog.com";

export interface View {
  connected: boolean;
  enabled?: boolean;
  state: string;
  spaceHost?: string;
  connectedUserName?: string;
  hasApiKey: boolean;
  authMethod?: string;
  selectedProjects?: string[];
  restored?: boolean;
  hasGitCredential?: boolean;
  gitCheck?: string;
}

export const notConnected: View = {
  connected: false,
  enabled: true,
  state: "not_connected",
  hasApiKey: false,
};
export const connected: View = {
  connected: true,
  enabled: true,
  state: "connected",
  spaceHost: HOST,
  connectedUserName: "Test User",
  hasApiKey: true,
};
export const incomplete: View = {
  connected: false,
  enabled: true,
  state: "error",
  spaceHost: HOST,
  hasApiKey: false,
};

/** An error shaped like the host's ApiError for a non-2xx action response. */
export function actionError(status: number, error: Record<string, unknown>): Error {
  return Object.assign(new Error(`Request failed: ${status}`), { status, body: { error } });
}

export interface Deferred<T> {
  promise: Promise<T>;
  resolve(value: T): void;
  reject(reason: unknown): void;
}

export function deferred<T>(): Deferred<T> {
  let resolve!: (v: T) => void;
  let reject!: (e: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

export interface ActionInputLike {
  workspaceId?: string;
  taskId?: string;
  sessionId?: string;
  repositoryId?: string;
  body?: unknown;
}

export type Invoke = (key: string, input?: ActionInputLike) => Promise<unknown>;

/** U4: the fake host's ChangeRequestDetail shows the detail it was given. */
export function FakeChangeRequestDetail(props: {
  detail: { title: string; state: string; sourceBranch: string; targetBranch: string } | null;
  error?: string | null;
}) {
  return React.createElement(
    "div",
    { "data-testid": "fake-change-request-detail" },
    props.detail
      ? `${props.detail.title}|${props.detail.state}|${props.detail.sourceBranch}|${props.detail.targetBranch}`
      : (props.error ?? ""),
  );
}

/** A fake host: React from the test, plain elements for the UI kit, scripted actions. */
export function fakeHost(invoke: Invoke, overrides: Record<string, unknown> = {}): PluginHostApi {
  return {
    pluginId: "nulab-backlog",
    React,
    jsx: React.createElement,
    ui: { Button: "button", Input: "input", Label: "label", ChangeRequestDetail: FakeChangeRequestDetail },
    api: { baseUrl: "", fetch: vi.fn(), invokeAction: vi.fn(invoke) },
    context: {
      getActiveWorkspaceId: () => "ws-1",
      subscribeActiveWorkspace: () => () => undefined,
      // U4: the workflow a watch's tasks go to (M4).
      getTaskCreationContext: (workspaceId: string) => ({
        workspaceId,
        workflowId: "wf-1",
        defaultStepId: "step-1",
        steps: [],
        repositories: [],
      }),
    },
    setIntegrationEnabled: vi.fn(),
    navigate: vi.fn(),
    // U4: Kandev-owned dialogs; tests read the options they were opened with.
    openTaskLinkDialog: vi.fn(() => ({ close: vi.fn() })),
    openModal: vi.fn(() => ({ close: vi.fn() })),
    ...overrides,
  } as unknown as PluginHostApi;
}

/** Runs axe on c and returns the violation ids (color contrast off: jsdom has no layout). */
export async function axeViolations(c: HTMLElement): Promise<string[]> {
  const { default: axe } = await import("axe-core");
  const result = await axe.run(c, { rules: { "color-contrast": { enabled: false } } });
  return result.violations.map((v) => v.id);
}

/** Every interactive element under c has a data-testid. */
export function expectTestIds(c: HTMLElement, expectFn: (ok: boolean, msg: string) => void) {
  c.querySelectorAll("button, input, select, textarea, a, summary").forEach((el) =>
    expectFn(Boolean(el.getAttribute("data-testid")), `no data-testid on ${el.outerHTML.slice(0, 80)}`),
  );
}

/** A catalogue whose every value is wrapped in ⟦⟧, for expectOnlyCatalogueText. */
export function pseudoCatalogue<T extends Record<string, string>>(messages: T): T {
  return Object.fromEntries(Object.entries(messages).map(([k, v]) => [k, `⟦${v}⟧`])) as T;
}

/** Selects an option the way React's onChange sees it. */
export function selectValue(select: HTMLSelectElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, "value")!.set!;
  setter.call(select, value);
  select.dispatchEvent(new Event("change", { bubbles: true }));
}

let root: Root | undefined;
let container: HTMLDivElement | undefined;

/** Renders a component into a fresh container. */
export async function mount<P extends object>(component: unknown, props: P): Promise<HTMLDivElement> {
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
  await act(async () => {
    root!.render(React.createElement(component as React.FC<P>, props));
  });
  return container;
}

export function unmount() {
  act(() => root?.unmount());
  container?.remove();
  root = undefined;
}

export const byTestId = (c: HTMLElement, id: string) => c.querySelector<HTMLElement>(`[data-testid="${id}"]`);
export const text = (c: HTMLElement) => c.textContent ?? "";

/** Fails on any visible text node that is not wrapped in ⟦⟧ by the pseudo catalogue. */
export function expectOnlyCatalogueText(c: HTMLElement, expectFn: (ok: boolean, msg: string) => void) {
  const walker = document.createTreeWalker(c, NodeFilter.SHOW_TEXT);
  for (let n = walker.nextNode(); n; n = walker.nextNode()) {
    const t = n.textContent ?? "";
    if (t.trim() === "") continue;
    expectFn(t.startsWith("⟦") && t.endsWith("⟧"), `literal text: ${t}`);
  }
}

/** Sets an input's value the way React's onChange sees it. */
export function setValue(input: HTMLInputElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")!.set!;
  setter.call(input, value);
  input.dispatchEvent(new Event("input", { bubbles: true }));
}

/** Dispatches a keydown on el. */
export function press(el: HTMLElement, key: string, shiftKey = false) {
  el.dispatchEvent(new KeyboardEvent("keydown", { key, shiftKey, bubbles: true, cancelable: true }));
}

/** A view that needs a new OAuth sign-in (AC1.4.3). */
export const signInAgain: View = {
  connected: false,
  enabled: true,
  state: "sign_in_again",
  spaceHost: HOST,
  connectedUserName: "Test User",
  hasApiKey: false,
  authMethod: "oauth",
};
