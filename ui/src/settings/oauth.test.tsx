import { act } from "react";
import axe from "axe-core";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { createSettingsScreen } from "./SettingsScreen";
import { browser, startOAuth } from "./oauth";
import { en } from "../messages/en";
import {
  byTestId,
  choose,
  deferred,
  fakeHost,
  HOST,
  mount,
  notConnected,
  setValue,
  signInAgain,
  text,
  unmount,
  type View,
} from "../testing/harness";

const anyHash = expect.stringMatching(/^[0-9a-f]{64}$/);
const COOKIE =
  /^nulab_backlog_oauth_verifier=([A-Za-z0-9_-]{43}); Path=\/api\/plugins\/nulab-backlog\/webhooks\/oauth-callback; Secure; SameSite=Lax; Max-Age=600$/;

async function sha256Hex(text: string): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text));
  return [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, "0")).join("");
}

const AUTHORIZE = `https://${HOST}/OAuth2AccessRequest.action?response_type=code&state=s`;

let assign: ReturnType<typeof vi.spyOn>;
let replace: ReturnType<typeof vi.spyOn>;
let setCookie: ReturnType<typeof vi.spyOn>;
beforeEach(() => {
  assign = vi.spyOn(browser, "assign").mockImplementation(() => undefined);
  setCookie = vi.spyOn(browser, "setCookie").mockImplementation(() => undefined);
  replace = vi.spyOn(browser, "replaceSearch").mockImplementation(() => undefined);
});
afterEach(() => {
  unmount();
  vi.restoreAllMocks();
});

function scripted(view: View, start: () => Promise<unknown> = async () => ({ authorizeUrl: AUTHORIZE })) {
  return fakeHost(async (key) => (key === "connection.get" ? view : start()));
}

async function render(host = scripted(notConnected)) {
  const c = await mount(createSettingsScreen(host), { workspaceId: "ws-1" });
  return { c, host };
}

/**
 * Clicks a sign-in button and waits for connection.start_oauth: hashing the
 * verifier with Web Crypto finishes outside act's microtask flush.
 */
async function clickStart(c: HTMLElement, host: ReturnType<typeof fakeHost>, testId: string) {
  await act(async () => byTestId(c, testId)!.click());
  await act(() =>
    vi.waitFor(() =>
      expect(vi.mocked(host.api.invokeAction).mock.calls.some(([k]) => k === "connection.start_oauth")).toBe(
        true,
      ),
    ),
  );
  await act(async () => undefined); // let the reply settle
}

async function chooseOAuth(c: HTMLElement, spaceUrl = HOST) {
  await act(async () => choose(c, "backlog-method", "oauth"));
  await act(async () => setValue(byTestId(c, "backlog-space-url") as HTMLInputElement, spaceUrl));
}

