import { describe, expect, it } from "vitest";

import { en, format } from "../messages/en";
import { actionError, connected, HOST, signInAgain } from "../testing/harness";
import {
  connectedNotice,
  failureNotice,
  initialState,
  isSignInAgain,
  reduce,
  type ScreenState,
} from "./state";

const ready: ScreenState = { ...initialState, load: "ready", view: connected };
const say = (n: { key: keyof typeof en; params?: Record<string, string | number> } | undefined) =>
  n ? format(en[n.key], n.params) : "";

describe("U2 reducer", () => {
  it("maps reconnect_required to the revoked-key notice", () => {
    const f = failureNotice(actionError(401, { code: "reconnect_required" }));
    expect(say(f.notice)).toContain("key is invalid or has been revoked");
  });

  it("treats a sign_in_again view as the Sign-in-again state", () => {
    expect(isSignInAgain(signInAgain)).toBe(true);
    expect(isSignInAgain(connected)).toBe(false);
    expect(connectedNotice(signInAgain)).toBeUndefined();
  });

  it("maps rate_limited on a read to the retrying countdown", () => {
    const f = failureNotice(actionError(429, { code: "rate_limited", retryAfterSeconds: 5 }), {
      retrying: true,
    });
    expect(say(f.notice)).toBe("Backlog is limiting requests. Retrying in 5 seconds");
    expect(f.retryAfterSeconds).toBe(5);
  });

  it("announces a countdown only once while the visible seconds change", () => {
    const notice = { key: "rateLimitedRetrying" as const, params: { seconds: 5 } };
    let s = reduce(ready, { type: "announce", notice });
    const first = s.announcement;
    s = reduce(s, { type: "viewChanged", view: connected });
    expect(s.announcement).toBe(first);
  });

  it.each([
    ["connected", en.oauthConnected],
    ["cancelled", en.oauthCancelled],
    ["failed", en.oauthFailed],
  ] as const)("shows and announces the oauth=%s return", (outcome, message) => {
    const s = reduce(ready, { type: "oauthReturned", outcome, restored: false });
    expect(say(s.notice)).toBe(message);
    expect(say(s.announcement)).toBe(message);
  });

  it("shows the M12 notice for a restored connection", () => {
    const want = `Reconnected to ${HOST}. Items from your earlier connection to this space were restored.`;
    expect(say(reduce(ready, { type: "oauthReturned", outcome: "connected", restored: true }).notice)).toBe(
      want,
    );
    const s = reduce(ready, { type: "connected", view: { ...connected, restored: true } });
    expect(say(s.notice)).toBe(want);
    expect(say(s.announcement)).toBe(want);
  });

  it("maps validation on oauth to the not-set-up notice", () => {
    const f = failureNotice(actionError(400, { code: "validation", field: "oauth" }));
    expect(say(f.notice)).toContain("OAuth is not set up on this Kandev server");
    const s = reduce(ready, {
      type: "connectFailed",
      error: actionError(400, { code: "validation", field: "oauth" }),
    });
    expect(say(s.notice)).toContain("OAuth is not set up on this Kandev server");
  });

  it("maps validation on projectKeys to a field error", () => {
    const f = failureNotice(actionError(400, { code: "validation", field: "projectKeys" }));
    expect(f.fieldError).toEqual({ field: "projectKeys", key: "errorProjectKeys" });
  });

  it("updates the view after a panel action", () => {
    const s = reduce(ready, { type: "viewChanged", view: { ...connected, connectedUserName: "New" } });
    expect(s.view?.connectedUserName).toBe("New");
  });
});
