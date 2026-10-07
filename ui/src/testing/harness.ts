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

/** U3: the host's Skeleton, a placeholder block. */
export function FakeSkeleton(props: Record<string, unknown>) {
  return React.createElement("div", { "data-testid": "fake-skeleton", ...props });
}

type Props = Record<string, unknown> & { children?: React.ReactNode };

/**
 * Intent 261007: the host UI kit as plain DOM. Every stub element carries
 * data-host="<Component>", and Button passes variant and size through as
 * data-variant / data-size, so tests can tell host controls from raw ones
 * and check the GitHub button style (BR5.1, BR5.2).
 */
function tag(name: string, element: string) {
  const Stub = ({ children, ...props }: Props) =>
    React.createElement(element, { "data-host": name, ...props }, children as React.ReactNode);
  Stub.displayName = name;
  return Stub;
}

function Button({ variant, size, children, ...props }: Props) {
  delete props.asChild;
  return React.createElement(
    "button",
    { "data-host": "Button", "data-variant": variant ?? "default", "data-size": size ?? "default", ...props },
    children as React.ReactNode,
  );
}

function Input(props: Props) {
  return React.createElement("input", { "data-host": "Input", ...props });
}

function Checkbox({ checked, onCheckedChange, disabled, ...props }: Props) {
  const on = checked === true;
  return React.createElement("button", {
    type: "button",
    role: "checkbox",
    "data-host": "Checkbox",
    "aria-checked": on,
    disabled,
    onClick: () => (onCheckedChange as (v: boolean) => void)?.(!on),
    ...props,
  });
}

interface Choice {
  value?: string;
  set?: (v: string) => void;
  disabled?: boolean;
}
const SelectCtx = React.createContext<Choice>({});

function Select({ value, onValueChange, disabled, children }: Props) {
  return React.createElement(
    SelectCtx.Provider,
    {
      value: {
        value: value as string,
        set: onValueChange as (v: string) => void,
        disabled: disabled as boolean,
      },
    },
    React.createElement("div", { "data-host": "Select", "data-value": value }, children as React.ReactNode),
  );
}

function SelectTrigger({ children, ...props }: Props) {
  const ctx = React.useContext(SelectCtx);
  return React.createElement(
    "button",
    {
      type: "button",
      role: "combobox",
      "aria-expanded": false,
      "data-host": "SelectTrigger",
      disabled: ctx.disabled,
      ...props,
    },
    children as React.ReactNode,
  );
}

function SelectValue({ placeholder }: Props) {
  const ctx = React.useContext(SelectCtx);
  return React.createElement(
    "span",
    { "data-host": "SelectValue" },
    ctx.value ? "" : ((placeholder as string) ?? ""),
  );
}

function SelectItem({ value, disabled, children, ...props }: Props) {
  const ctx = React.useContext(SelectCtx);
  return React.createElement(
    "div",
    {
      role: "option",
      "data-host": "SelectItem",
      "data-value": value,
      "aria-selected": ctx.value === value,
      "aria-disabled": disabled ? true : undefined,
      onClick: () => !disabled && ctx.set?.(value as string),
      ...props,
    },
    children as React.ReactNode,
  );
}

interface Open {
  onOpenChange?: (open: boolean) => void;
  value?: string;
  set?: (v: string) => void;
}
const DialogCtx = React.createContext<Open>({});
const TabsCtx = React.createContext<Open>({});

function Dialog({ open, onOpenChange, children }: Props) {
  if (!open) return null;
  return React.createElement(
    DialogCtx.Provider,
    { value: { onOpenChange: onOpenChange as (o: boolean) => void } },
    children as React.ReactNode,
  );
}

function DialogContent({ children, ...props }: Props) {
  const ctx = React.useContext(DialogCtx);
  return React.createElement(
    "div",
    {
      role: "dialog",
      "aria-modal": true,
      "data-host": "DialogContent",
      onKeyDown: (e: KeyboardEvent) => e.key === "Escape" && ctx.onOpenChange?.(false),
      ...props,
    },
    children as React.ReactNode,
  );
}

function Tabs({ value, onValueChange, children, ...props }: Props) {
  return React.createElement(
    TabsCtx.Provider,
    { value: { value: value as string, set: onValueChange as (v: string) => void } },
    React.createElement("div", { "data-host": "Tabs", ...props }, children as React.ReactNode),
  );
}

