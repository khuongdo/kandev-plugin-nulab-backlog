import type { PluginHostApi } from "@kandev/plugin-sdk";

export const PLUGIN_ID = "nulab-backlog";

type Listener = (workspaceId: string, enabled: boolean) => void;

// Kandev v0.96.0 lets a plugin publish the switch value (setIntegrationEnabled)
// but offers no way to read it back or observe it, so the plugin's own screens
// follow it through this channel.
const listeners = new Set<Listener>();

/** Publishes a loaded or saved switch value to Kandev and to the plugin's own screens. */
export function publishEnabled(host: PluginHostApi, workspaceId: string, enabled: boolean): void {
  host.setIntegrationEnabled(PLUGIN_ID, workspaceId, enabled);
  for (const listener of listeners) listener(workspaceId, enabled);
}

/** Calls the listener on every published switch value; returns the unsubscribe function. */
export function subscribeEnabled(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}
