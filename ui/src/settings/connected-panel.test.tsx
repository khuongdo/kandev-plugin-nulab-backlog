import { act } from "react";
import axe from "axe-core";
import { afterEach, describe, expect, it, vi } from "vitest";

import { createSettingsScreen } from "./SettingsScreen";
import { en } from "../messages/en";
import {
  actionError,
  byTestId,
  connected,
  deferred,
  fakeHost,
  HOST,
  mount,
  notConnected,
  press,
  setValue,
  text,
  unmount,
  type View,
} from "../testing/harness";

afterEach(unmount);

type Handler = (body: unknown) => Promise<unknown>;

function scripted(view: View, handlers: Record<string, Handler> = {}) {
  return fakeHost(async (key, input) => {
    if (key === "connection.get") return view;
    if (key === "connection.list_projects") return { projects: [] };
    if (key === "git.impact" && !handlers[key]) return { prLinks: 2, prWatches: 1 };
    const h = handlers[key];
    if (!h) throw new Error(`unexpected ${key}`);
    return h(input?.body);
  });
}

async function render(host: ReturnType<typeof scripted>) {
  return mount(createSettingsScreen(host), { workspaceId: "ws-1" });
}

const calls = (host: ReturnType<typeof scripted>, key: string) =>
  vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);

async function submitReplace(c: HTMLElement, spaceUrl: string) {
  await act(async () => {
    setValue(byTestId(c, "backlog-space-url") as HTMLInputElement, spaceUrl);
    setValue(byTestId(c, "backlog-api-key") as HTMLInputElement, "k-1");
  });
  await act(async () => byTestId(c, "backlog-connect")!.click());
}