function TabsTrigger({ value, children, ...props }: Props) {
  const ctx = React.useContext(TabsCtx);
  return React.createElement(
    "button",
    {
      type: "button",
      role: "tab",
      "data-host": "TabsTrigger",
      "aria-selected": ctx.value === value,
      onClick: () => ctx.set?.(value as string),
      ...props,
    },
    children as React.ReactNode,
  );
}

function TabsContent({ value, forceMount, children, ...props }: Props) {
  const ctx = React.useContext(TabsCtx);
  if (ctx.value !== value && !forceMount) return null;
  return React.createElement(
    "div",
    { role: "tabpanel", "data-host": "TabsContent", ...props },
    children as React.ReactNode,
  );
}

/** The menu is always open in tests; an item runs onSelect on click. */
function DropdownMenuItem({ onSelect, disabled, children, ...props }: Props) {
  return React.createElement(
    "div",
    {
      role: "menuitem",
      tabIndex: -1,
      "data-host": "DropdownMenuItem",
      "aria-disabled": disabled ? true : undefined,
      onClick: disabled ? undefined : onSelect,
      ...props,
    },
    children as React.ReactNode,
  );
}

function DropdownMenuTrigger({ children }: Props) {
  return children as React.ReactElement; // asChild: the plugin's Button
}

function SettingsSection({ title, description, action, icon, children }: Props) {
  return React.createElement(
    "section",
    { "data-host": "SettingsSection" },
    React.createElement("h3", null, icon as React.ReactNode, title as string),
    description ? React.createElement("p", null, description as string) : null,
    action as React.ReactNode,
    children as React.ReactNode,
  );
}

function IntegrationRepositoryFilter({ value, onValueChange, options, ariaLabel, testId }: Props) {
  const list = (options as { value: string; label: string }[]) ?? [];
  return React.createElement(
    "div",
    {
      role: "group",
      "aria-label": ariaLabel,
      "data-host": "IntegrationRepositoryFilter",
      "data-testid": testId,
      "data-value": value,
    },
    list.map((o) =>
      React.createElement(
        "button",
        {
          key: o.value,
          type: "button",
          "data-host": "IntegrationRepositoryFilterOption",
          "data-testid": `${testId}-option-${o.value}`,
          "aria-pressed": o.value === value,
          onClick: () => (onValueChange as (v: string) => void)(o.value),
        },
        o.label,
      ),
    ),
  );
}

function ChangeRequestList({ loading, error, emptyMessage, isEmpty, children }: Props) {
  let content: React.ReactNode = children as React.ReactNode;
  if (loading) content = React.createElement("p", null, "…");
  else if (error) content = React.createElement("p", null, error as string);
  else if (isEmpty) content = React.createElement("p", null, emptyMessage as string);
  return React.createElement("div", { "data-host": "ChangeRequestList" }, content);
}

function ChangeRequestRow({ stateIcon, title, href, metadata, taskIndicator, testId }: Props) {
  return React.createElement(
    "div",
    { "data-host": "ChangeRequestRow", "data-testid": testId },
    stateIcon as React.ReactNode,
    React.createElement(
      "a",
      { href, target: "_blank", rel: "noopener noreferrer", "data-testid": `${testId}-link` },
      title as string,
    ),
    React.createElement("div", null, metadata as React.ReactNode, taskIndicator as React.ReactNode),
  );
}

function IntegrationIcon({ name, className }: Props) {
  return React.createElement("svg", { "data-integration-icon": name, className, "aria-hidden": true });
}

