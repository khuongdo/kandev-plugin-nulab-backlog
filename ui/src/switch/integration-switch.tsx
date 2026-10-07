import type { Component, IntegrationSettingsActionProps, PluginHostApi } from "@kandev/plugin-sdk";

import { en, type MessageKey, type Messages } from "../messages/en";
import { readFailure, type ConnectionView } from "../settings/state";
import { PLUGIN_ID, publishEnabled } from "./enabled-events";

interface ControlProps {
  id: string;
  enabled: boolean;
  persist: (enabled: boolean) => Promise<void>;
  name: string;
}

/** The catalogue message for a failed switch save. */
function switchErrorKey(error: unknown): MessageKey {
  const f = readFailure(error);
  if (f.status === 403) return "switchForbidden";
  if (f.code === "validation") return "errorInput";
  return "switchFailed";
}

/**
 * The on/off switch in the Backlog card's action slot (WF6, BR7.5). It renders
 * the host's drafted switch for the routed workspace; the host calls persist
 * on Save, and a rejected persist keeps the change unsaved.
 */
export function createIntegrationSwitch(
  host: PluginHostApi,
  messages: Messages = en,
): Component<IntegrationSettingsActionProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const Control = host.ui.IntegrationEnabledControl as Component<ControlProps>;

  return function BacklogIntegrationSwitch({ workspaceId }: IntegrationSettingsActionProps) {
    const [enabled, setEnabled] = useState<boolean | undefined>(undefined);

    useEffect(() => {
      if (!workspaceId) return;
      let live = true;
      host.api
        .invokeAction<ConnectionView>("connection.get", { workspaceId })
        .then((view) => {
          if (!live) return;
          const value = view.enabled === true; // opt-in: anything but an explicit on is off
          setEnabled(value);
          publishEnabled(host, workspaceId, value);
        })
        .catch(() => undefined); // the settings screen below shows the load failure
      return () => {
        live = false;
      };
    }, [workspaceId]);

    if (!workspaceId || enabled === undefined) return null;

    const persist = async (next: boolean) => {
      try {
        // The reply carries only { enabled } (NFR1.3: no secret read for a full view).
        await host.api.invokeAction<{ enabled: boolean }>("connection.set_enabled", {
          workspaceId,
          body: { enabled: next },
        });
      } catch (error) {
        throw new Error(messages[switchErrorKey(error)], { cause: error });
      }
      setEnabled(next);
      publishEnabled(host, workspaceId, next);
    };

    return (
      <span data-testid="backlog-integration-switch">
        <Control id={PLUGIN_ID} enabled={enabled} persist={persist} name={messages.integrationLabel} />
      </span>
    );
  };
}
