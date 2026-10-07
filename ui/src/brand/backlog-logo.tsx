import type { Component, PluginHostApi, PluginIcon, PluginIconProps } from "@kandev/plugin-sdk";

/**
 * The plugin icon (FR6, BR6.1): an original outline drawing of a ticket card
 * with two text lines and a check mark, in the surrounding text colour like
 * Kandev's own outline icons. It shares no shape with the Nulab Backlog mark,
 * which the plugin no longer ships (docs/brand/backlog-logo.md). Only the
 * host's sizing class is passed through; every other prop is ignored.
 */
export function createBacklogIcon(host: PluginHostApi): Component<PluginIconProps> {
  const h = host.jsx;
  return function BacklogIcon({ className }: PluginIconProps) {
    return (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
        focusable="false"
        className={className}
      >
        <rect x={3} y={4} width={18} height={16} rx={2} />
        <path d="M7 9h6" />
        <path d="M7 13h4" />
        <path d="M13 15l2 2 4-4" />
      </svg>
    );
  };
}

/** The one place the plugin icon is chosen (nav entry, settings card, page header). */
export const PLUGIN_ICON: (host: PluginHostApi) => PluginIcon = createBacklogIcon;
