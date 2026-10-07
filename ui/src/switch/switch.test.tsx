import * as React from "react";
import { act } from "react";
import { afterEach, describe, expect, it } from "vitest";

import { publishEnabled } from "./enabled-events";
import { createIntegrationSwitch } from "./integration-switch";
import { createSettingsScreen } from "../settings/SettingsScreen";
import { en } from "../messages/en";
import {
  actionError,
  byTestId,
  connected,
  deferred,
  notConnected,
  fakeHost,
  fakeUi,
  mount,
  unmount,
  type Invoke,
  type View,
} from "../testing/harness";

interface ControlProps {
  id: string;
  enabled: boolean;
  persist: (enabled: boolean) => Promise<void>;
  name: string;
}

afterEach(unmount);

/** Renders the card action with a stand-in for host.ui.IntegrationEnabledControl; `null` routes no workspace. */
async function renderSwitch(invoke: Invoke, workspace: string | null = "ws-9") {
  const workspaceId = workspace ?? undefined;
  let control: ControlProps | undefined;
  const IntegrationEnabledControl = (props: ControlProps) => {
    control = props;
    return React.createElement("button", { role: "switch", "aria-checked": props.enabled });
  };
  const host = fakeHost(invoke, {
    ui: { ...fakeUi, IntegrationEnabledControl },
  });
  const c = await mount(createIntegrationSwitch(host), { workspaceId, surface: "detail" });
  return { c, host, control: () => control };
}

function scripted(view: View, setEnabled?: () => Promise<unknown>) {
  return async (key: string) => {
    if (key === "connection.get") return view;
    return setEnabled ? setEnabled() : { ...view, enabled: !view.enabled };
  };
}

describe("integration switch (card action)", () => {
  it("renders the host's drafted switch for the routed workspace with the stored value", async () => {
    const { c, host, control } = await renderSwitch(scripted({ ...connected, enabled: false }));
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.get", { workspaceId: "ws-9" });
    expect(byTestId(c, "backlog-integration-switch")).not.toBeNull();
    expect(control()).toMatchObject({ id: "nulab-backlog", enabled: false, name: en.integrationLabel });
  });

  it("renders nothing while loading and when no workspace is routed", async () => {
    const pending = deferred<unknown>();
    const loading = await renderSwitch(() => pending.promise);
    expect(loading.control()).toBeUndefined();
    unmount();
    const none = await renderSwitch(scripted(connected), null);
    expect(none.control()).toBeUndefined();
    expect(none.host.api.invokeAction).not.toHaveBeenCalled();
  });

  it("publishes the loaded value after a successful load", async () => {
    const { host } = await renderSwitch(scripted({ ...connected, enabled: false }));
    expect(host.setIntegrationEnabled).toHaveBeenCalledWith("nulab-backlog", "ws-9", false);
  });

  it("persists through connection.set_enabled and publishes only after success", async () => {
    const { host, control } = await renderSwitch(scripted(connected));
    (host.setIntegrationEnabled as unknown as { mockClear(): void }).mockClear();
    await control()!.persist(false);
    expect(host.api.invokeAction).toHaveBeenCalledWith("connection.set_enabled", {
      workspaceId: "ws-9",
      body: { enabled: false },
    });
    expect(host.setIntegrationEnabled).toHaveBeenCalledWith("nulab-backlog", "ws-9", false);
  });

  it.each([
    ["403", actionError(403, {}), en.switchForbidden],
    ["validation", actionError(400, { code: "validation", field: "enabled" }), en.errorInput],
    ["internal", actionError(500, { code: "internal" }), en.switchFailed],
  ])(
    "rejects with a catalogue message on %s so the host keeps the change unsaved",
    async (_, err, message) => {
      const { host, control } = await renderSwitch(
        scripted(connected, async () => {
          throw err;
        }),
      );
      (host.setIntegrationEnabled as unknown as { mockClear(): void }).mockClear();
      await expect(control()!.persist(false)).rejects.toThrow(message);
      expect(host.setIntegrationEnabled).not.toHaveBeenCalled();
    },
  );
});

describe("settings screen and switch together", () => {
  /** Mounts the card switch and the M1 screen for one workspace, on a fake store that keeps the switch. */
  async function renderBoth(initiallyEnabled: boolean) {
    let enabled = initiallyEnabled;
    let control: ControlProps | undefined;
    const IntegrationEnabledControl = (props: ControlProps) => {
      control = props;
      return React.createElement("button", { role: "switch", "aria-checked": props.enabled });
    };
    const host = fakeHost(
      async (key, input) => {
        if (key === "connection.get") return { ...notConnected, enabled };
        enabled = (input?.body as { enabled: boolean }).enabled;
        return { enabled };
      },
      { ui: { ...fakeUi, IntegrationEnabledControl } },
    );
    const Switch = createIntegrationSwitch(host);
    const Settings = createSettingsScreen(host);
    const Both = () =>
      React.createElement(
        React.Fragment,
        null,
        React.createElement(Switch, { workspaceId: "ws-1", surface: "detail" }),
        React.createElement(Settings, { workspaceId: "ws-1" }),
      );
    const c = await mount(Both, {});
    const persist = (next: boolean) => act(() => control!.persist(next));
    return { c, host, persist };
  }

  it("restores the Connect form after a saved Off then On, without a reload", async () => {
    const { c, host, persist } = await renderBoth(false);
    expect(byTestId(c, "backlog-off")).not.toBeNull();
    expect(byTestId(c, "backlog-connect-form")).toBeNull();

    await persist(true);
    expect(byTestId(c, "backlog-off")).toBeNull();
    expect(byTestId(c, "backlog-connect-form")).not.toBeNull();

    await persist(false);
    expect(byTestId(c, "backlog-off")).not.toBeNull();
    expect(byTestId(c, "backlog-connect-form")).toBeNull();

    const gets = (host.api.invokeAction as unknown as { mock: { calls: unknown[][] } }).mock.calls.filter(
      ([key]) => key === "connection.get",
    );
    expect(gets).toHaveLength(2); // the two first loads only: no reload
  });

  it("ignores a switch published for another workspace", async () => {
    const { c, host } = await renderBoth(false);
    await act(async () => publishEnabled(host, "ws-other", true));
    expect(byTestId(c, "backlog-off")).not.toBeNull();
    expect(byTestId(c, "backlog-connect-form")).toBeNull();
  });

  it("keeps the screen as it is when the save fails", async () => {
    const { c, host, persist } = await renderBoth(false);
    (host.api.invokeAction as unknown as { mockImplementation(f: Invoke): void }).mockImplementation(
      async (key) => {
        if (key === "connection.get") return { ...notConnected, enabled: false };
        throw actionError(500, { code: "internal" });
      },
    );
    await expect(persist(true)).rejects.toThrow(en.switchFailed);
    expect(byTestId(c, "backlog-off")).not.toBeNull();
    expect(byTestId(c, "backlog-connect-form")).toBeNull();
  });
});
