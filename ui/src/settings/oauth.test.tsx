import { act } from "react";
import axe from "axe-core";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { createSettingsScreen } from "./SettingsScreen";
import { browser } from "./oauth";
import { en } from "../messages/en";
import {
  byTestId,
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

const AUTHORIZE = `https://${HOST}/OAuth2AccessRequest.action?response_type=code&state=s`;

let assign: ReturnType<typeof vi.spyOn>;
let replace: ReturnType<typeof vi.spyOn>;
beforeEach(() => {
  assign = vi.spyOn(browser, "assign").mockImplementation(() => undefined);
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

async function chooseOAuth(c: HTMLElement, spaceUrl = HOST) {
  await act(async () => (byTestId(c, "backlog-method-oauth") as HTMLInputElement).click());
  await act(async () => setValue(byTestId(c, "backlog-space-url") as HTMLInputElement, spaceUrl));
}

describe("OAuth sign-in (US1.3)", () => {
  it("offers a sign-in method radiogroup with API key as the default", async () => {
    const { c } = await render();
    const group = c.querySelector('[role="radiogroup"]')!;
    expect(group).not.toBeNull();
    expect(group.getAttribute("aria-labelledby")).toBeTruthy();
    expect((byTestId(c, "backlog-method-api-key") as HTMLInputElement).checked).toBe(true);
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
    await act(async () => byTestId(c, "backlog-sign-in-nulab")!.click());
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.start_oauth", {
      workspaceId: "ws-1",
      body: { spaceUrl: ` https://${HOST.toUpperCase()}/ ` },
    });
    expect(assign).toHaveBeenCalledWith(AUTHORIZE);
  });

  it("refuses an authorize URL for another host", async () => {
    const { c } = await render(
      scripted(notConnected, async () => ({ authorizeUrl: "https://evil.example.test/x" })),
    );
    await chooseOAuth(c);
    await act(async () => byTestId(c, "backlog-sign-in-nulab")!.click());
    expect(assign).not.toHaveBeenCalled();
    expect(text(c)).toContain(en.connectFailed);
  });

  it("shows Connecting... and locks the button while waiting", async () => {
    const pending = deferred<unknown>();
    const { c, host } = await render(scripted(notConnected, () => pending.promise));
    await chooseOAuth(c);
    const button = byTestId(c, "backlog-sign-in-nulab") as HTMLButtonElement;
    await act(async () => button.click());
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
    await act(async () => byTestId(c, "backlog-sign-in-again")!.click());
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.start_oauth", {
      workspaceId: "ws-1",
      body: { spaceUrl: HOST },
    });
    expect(assign).toHaveBeenCalledWith(AUTHORIZE);
  });
});
