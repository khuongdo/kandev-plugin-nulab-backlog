import { describe, expect, it } from "vitest";

import { en } from "../messages/en";
import { readFailure } from "../settings/state";
import { actionError } from "../testing/harness";
import {
  gitNotice,
  noticeText,
  prBadgeLabel,
  providerName,
  scmNotice,
  stateText,
  watchProgress,
  watchStatusKey,
} from "./git-state";

describe("git-state (M4, M6)", () => {
  it("builds the PR badge screen-reader text", () => {
    expect(prBadgeLabel({ changeRequestNumber: 42, state: "open", assignee: "Lan" })).toBe(
      "Pull request #42, Open, assignee Lan",
    );
    expect(prBadgeLabel({ changeRequestNumber: 7, state: "merged" })).toBe("Pull request #7, Merged");
    expect(prBadgeLabel({ changeRequestNumber: 8, state: "closed" })).toBe("Pull request #8, Closed");
    expect(prBadgeLabel({ changeRequestNumber: 9, state: "unknown" })).toBe(
      "Pull request #9, Status unknown",
    );
  });

  it("names every watch state", () => {
    expect(en[watchStatusKey("active")]).toBe("Active");
    expect(en[watchStatusKey("paused")]).toBe("Paused");
    expect(en[watchStatusKey("not_connected")]).toBe("Not connected");
  });

  it("shows the progress only while tasks are pending", () => {
    const n = watchProgress({ createdCount: 10, pendingCount: 15 });
    expect(n && noticeText(n)).toBe("Created 10/25 tasks, the rest in later cycles");
    expect(watchProgress({ createdCount: 25, pendingCount: 0 })).toBeUndefined();
  });

  it("maps action errors to notices", () => {
    expect(gitNotice(actionError(401, { code: "reconnect_required" }))).toEqual({ key: "reconnectRequired" });
    expect(noticeText(gitNotice(actionError(429, { code: "rate_limited", retryAfterSeconds: 30 })))).toBe(
      "Backlog is limiting requests. Try again in 30 s",
    );
    expect(gitNotice(actionError(429, { code: "rate_limited", retryAfterSeconds: 30 })).retry).toBe(true);
    expect(gitNotice(actionError(503, { code: "unreachable" }))).toEqual({ key: "unreachable", retry: true });
    expect(gitNotice(actionError(404, { code: "not_found" }))).toEqual({ key: "gitNotFound" });
    expect(gitNotice(actionError(409, { code: "conflict" }))).toEqual({ key: "gitConflict" });
    expect(gitNotice(actionError(409, { code: "integration_disabled" }))).toEqual({ key: "pageOff" });
    expect(gitNotice(new Error("boom"))).toEqual({ key: "actionFailed", retry: true });
  });

  it("reads the open pull request number of a conflict", () => {
    expect(readFailure(actionError(409, { code: "conflict", pullRequestNumber: 7 })).pullRequestNumber).toBe(
      7,
    );
    expect(readFailure(actionError(409, { code: "conflict" })).pullRequestNumber).toBeUndefined();
  });

  it("formats notices from a given catalogue", () => {
    const pseudo = { ...en, gitSaved: "⟦Saved⟧" };
    expect(noticeText({ key: "gitSaved" }, pseudo)).toBe("⟦Saved⟧");
  });
});

describe("source control states and notices (A3, FR2.4, NFR5)", () => {
  it("names draft and declined pull requests", () => {
    expect(stateText("draft")).toBe(en.stateDraft);
    expect(stateText("declined")).toBe(en.stateDeclined);
  });

  it("maps provider failures to provider words", () => {
    expect(scmNotice(actionError(401, { code: "reconnect_required" }))).toEqual({ key: "scmTokenRefused" });
    expect(scmNotice(actionError(400, { code: "validation", field: "token" }))).toEqual({
      key: "scmNoToken",
    });
    expect(noticeText(scmNotice(actionError(429, { code: "rate_limited", retryAfterSeconds: 9 })))).toBe(
      "The provider is limiting requests. Try again in 9 s",
    );
    expect(scmNotice(actionError(503, { code: "unreachable" }))).toEqual({
      key: "scmUnreachable",
      retry: true,
    });
    expect(scmNotice(actionError(409, { code: "conflict" }))).toEqual({ key: "scmUnmapped" });
    expect(noticeText(scmNotice(actionError(503, { code: "cli_unavailable" })))).toBe(
      "The gh / glab CLI is not available or not logged in on the Kandev server.",
    );
    expect(scmNotice(actionError(404, { code: "not_found" }))).toEqual({ key: "gitNotFound" });
  });

  it("names the providers", () => {
    expect(providerName("github")).toBe("GitHub");
    expect(providerName("gitlab")).toBe("GitLab");
    expect(providerName("bitbucket")).toBe("Bitbucket");
    expect(providerName("backlog")).toBe("Backlog Git");
  });
});