describe("OAuth sign-in (US1.3)", () => {
  it("offers a sign-in method dropdown with API key as the default (BR1.4)", async () => {
    const { c } = await render();
    expect(c.querySelector('label[for="backlog-method"]')!.textContent).toBe(en.signInMethodLabel);
    expect(byTestId(c, "backlog-method-api-key")!.getAttribute("aria-selected")).toBe("true");
    expect(byTestId(c, "backlog-api-key")).not.toBeNull();
    await chooseOAuth(c);
    expect(byTestId(c, "backlog-api-key")).toBeNull();
    expect(byTestId(c, "backlog-sign-in-nulab")!.textContent).toBe(en.signInWithNulab);
    const result = await axe.run(c, { rules: { "color-contrast": { enabled: false } } });
    expect(result.violations.map((v) => v.id)).toEqual([]);
  });

  it("starts OAuth and navigates to Backlog's authorization page", async () => {
    const { c, host } = await render();
    await chooseOAuth(c, ` https://${HOST.toUpperCase()}/ `);
    await clickStart(c, host, "backlog-sign-in-nulab");
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.start_oauth", {
      workspaceId: "ws-1",
      body: { spaceUrl: ` https://${HOST.toUpperCase()}/ `, verifierHash: anyHash },
    });
    expect(assign).toHaveBeenCalledWith(AUTHORIZE);
  });

  it("binds the sign-in to this browser with a verifier cookie (R-01)", async () => {
    const random = vi.spyOn(crypto, "getRandomValues");
    const { c, host } = await render();
    await chooseOAuth(c);
    await clickStart(c, host, "backlog-sign-in-nulab");
    expect((random.mock.calls[0]![0] as Uint8Array).length).toBe(32);
    expect(setCookie).toHaveBeenCalledTimes(1);
    const verifier = COOKIE.exec(String(setCookie.mock.calls[0]![0]))?.[1];
    expect(verifier).toBeDefined();
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.start_oauth", {
      workspaceId: "ws-1",
      body: { spaceUrl: HOST, verifierHash: await sha256Hex(verifier!) },
    });
    expect(JSON.stringify(vi.mocked(host.api.invokeAction).mock.calls)).not.toContain(verifier);
    expect(setCookie.mock.invocationCallOrder[0]).toBeLessThan(assign.mock.invocationCallOrder[0]!);
  });

  it("uses a new verifier for every sign-in", async () => {
    const host = scripted(notConnected);
    await startOAuth(host, "ws-1", HOST);
    await startOAuth(host, "ws-1", HOST);
    const [first, second] = setCookie.mock.calls.map((call: unknown[]) => COOKIE.exec(String(call[0]))?.[1]);
    expect(first).toBeDefined();
    expect(second).toBeDefined();
    expect(first).not.toBe(second);
  });

  it("refuses an authorize URL for another host", async () => {
    const { c, host } = await render(
      scripted(notConnected, async () => ({ authorizeUrl: "https://evil.example.test/x" })),
    );
    await chooseOAuth(c);
    await clickStart(c, host, "backlog-sign-in-nulab");
    expect(assign).not.toHaveBeenCalled();
    expect(setCookie).not.toHaveBeenCalled();
    expect(text(c)).toContain(en.connectFailed);
  });

  it("shows Connecting... and locks the button while waiting", async () => {
    const pending = deferred<unknown>();
    const { c, host } = await render(scripted(notConnected, () => pending.promise));
    await chooseOAuth(c);
    const button = byTestId(c, "backlog-sign-in-nulab") as HTMLButtonElement;
    await clickStart(c, host, "backlog-sign-in-nulab");
    expect(button.textContent).toBe(en.connecting);
    expect(button.disabled).toBe(true);
    await act(async () => button.click());
    expect(
      vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "connection.start_oauth"),
    ).toHaveLength(1);
    await act(async () => pending.resolve({ authorizeUrl: AUTHORIZE }));
  });

  it.each([
    ["cancelled", en.oauthCancelled],
    ["failed", en.oauthFailed],
  ])("shows and announces ?oauth=%s exactly once", async (outcome, message) => {
    vi.spyOn(browser, "search").mockReturnValue(`?oauth=${outcome}`);
    const { c } = await render();
    expect(byTestId(c, "backlog-notice")!.textContent).toContain(message);
    expect(byTestId(c, "backlog-announcement")!.textContent).toBe(message);
    expect(replace).toHaveBeenCalledTimes(1);
    expect(text(c).split(message)).toHaveLength(3); // once in the notice, once in the live region
  });

  it("offers Sign in again with the stored host", async () => {
    const { c, host } = await render(scripted(signInAgain));
    expect(byTestId(c, "backlog-sign-in-again-notice")!.textContent).toBe(en.signInAgainNotice);
    await clickStart(c, host, "backlog-sign-in-again");
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.start_oauth", {
      workspaceId: "ws-1",
      body: { spaceUrl: HOST, verifierHash: anyHash },
    });
    expect(assign).toHaveBeenCalledWith(AUTHORIZE);
  });
});
