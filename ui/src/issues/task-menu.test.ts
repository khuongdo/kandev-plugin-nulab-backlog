import { describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import { actionError, fakeHost } from "../testing/harness";
import { createLinksStore } from "./links-store";
import { createUnlinkMenuAction } from "./task-menu";

const CTX = {
  workspaceId: "ws-1",
  taskId: "task-17",
  taskTitle: "T",
  workflowStepId: null,
  presentation: "desktop" as const,
};

function setup(fail = false) {
  let links = [{ taskId: "task-17", issueKey: "PROJ-120", state: "active" }];
  const host = fakeHost(async (key) => {
    if (key === "issues.links.list") return { links };
    if (key === "issues.unlink") {
      if (fail) throw actionError(500, { code: "internal" });
      links = [];
      return { ok: true };
    }
    throw new Error(`unexpected ${key}`);
  });
  const store = createLinksStore(host);
  return { host, store, action: createUnlinkMenuAction(host, store) };
}

describe("Unlink Backlog issue in the task menu (US3.3)", () => {
  it("is a primary menu action with the catalogue label", () => {
    const { action } = setup();
    expect(action.label).toBe(en.unlinkIssue);
    expect(action.group).toBe("primary");
    expect(action.id).toBe("backlog-unlink-issue");
  });

  it("is hidden until the links are loaded, and starts the load", async () => {
    const { host, action, store } = setup();
    expect(action.visible!(CTX)).toBe(false);
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenCalledWith("issues.links.list", {
      workspaceId: "ws-1",
    });
    await store.load("ws-1");
    expect(action.visible!(CTX)).toBe(true);
  });

  it("is visible only for a linked task, synchronously", async () => {
    const { action, store } = setup();
    await store.load("ws-1");
    expect(action.visible!({ ...CTX, taskId: "task-99" })).toBe(false);
    expect(action.visible!(CTX)).toBe(true);
  });

  it("unlinks with the task id and refreshes the shared links (AC3.3.5)", async () => {
    const { host, action, store } = setup();
    await store.load("ws-1");
    await action.run(CTX);
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenCalledWith("issues.unlink", {
      workspaceId: "ws-1",
      taskId: "task-17",
    });
    expect(action.visible!(CTX)).toBe(false);
    expect(host.toast.success).toHaveBeenCalledWith(en.unlinked);
  });

  it("shows an error toast and keeps the link when the unlink fails", async () => {
    const { host, action, store } = setup(true);
    await store.load("ws-1");
    await action.run(CTX);
    expect(host.toast.error).toHaveBeenCalledWith(en.actionFailed);
    expect(action.visible!(CTX)).toBe(true);
  });
});
