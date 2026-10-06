/** English message catalogue. Every user-facing string comes from here (BR6.2). */
export const en = {
  integrationLabel: "Backlog",
  integrationDescription: "Connect this workspace to a Nulab Backlog space.",
  loading: "Loading the Backlog settings...",
  loadFailed: "Could not load the Backlog settings.",
  retry: "Retry",
  notConnectedMember: "Backlog is not connected. Ask a workspace admin to connect it.",
  connectedAs: "Connected as {name} @ {host}",
  replaceHeading: "Replace connection",
  incomplete: "The Backlog connection is incomplete. Connect again.",
  spaceUrlLabel: "Space address",
  spaceUrlPlaceholder: "myteam.backlog.com",
  apiKeyLabel: "API key",
  connect: "Connect",
  connecting: "Connecting...",
  errorSpaceUrl:
    "Enter a Backlog space address such as myteam.backlog.com, myteam.backlog.jp or myteam.backlogtool.com, and check that the space exists.",
  errorApiKey:
    "The API key is invalid. Create one in Backlog under Personal Settings > API, then paste it here.",
  errorInput: "The request was not valid. Check the fields and try again.",
  unreachable: "Could not reach Backlog.",
  rateLimited: "Backlog is limiting requests. Try again in {seconds} s",
  busy: "Another connection is being set up. Try again shortly.",
  connectFailed: "Could not save the connection. Try again.",
  integrationOff: "Backlog is turned off for this workspace. Turn it on with the switch above to connect.",
  switchForbidden: "Only a Kandev admin can turn Backlog on or off.",
  switchFailed: "Could not save the Backlog setting. Try again.",
  pageNoWorkspace: "Select a workspace to use Backlog.",
  pageLoading: "Loading Backlog...",
  pageLoadFailed: "Could not load Backlog.",
  pageOff: "Backlog is turned off for this workspace.",
  pageNotConnected: "Backlog is not connected.",
  openSettings: "Open the Backlog settings",
} as const;

export type MessageKey = keyof typeof en;
export type Messages = Record<MessageKey, string>;

/** Replaces {name} placeholders with params. */
export function format(template: string, params: Record<string, string | number> = {}): string {
  return template.replace(/\{(\w+)\}/g, (match, name: string) =>
    name in params ? String(params[name]) : match,
  );
}
