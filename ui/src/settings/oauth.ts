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
  setCookie(cookie: string): void {
    document.cookie = cookie;
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

/** The callback reads this cookie to check the sign-in started in this browser (R-01). */
const VERIFIER_COOKIE = "nulab_backlog_oauth_verifier";
const CALLBACK_PATH = "/api/plugins/nulab-backlog/webhooks/oauth-callback";

/** 32 random bytes, base64url without padding (43 characters). */
function newVerifier(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(32));
  return btoa(String.fromCharCode(...bytes))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

/** Hex SHA-256 of the verifier text; only this is sent to the server at start. */
async function sha256Hex(text: string): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text));
  return Array.from(new Uint8Array(digest), (b) => b.toString(16).padStart(2, "0")).join("");
}

/**
 * Starts the OAuth sign-in and leaves for Backlog's authorization page. A
 * fresh verifier cookie, scoped to the callback path, binds the sign-in to
 * this browser; the server only sees its hash until the callback. It throws
 * when the reply is not that page on the entered host.
 */
export async function startOAuth(host: PluginHostApi, workspaceId: string, spaceUrl: string): Promise<void> {
  const verifier = newVerifier();
  const reply = await host.api.invokeAction<{ authorizeUrl?: unknown }>("connection.start_oauth", {
    workspaceId,
    body: { spaceUrl, verifierHash: await sha256Hex(verifier) },
  });
  const url = typeof reply?.authorizeUrl === "string" ? reply.authorizeUrl : "";
  if (!isAuthorizeUrl(url, normalizeHost(spaceUrl))) throw new Error("unexpected authorize URL");
  browser.setCookie(
    `${VERIFIER_COOKIE}=${verifier}; Path=${CALLBACK_PATH}; Secure; SameSite=Lax; Max-Age=600`,
  );
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