describe("Connected panel (US1.5, US1.6, US1.8)", () => {
  it("tests the connection and announces the user name", async () => {
    const pending = deferred<unknown>();
    const c = await render(scripted(connected, { "connection.test": () => pending.promise }));
    const button = byTestId(c, "backlog-test-connection") as HTMLButtonElement;
    await act(async () => button.click());
    expect(button.textContent).toBe(en.testing);
    expect(button.disabled).toBe(true);
    await act(async () => pending.resolve({ ...connected, connectedUserName: "Fresh Name" }));
    const want = "The connection works. Signed in as Fresh Name.";
    expect(byTestId(c, "backlog-test-result")!.textContent).toBe(want);
    expect(byTestId(c, "backlog-announcement")!.textContent).toBe(want);
    expect(button.disabled).toBe(false);
  });

  it("shows the revoked-key message on reconnect_required", async () => {
    const c = await render(
      scripted(connected, {
        "connection.test": async () => {
          throw actionError(401, { code: "reconnect_required" });
        },
      }),
    );
    await act(async () => byTestId(c, "backlog-test-connection")!.click());
    expect(byTestId(c, "backlog-test-result")!.textContent).toBe(en.reconnectRequired);
  });

  it("asks before disconnecting, with Cancel focused, and Esc returns focus", async () => {
    const host = scripted(connected, { "connection.disconnect": async () => notConnected });
    const c = await render(host);
    const disconnect = byTestId(c, "backlog-disconnect")!;
    disconnect.focus();
    await act(async () => disconnect.click());
    const dialog = byTestId(c, "backlog-disconnect-dialog")!;
    expect(dialog.getAttribute("role")).toBe("alertdialog");
    expect(dialog.textContent).toContain("everyone in the workspace");
    expect(dialog.textContent).toContain("credentials are deleted");
    expect(document.activeElement).toBe(byTestId(c, "backlog-disconnect-dialog-cancel"));
    const result = await axe.run(c, { rules: { "color-contrast": { enabled: false } } });
    expect(result.violations.map((v) => v.id)).toEqual([]);
    await act(async () => press(document.activeElement as HTMLElement, "Escape"));
    expect(byTestId(c, "backlog-disconnect-dialog")).toBeNull();
    expect(document.activeElement).toBe(byTestId(c, "backlog-disconnect"));
    expect(calls(host, "connection.disconnect")).toHaveLength(0);
  });

  it("disconnects on Confirm and shows the not-connected form", async () => {
    const host = scripted(connected, { "connection.disconnect": async () => notConnected });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-disconnect")!.click());
    await act(async () => byTestId(c, "backlog-disconnect-dialog-confirm")!.click());
    expect(calls(host, "connection.disconnect")).toHaveLength(1);
    expect(byTestId(c, "backlog-status")).toBeNull();
    expect(byTestId(c, "backlog-connect")!.textContent).toBe(en.connect);
  });

  it("confirms a replacement on the same host, and Cancel calls nothing", async () => {
    const host = scripted(connected, { "connection.connect_api_key": async () => connected });
    const c = await render(host);
    expect(byTestId(c, "backlog-connect")!.textContent).toBe(en.replaceCredentials);
    await submitReplace(c, HOST.toUpperCase());
    expect(byTestId(c, "backlog-replace-dialog")!.textContent).toContain(en.replaceTitle);
    await act(async () => byTestId(c, "backlog-replace-dialog-cancel")!.click());
    expect(byTestId(c, "backlog-replace-dialog")).toBeNull();
    expect(calls(host, "connection.connect_api_key")).toHaveLength(0);
  });

  it("confirms a space change, says projects are cleared, then connects", async () => {
    const host = scripted(connected, {
      "connection.connect_api_key": async () => ({ ...connected, spaceHost: "other.backlog.jp" }),
    });
    const c = await render(host);
    await submitReplace(c, "other.backlog.jp");
    const dialog = byTestId(c, "backlog-change-space-dialog")!;
    expect(dialog.textContent).toContain("other.backlog.jp");
    expect(dialog.textContent).toContain("Selected projects are cleared");
    await act(async () => byTestId(c, "backlog-change-space-dialog-confirm")!.click());
    expect(calls(host, "connection.connect_api_key")).toHaveLength(1);
    expect(byTestId(c, "backlog-status")!.textContent).toBe("Connected as Test User @ other.backlog.jp");
  });

  it("shows and announces the M12 restore notice", async () => {
    const host = scripted(notConnected, {
      "connection.connect_api_key": async () => ({ ...connected, restored: true }),
    });
    const c = await render(host);
    await submitReplace(c, HOST);
    const want = `Reconnected to ${HOST}. Items from your earlier connection to this space were restored.`;
    expect(byTestId(c, "backlog-notice")!.textContent).toContain(want);
    expect(byTestId(c, "backlog-announcement")!.textContent).toBe(want);
    // U4 (M12): restored watches come back Paused.
    expect(byTestId(c, "backlog-notice")!.textContent).toContain(en.restoredWatches);
    await act(async () => byTestId(c, "backlog-review-watches")!.click());
    expect(host.navigate).toHaveBeenCalledWith("/backlog/watches");
  });

  it("puts a data-testid on every interactive element of the connected screen", async () => {
    const c = await render(scripted(connected));
    await act(async () => byTestId(c, "backlog-disconnect")!.click());
    const interactive = c.querySelectorAll("button, input, select, textarea, a");
    expect(interactive.length).toBeGreaterThan(4);
    interactive.forEach((el) => expect(el.getAttribute("data-testid")).toBeTruthy());
    expect(text(c)).toContain(en.disconnectTitle);
  });

  it("mounts the Git access block (US5.5)", async () => {
    const c = await render(scripted({ ...connected, hasGitCredential: true }));
    expect(byTestId(c, "backlog-git-access")).not.toBeNull();
    expect(byTestId(c, "backlog-git-status")!.textContent).toBe(en.gitStored);
  });

  it("passes the Git check of Test connection to the Git block (AC5.5.2)", async () => {
    const view = { ...connected, hasGitCredential: true };
    const c = await render(
      scripted(view, { "connection.test": async () => ({ ...view, gitCheck: "invalid" }) }),
    );
    await act(async () => byTestId(c, "backlog-test-connection")!.click());
    expect(byTestId(c, "backlog-git-invalid")!.textContent).toBe(en.gitInvalid);
  });

  it("says how many PR links and watches a disconnect turns off (AC1.8.1)", async () => {
    const host = scripted(connected, { "connection.disconnect": async () => notConnected });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-disconnect")!.click());
    expect(byTestId(c, "backlog-disconnect-dialog")!.textContent).toContain(
      "2 PR links and 1 PR watches will be turned off.",
    );
    expect(calls(host, "git.impact")[0]![1]).toEqual({ workspaceId: "ws-1", body: {} });
  });

  it("says how many PR links and watches a space change turns off", async () => {
    const c = await render(scripted(connected));
    await submitReplace(c, "other.backlog.jp");
    expect(byTestId(c, "backlog-change-space-dialog")!.textContent).toContain(
      "2 PR links and 1 PR watches will be turned off.",
    );
  });

  it("keeps the dialog usable when the counts cannot be loaded", async () => {
    const c = await render(
      scripted(connected, {
        "git.impact": async () => {
          throw actionError(500, { code: "internal" });
        },
      }),
    );
    await act(async () => byTestId(c, "backlog-disconnect")!.click());
    expect(byTestId(c, "backlog-disconnect-dialog")!.textContent).not.toContain("PR links");
  });
});
