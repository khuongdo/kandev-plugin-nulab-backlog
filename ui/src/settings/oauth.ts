import type { PluginHostApi } from "@kandev/plugin-sdk";

/** The ?oauth= outcomes the callback redirect carries. */
export type OAuthOutcome = "connected" | "cancelled" | "failed";

/** Browser calls, kept in one object so tests can replace them. */
export const browser = {
  assign(url: string): void {
    window.location.assign(url);
  },
  search(): string {
    return window.location.search;
  },
  replaceSearch(search: string): void {
    const url = new URL(window.location.href);
    url.search = search;
    window.history.replaceState(window.history.state, "", url.toString());
  },
};

/** Lower-cases a typed space address and drops https:// and trailing slashes. */
export function normalizeHost(input: string): string {
  return input
    .trim()
    .toLowerCase()
    .replace(/^https:\/\//, "")
    .replace(/\/+$/, "");
}

/** True only for Backlog's authorization page on the entered host. */
export function isAuthorizeUrl(url: string, host: string): boolean {
  return host !== "" && url.startsWith(`https://${host}/OAuth2AccessRequest.action?`);
}

/**
 * Starts the OAuth sign-in and leaves for Backlog's authorization page. It
 * throws when the reply is not that page on the entered host.
 */
export async function startOAuth(host: PluginHostApi, workspaceId: string, spaceUrl: string): Promise<void> {
  const reply = await host.api.invokeAction<{ authorizeUrl?: unknown }>("connection.start_oauth", {
    workspaceId,
    body: { spaceUrl },
  });
  const url = typeof reply?.authorizeUrl === "string" ? reply.authorizeUrl : "";
  if (!isAuthorizeUrl(url, normalizeHost(spaceUrl))) throw new Error("unexpected authorize URL");
  browser.assign(url);
}

/** Reads the callback result from the address once and removes it, so it is shown once. */
export function takeOAuthReturn(): { outcome: OAuthOutcome; restored: boolean } | undefined {
  const params = new URLSearchParams(browser.search());
  const outcome = params.get("oauth");
  if (outcome !== "connected" && outcome !== "cancelled" && outcome !== "failed") return undefined;
  const restored = params.get("restored") === "1";
  params.delete("oauth");
  params.delete("restored");
  const rest = params.toString();
  browser.replaceSearch(rest ? `?${rest}` : "");
  return { outcome, restored };
}
