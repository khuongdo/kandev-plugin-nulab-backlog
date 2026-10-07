import { describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import { actionError, fakeHost } from "../testing/harness";
import {
  badgeDetail,
  badgeText,
  issueQueryFilters,
  issuesFailure,
  loadImpactText,
  openStatusIds,
  showingText,
  type LinkView,
} from "./issues-state";

const LINK: LinkView = {
  taskId: "task-17",
  taskKey: "T-17",
  issueKey: "PROJ-120",
  spaceHost: "example-space.backlog.com",
  state: "active",
  status: "Resolved",
  statusUpdatedAt: "2026-10-06T00:00:00Z",
  stale: false,
  unavailable: false,
  url: "https://example-space.backlog.com/view/PROJ-120",
};

describe("issues state helpers (M2, M6)", () => {
  it("maps list failures to the page states (AC1.4.3, AC1.7.2, AC2.1.4)", () => {
    expect(issuesFailure(actionError(401, { code: "reconnect_required" })).kind).toBe("sign_in_again");
    expect(issuesFailure(actionError(400, { code: "validation", field: "projectKeys" })).kind).toBe(
      "no_project",
    );
    expect(issuesFailure(actionError(409, { code: "integration_disabled" })).kind).toBe("not_connected");
    expect(issuesFailure(actionError(503, { code: "unreachable" }))).toEqual({
      kind: "error",
      notice: { key: "unreachable", retry: true },
    });
    expect(issuesFailure(actionError(500, { code: "internal" }))).toEqual({
      kind: "error",
      notice: { key: "issuesLoadFailed", retry: true },
    });
  });

  it("counts down a rate limit (AC8.4.4)", () => {
    expect(issuesFailure(actionError(429, { code: "rate_limited", retryAfterSeconds: 12 }))).toEqual({
      kind: "rate_limited",
      retryAfterSeconds: 12,
      notice: { key: "rateLimitedRetrying", params: { seconds: 12 } },
    });
  });

  it("never shows a server body", () => {
    const err = Object.assign(new Error("<html>secret body</html>"), {
      status: 500,
      body: { error: { code: "internal", message: "secret body" } },
    });
    const f = issuesFailure(err);
    expect(JSON.stringify(f)).not.toContain("secret");
  });

  it('says "Showing x-y of z" (AC2.1.1, AC2.1.2)', () => {
    expect(showingText({ page: 1, pageSize: 20, total: 57 })).toBe("Showing 1-20 of 57");
    expect(showingText({ page: 3, pageSize: 20, total: 57 })).toBe("Showing 41-57 of 57");
    expect(showingText({ page: 1, pageSize: 20, total: 0 })).toBe("");
  });

  it("labels badges with text, never by colour alone (M6)", () => {
    expect(badgeText(LINK)).toBe("PROJ-120 · Resolved");
    expect(badgeText({ ...LINK, state: "not_connected" })).toBe("PROJ-120 – not connected");
    expect(badgeText({ ...LINK, unavailable: true })).toBe("PROJ-120 · Issue unavailable");
    const rel = (v: string) => `rel:${v}`;
    expect(badgeDetail(LINK, rel)).toBe("updated at rel:2026-10-06T00:00:00Z");
    expect(badgeDetail({ ...LINK, stale: true }, rel)).toBe(
      "updated at rel:2026-10-06T00:00:00Z · may be out of date",
    );
    expect(badgeDetail({ ...LINK, state: "not_connected" }, rel)).toBe(
      "Reconnect example-space.backlog.com to restore this link",
    );
  });

  it("treats every status but Closed as open and describes a saved issue query (FR4.1, FR4.3)", () => {
    expect(openStatusIds([{ id: 1 }, { id: 4 }, { id: 2 }, {}])).toEqual([1, 2]);
    expect(openStatusIds(undefined)).toEqual([]);
    expect(
      issueQueryFilters({
        name: "Q",
        projectKey: "PROJ",
        statusIds: [1, 2],
        assignee: "me",
        keyword: "login",
      }),
    ).toBe("PROJ · 2 statuses · assignee Me · keyword “login”");
    expect(issueQueryFilters({ name: "Q", statusIds: [], assignee: "7", keyword: "" })).toBe(
      "all projects · all statuses · assignee user 7 · no keyword",
    );
    expect(issueQueryFilters({ name: "Q", statusIds: [], assignee: "", keyword: "" })).toContain(
      "assignee Anyone",
    );
  });

  it("sums issue and PR impact for the confirm dialogs (AC1.8.1, AC1.9.1)", async () => {
    const host = fakeHost(async (key) => {
      if (key === "issues.impact") return { issueLinks: 4 };
      if (key === "git.impact") return { prLinks: 2, prWatches: 1 };
      throw new Error(key);
    });
    expect(await loadImpactText(host, "ws-1", ["PROJ"])).toBe(
      "4 issue links, 2 PR links and 1 PR watches will be turned off.",
    );
    expect(vi.mocked(host.api.invokeAction).mock.calls).toEqual([
      ["issues.impact", { workspaceId: "ws-1", body: { projectKeys: ["PROJ"] } }],
      ["git.impact", { workspaceId: "ws-1", body: { projectKeys: ["PROJ"] } }],
    ]);
    const none = fakeHost(async () => {
      throw actionError(500, {});
    });
    expect(await loadImpactText(none, "ws-1")).toBe("");
    expect(en.impactAll).toContain("{issues}");
  });
});