export const fakeUi = {
  Button,
  Input,
  Label: "label",
  Checkbox,
  Select,
  SelectTrigger,
  SelectValue,
  // Closed, like the host's popover: hidden from axe, still clickable in tests.
  SelectContent: (props: Props) =>
    React.createElement("div", { role: "listbox", hidden: true, "data-host": "SelectContent", ...props }),
  SelectItem,
  Dialog,
  DialogContent,
  DialogHeader: tag("DialogHeader", "div"),
  DialogFooter: tag("DialogFooter", "div"),
  DialogTitle: tag("DialogTitle", "h2"),
  DialogDescription: tag("DialogDescription", "p"),
  Table: tag("Table", "table"),
  TableHeader: tag("TableHeader", "thead"),
  TableBody: tag("TableBody", "tbody"),
  TableRow: tag("TableRow", "tr"),
  TableHead: tag("TableHead", "th"),
  TableCell: tag("TableCell", "td"),
  Card: tag("Card", "div"),
  CardContent: tag("CardContent", "div"),
  Badge: tag("Badge", "span"),
  Alert: tag("Alert", "div"),
  AlertTitle: tag("AlertTitle", "div"),
  AlertDescription: tag("AlertDescription", "div"),
  Empty: tag("Empty", "div"),
  EmptyHeader: tag("EmptyHeader", "div"),
  EmptyTitle: tag("EmptyTitle", "div"),
  EmptyDescription: tag("EmptyDescription", "p"),
  EmptyContent: tag("EmptyContent", "div"),
  Pagination: tag("Pagination", "nav"),
  PaginationContent: tag("PaginationContent", "ul"),
  PaginationItem: tag("PaginationItem", "li"),
  Tabs,
  TabsList: (props: Props) =>
    React.createElement("div", { role: "tablist", "data-host": "TabsList", ...props }),
  TabsTrigger,
  TabsContent,
  DropdownMenu: tag("DropdownMenu", "div"),
  DropdownMenuTrigger,
  DropdownMenuContent: (props: Props) =>
    React.createElement("div", { role: "menu", "data-host": "DropdownMenuContent", ...props }),
  DropdownMenuItem,
  SettingsSection,
  IntegrationRepositoryFilter,
  ChangeRequestList,
  ChangeRequestRow,
  IntegrationIcon,
};

/** Picks value in the host Select whose trigger has data-testid triggerId. */
export function choose(c: HTMLElement, triggerId: string, value: string) {
  const select = byTestId(c, triggerId)!.closest('[data-host="Select"]')!;
  const item = select.querySelector<HTMLElement>(`[role="option"][data-value="${value}"]`);
  if (!item) throw new Error(`no option ${value} in ${triggerId}`);
  item.click();
}

/** The options of the host Select whose trigger has data-testid triggerId. */
export function optionValues(c: HTMLElement, triggerId: string): string[] {
  const select = byTestId(c, triggerId)!.closest('[data-host="Select"]')!;
  return [...select.querySelectorAll('[role="option"]')].map((o) => o.getAttribute("data-value") ?? "");
}

/** Raw controls a plugin component must not render (BR5.1): any without data-host. */
export function rawControls(c: HTMLElement): string[] {
  return [...c.querySelectorAll("button, select, table, details, input[type=checkbox], input[type=radio]")]
    .filter((el) => !el.hasAttribute("data-host"))
    .map((el) => el.outerHTML.slice(0, 80));
}

/** U3: host.i18n with an optional catalogue of the active locale. */
export function fakeI18n(locale = "en", catalogue: Record<string, string> = {}) {
  const t = (key: string, options?: { defaultValue?: string }) =>
    catalogue[key] ?? options?.defaultValue ?? key;
  return { locale, t, useTranslation: () => ({ locale, t }) };
}

/** U3: a registry that records every registration. */
export function fakeRegistry() {
  return {
    registerTranslations: vi.fn(),
    registerIntegrationSettings: vi.fn(),
    registerNavItem: vi.fn(),
    registerRoute: vi.fn(),
    registerRepositoryProvider: vi.fn(),
    registerTaskAction: vi.fn(),
    registerReviewProvider: vi.fn(),
    registerComponent: vi.fn(),
    registerTaskMenuAction: vi.fn(),
    registerTaskPanel: vi.fn(),
  };
}

/** A fake host: React from the test, plain elements for the UI kit, scripted actions. */
export function fakeHost(invoke: Invoke, overrides: Record<string, unknown> = {}): PluginHostApi {
  return {
    pluginId: "nulab-backlog",
    React,
    jsx: React.createElement,
    ui: {
      ...fakeUi,
      ChangeRequestDetail: FakeChangeRequestDetail,
      Skeleton: FakeSkeleton, // U3
    },
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
    // U3: English unless a test gives a catalogue; the fallback is the default value.
    i18n: fakeI18n(),
    useResponsiveBreakpoint: () => ({ isMobile: false }),
    toast: Object.assign(vi.fn(), {
      success: vi.fn(),
      error: vi.fn(),
      info: vi.fn(),
      warning: vi.fn(),
      dismiss: vi.fn(),
    }),
    utils: { formatRelativeTime: (v: string) => `rel:${v}` },
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
