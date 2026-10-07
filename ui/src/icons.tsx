import type { PluginHostApi } from "@kandev/plugin-sdk";

// Small stroke icons for icon-only buttons (the plugin cannot bundle
// @tabler/icons-react). Always decorative: the button carries the label.
const PATHS = {
  refresh: ["M20 11a8 8 0 0 0-14.9-4M4 5v4h4", "M4 13a8 8 0 0 0 14.9 4M20 19v-4h-4"],
  more: ["M5 12h.01", "M12 12h.01", "M19 12h.01"],
  plus: ["M12 5v14", "M5 12h14"],
  // An open issue: a ring with a dot, like Kandev's issue state icon.
  issue: ["M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18", "M12 12h.01"],
} as const;

export type IconName = keyof typeof PATHS;

/** An inline stroke icon sized by className, in the current text colour. */
export function icon(host: PluginHostApi, name: IconName, className = "h-4 w-4") {
  const h = host.jsx;
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
      {PATHS[name].map((d) => (
        <path key={d} d={d} />
      ))}
    </svg>
  );
}
