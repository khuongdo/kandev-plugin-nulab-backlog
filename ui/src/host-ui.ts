import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

export type AnyProps = Record<string, unknown>;

/** The host UI kit as components with loose props (the SDK types them as opaque). */
export function hostUi(host: PluginHostApi): Record<string, Component<AnyProps>> {
  return host.ui as unknown as Record<string, Component<AnyProps>>;
}
