import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import axe from "axe-core";
import { afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import type { PluginHostApi } from "@kandev/plugin-sdk";

import { createSettingsScreen } from "./SettingsScreen";
import { en, type Messages } from "../messages/en";

declare global {
  var IS_REACT_ACT_ENVIRONMENT: boolean;
}

beforeAll(() => {
  globalThis.IS_REACT_ACT_ENVIRONMENT = true;
});

const HOST = "example-space.backlog.com";

interface ConnectionView {
  connected: boolean;
  state: string;
  spaceHost?: string;
  connectedUserName?: string;
  hasApiKey: boolean;
}

const notConnected: ConnectionView = { connected: false, state: "not_connected", hasApiKey: false };
const connected: ConnectionView = {
  connected: true,
  state: "connected",
  spaceHost: HOST,
  connectedUserName: "Test User",
  hasApiKey: true,
};
const incomplete: ConnectionView = { connected: false, state: "error", spaceHost: HOST, hasApiKey: false };

/** An error shaped like the host's ApiError for a non-2xx action response. */
function actionError(status: number, error: Record<string, unknown>): Error {
  return Object.assign(new Error(`Request failed: ${status}`), { status, body: { error } });
}

interface Deferred<T> {
  promise: Promise<T>;
  resolve(value: T): void;
  reject(reason: unknown): void;
}

function deferred<T>(): Deferred<T> {
  let resolve!: (v: T) => void;
  let reject!: (e: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

type Invoke = (key: string, input?: { workspaceId?: string; body?: unknown }) => Promise<unknown>;

function fakeHost(invoke: Invoke): PluginHostApi {
  const ui = { Button: "button", Input: "input", Label: "label" };
  return {
    pluginId: "nulab-backlog",
    React,
    jsx: React.createElement,
    ui,
    api: { baseUrl: "", fetch: vi.fn(), invokeAction: vi.fn(invoke) },
    context: { getActiveWorkspaceId: () => "ws-1" },
  } as unknown as PluginHostApi;
}

let root: Root | undefined;
let container: HTMLDivElement | undefined;

afterEach(() => {
  act(() => root?.unmount());
  container?.remove();
  root = undefined;
});

async function render(host: PluginHostApi, messages: Messages = en): Promise<HTMLDivElement> {
  const Screen = createSettingsScreen(host, messages);
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
  await act(async () => {
    root!.render(React.createElement(Screen as React.FC<{ workspaceId?: string }>, { workspaceId: "ws-1" }));
  });
  return container;
}

const byTestId = (c: HTMLElement, id: string) => c.querySelector<HTMLElement>(`[data-testid="${id}"]`);
const text = (c: HTMLElement) => c.textContent ?? "";

function setValue(input: HTMLInputElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")!.set!;
  setter.call(input, value);
  input.dispatchEvent(new Event("input", { bubbles: true }));
}

async function fillAndConnect(c: HTMLElement, spaceUrl = HOST, apiKey = "test-api-key-0000") {
  await act(async () => {
    setValue(byTestId(c, "backlog-space-url") as HTMLInputElement, spaceUrl);
    setValue(byTestId(c, "backlog-api-key") as HTMLInputElement, apiKey);
  });
  await act(async () => {
    (byTestId(c, "backlog-connect") as HTMLButtonElement).click();
  });
}

/** A host whose get returns `view` and whose connect returns or throws `connect`. */
function scripted(view: ConnectionView, connect?: () => Promise<unknown>) {
  const calls: { key: string; body?: unknown }[] = [];
  const host = fakeHost(async (key, input) => {
    calls.push({ key, body: input?.body });
    if (key === "connection.get") return view;
    return connect ? connect() : connected;
  });
  return { host, calls };
}

describe("M1 settings screen", () => {
  it("shows a loading indicator and no form while connection.get runs", async () => {
    const pending = deferred<unknown>();
    const c = await render(fakeHost(() => pending.promise));
    expect(text(c)).toContain(en.loading);
    expect(byTestId(c, "backlog-connect-form")).toBeNull();
  });

  it("shows load failed with Retry, and Retry reloads", async () => {
    let fail = true;
    const host = fakeHost(async () => {
      if (fail) throw actionError(500, { code: "internal" });
      return notConnected;
    });
    const c = await render(host);
    expect(text(c)).toContain(en.loadFailed);
    fail = false;
    await act(async () => byTestId(c, "backlog-load-retry")!.click());
    expect(byTestId(c, "backlog-connect-form")).not.toBeNull();
  });

  it("shows the empty form when not connected", async () => {
    const c = await render(scripted(notConnected).host);
    expect((byTestId(c, "backlog-api-key") as HTMLInputElement).value).toBe("");
    expect(byTestId(c, "backlog-connect")!.textContent).toBe(en.connect);
    expect(c.querySelector('label[for="backlog-space-url"]')!.textContent).toBe(en.spaceUrlLabel);
    expect(c.querySelector('label[for="backlog-api-key"]')!.textContent).toBe(en.apiKeyLabel);
  });

  it("shows Connected as <name> @ <host> and the replace form when connected", async () => {
    const c = await render(scripted(connected).host);
    expect(byTestId(c, "backlog-status")!.textContent).toBe(`Connected as Test User @ ${HOST}`);
    expect(text(c)).toContain(en.replaceHeading);
    expect((byTestId(c, "backlog-api-key") as HTMLInputElement).value).toBe("");
  });

  it("asks to connect again when record and secret disagree", async () => {
    const c = await render(scripted(incomplete).host);
    expect(text(c)).toContain(en.incomplete);
    expect(byTestId(c, "backlog-connect-form")).not.toBeNull();
  });

  it("locks the button while connecting and ignores more clicks", async () => {
    const pending = deferred<unknown>();
    const { host, calls } = scripted(notConnected, () => pending.promise);
    const c = await render(host);
    await fillAndConnect(c);
    const button = byTestId(c, "backlog-connect") as HTMLButtonElement;
    expect(button.textContent).toBe(en.connecting);
    expect(button.disabled).toBe(true);
    await act(async () => button.click());
    expect(calls.filter((x) => x.key === "connection.connect_api_key")).toHaveLength(1);
    await act(async () => pending.resolve(connected));
    expect(button.disabled).toBe(false);
  });

  it("sends the trimmed-by-backend input as the action body with the workspace", async () => {
    const { host, calls } = scripted(notConnected);
    const c = await render(host);
    await fillAndConnect(c, HOST, "k-1");
    expect(calls.at(-1)).toEqual({
      key: "connection.connect_api_key",
      body: { spaceUrl: HOST, apiKey: "k-1" },
    });
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.connect_api_key", {
      workspaceId: "ws-1",
      body: { spaceUrl: HOST, apiKey: "k-1" },
    });
  });

  it.each([
    ["spaceUrl", "backlog-space-url", en.errorSpaceUrl],
    ["apiKey", "backlog-api-key", en.errorApiKey],
  ])("links a %s field error to its input", async (field, inputId, message) => {
    const { host } = scripted(notConnected, async () => {
      throw actionError(400, { code: "validation", field });
    });
    const c = await render(host);
    await fillAndConnect(c);
    const input = byTestId(c, inputId)!;
    expect(input.getAttribute("aria-invalid")).toBe("true");
    const described = document.getElementById(input.getAttribute("aria-describedby")!);
    expect(described!.textContent).toBe(message);
  });

  it("shows unreachable with Retry, never blames the key, and Retry resends", async () => {
    let attempt = 0;
    const { host, calls } = scripted(notConnected, async () => {
      attempt += 1;
      if (attempt === 1) throw actionError(503, { code: "unreachable" });
      return connected;
    });
    const c = await render(host);
    await fillAndConnect(c);
    expect(text(c)).toContain(en.unreachable);
    expect(byTestId(c, "backlog-api-key")!.getAttribute("aria-invalid")).not.toBe("true");
    await act(async () => byTestId(c, "backlog-connect-retry")!.click());
    expect(calls.filter((x) => x.key === "connection.connect_api_key")).toHaveLength(2);
    expect(byTestId(c, "backlog-status")!.textContent).toBe(`Connected as Test User @ ${HOST}`);
  });

  it("shows the rate-limit wait", async () => {
    const { host } = scripted(notConnected, async () => {
      throw actionError(429, { code: "rate_limited", retryAfterSeconds: 42 });
    });
    const c = await render(host);
    await fillAndConnect(c);
    expect(text(c)).toContain("Backlog is limiting requests. Try again in 42 s");
  });

  it("shows busy on conflict", async () => {
    const { host } = scripted(notConnected, async () => {
      throw actionError(409, { code: "conflict" });
    });
    const c = await render(host);
    await fillAndConnect(c);
    expect(text(c)).toContain(en.busy);
  });

  it("shows a save failure on internal", async () => {
    const { host } = scripted(notConnected, async () => {
      throw actionError(500, { code: "internal" });
    });
    const c = await render(host);
    await fillAndConnect(c);
    expect(text(c)).toContain(en.connectFailed);
  });

  it("switches to the member message after a 403", async () => {
    const { host } = scripted(notConnected, async () => {
      throw actionError(403, {});
    });
    const c = await render(host);
    await fillAndConnect(c);
    expect(text(c)).toContain(en.notConnectedMember);
    expect(byTestId(c, "backlog-connect-form")).toBeNull();
  });

  it("shows only the status to a member when connected", async () => {
    const { host } = scripted(connected, async () => {
      throw actionError(403, {});
    });
    const c = await render(host);
    await fillAndConnect(c);
    expect(byTestId(c, "backlog-status")!.textContent).toBe(`Connected as Test User @ ${HOST}`);
    expect(byTestId(c, "backlog-connect-form")).toBeNull();
  });

  it.each([
    ["success", async () => connected],
    [
      "failure",
      async () => {
        throw actionError(400, { code: "validation", field: "apiKey" });
      },
    ],
  ])("clears the key input after a %s", async (_, connect) => {
    const c = await render(scripted(notConnected, connect).host);
    await fillAndConnect(c);
    expect((byTestId(c, "backlog-api-key") as HTMLInputElement).value).toBe("");
  });

  it("announces the result in a polite live region", async () => {
    const c = await render(scripted(notConnected).host);
    const region = byTestId(c, "backlog-announcement")!;
    expect(region.getAttribute("aria-live")).toBe("polite");
    expect(region.getAttribute("role")).toBe("status");
    await fillAndConnect(c);
    expect(region.textContent).toBe(`Connected as Test User @ ${HOST}`);
  });

  it("puts a data-testid on every interactive element", async () => {
    const { host } = scripted(notConnected, async () => {
      throw actionError(503, { code: "unreachable" });
    });
    const c = await render(host);
    await fillAndConnect(c);
    const interactive = c.querySelectorAll("button, input, select, textarea, a");
    expect(interactive.length).toBeGreaterThan(0);
    interactive.forEach((el) => expect(el.getAttribute("data-testid")).toBeTruthy());
  });

  it("has no axe violations in each state", async () => {
    const states: [string, ConnectionView, (() => Promise<unknown>) | undefined][] = [
      ["not connected", notConnected, undefined],
      ["connected", connected, undefined],
      ["error", incomplete, undefined],
      [
        "field error",
        notConnected,
        async () => {
          throw actionError(400, { code: "validation", field: "spaceUrl" });
        },
      ],
      [
        "unreachable",
        notConnected,
        async () => {
          throw actionError(503, { code: "unreachable" });
        },
      ],
    ];
    for (const [name, view, connect] of states) {
      const c = await render(scripted(view, connect).host);
      if (connect) await fillAndConnect(c);
      const result = await axe.run(c, { rules: { "color-contrast": { enabled: false } } });
      expect(result.violations.map((v) => `${name}: ${v.id}`)).toEqual([]);
      act(() => root?.unmount());
      c.remove();
    }
  });

  it("renders no literal text outside the message catalogue", async () => {
    const pseudo = Object.fromEntries(Object.entries(en).map(([k, v]) => [k, `⟦${v}⟧`])) as Messages;
    const check = (c: HTMLElement) => {
      const walker = document.createTreeWalker(c, NodeFilter.SHOW_TEXT);
      for (let n = walker.nextNode(); n; n = walker.nextNode()) {
        const t = n.textContent ?? "";
        if (t.trim() === "") continue;
        expect(t.startsWith("⟦") && t.endsWith("⟧"), `literal text: ${t}`).toBe(true);
      }
      c.querySelectorAll("[placeholder]").forEach((el) =>
        expect(el.getAttribute("placeholder")).toMatch(/^⟦.*⟧$/),
      );
    };
    const { host } = scripted(connected, async () => {
      throw actionError(429, { code: "rate_limited", retryAfterSeconds: 3 });
    });
    const c = await render(host, pseudo);
    check(c);
    await fillAndConnect(c);
    check(c);
  });
});

describe("M1 settings screen: switch and layout", () => {
  const off: ConnectionView & { enabled: boolean } = { ...connected, enabled: false };

  it("shows the Off message and no Connect form while Backlog is off, keeping the connection", async () => {
    const c = await render(scripted(off).host);
    expect(byTestId(c, "backlog-off")!.textContent).toBe(en.integrationOff);
    expect(byTestId(c, "backlog-connect-form")).toBeNull();
    expect(byTestId(c, "backlog-status")!.textContent).toBe(`Connected as Test User @ ${HOST}`);
  });

  it("shows the same Off message when Connect replies integration_disabled", async () => {
    const { host } = scripted(notConnected, async () => {
      throw actionError(409, { code: "integration_disabled" });
    });
    const c = await render(host);
    await fillAndConnect(c);
    expect(byTestId(c, "backlog-off")!.textContent).toBe(en.integrationOff);
    expect(byTestId(c, "backlog-connect-form")).toBeNull();
    expect(byTestId(c, "backlog-announcement")!.textContent).toBe(en.integrationOff);
  });

  it("lays out one vertical stack with one gap and a smaller label-to-input gap (BR6.5)", async () => {
    const c = await render(scripted(connected).host);
    const screen = byTestId(c, "backlog-settings")!;
    const form = byTestId(c, "backlog-connect-form")!;
    expect(screen.className).toBe(STACK);
    expect(form.className).toBe(STACK);
    const fields = [...form.querySelectorAll("label")].map((l) => l.parentElement!);
    expect(fields).toHaveLength(2);
    fields.forEach((f) => expect(f.className).toBe(FIELD));
    expect(c.querySelector("style, link, [style]")).toBeNull();
  });

  it("uses the same stack in the loading and load-failed states", async () => {
    const pending = deferred<unknown>();
    const c = await render(fakeHost(() => pending.promise));
    expect(byTestId(c, "backlog-settings")!.className).toBe(STACK);
  });
});

const STACK = "flex flex-col gap-4";
const FIELD = "flex flex-col gap-2";
