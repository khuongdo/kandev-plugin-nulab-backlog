import { describe, expect, it, vi } from "vitest";

import { fakeHost } from "../testing/harness";
import { interpolate, loadQuickActions, NO_ACTIONS, taskTitle, TITLE_MAX } from "./quick-actions";

describe("interpolate (FR1.2)", () => {
  it("fills {{url}} and {{title}}, also the legacy single braces", () => {
    expect(
      interpolate("Fix {{url}} ({{title}}) {url} {title}", { url: "https://x/view/P-1", title: "Bug" }),
    ).toBe("Fix https://x/view/P-1 (Bug) https://x/view/P-1 Bug");
  });

  it("keeps unknown placeholders as written", () => {
    expect(interpolate("{{branch}} {{url}} {{ title }}", { url: "u", title: "t" })).toBe(
      "{{branch}} u {{ title }}",
    );
  });

  it("does not expand placeholders inside the inserted title", () => {
    expect(interpolate("{{title}}", { url: "u", title: "{{url}}" })).toBe("{{url}}");
  });
});

describe("taskTitle (FR1.2)", () => {
  it("is <label>: <title>", () => {
    expect(taskTitle("Investigate", "Fix login")).toBe("Investigate: Fix login");
  });

  it("truncates like Kandev, by characters, with an ellipsis", () => {
    const long = taskTitle("Implement", "ログ".repeat(60));
    expect(Array.from(long)).toHaveLength(TITLE_MAX);
    expect(long.endsWith("…")).toBe(true);
    expect(taskTitle("A", "b".repeat(TITLE_MAX - 3))).toHaveLength(TITLE_MAX);
  });
});

describe("loadQuickActions (FR2.2, NFR4)", () => {
  it("reads issues.quick_actions.get once and falls back to no actions on failure", async () => {
    const actions = { issue: [{ id: "a", label: "A", hint: "", icon: "eye", promptTemplate: "" }], pr: [] };
    const host = fakeHost(async () => actions);
    expect(await loadQuickActions(host, "ws-1")).toEqual(actions);
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenCalledWith("issues.quick_actions.get", {
      workspaceId: "ws-1",
    });
    const failing = fakeHost(async () => {
      throw new Error("down");
    });
    expect(await loadQuickActions(failing, "ws-1")).toEqual(NO_ACTIONS);
  });
});
