import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { PluginHostApi, PluginHostRepository } from "@kandev/plugin-sdk";

import { en } from "../messages/en";
import { actionError, byTestId, fakeHost, mount, unmount, type Invoke } from "../testing/harness";
import { createChangeRequest } from "./create-pr";

afterEach(unmount);

const REPO = {
  id: "repo-k1",
  workspace_id: "ws-1",
  name: "web-app",
  provider: "nulab-backlog",
} as PluginHostRepository;

function request(overrides: Record<string, unknown> = {}) {
  return {
    workspaceId: "ws-1",
    taskId: "task-17",
    sessionId: "session-1",
    repositoryId: "repo-k1",
    repository: REPO,
    title: "Add search",
    body: "Body",
    baseBranch: "main",
    draft: false,
    signal: new AbortController().signal,
    ...overrides,
  };
}

type ModalOptions = Parameters<PluginHostApi["openModal"]>[0];

describe("Create a pull request (M11, US5.3)", () => {
  it("creates with the task selectors and never a source branch (AC5.3.2)", async () => {
    const host = fakeHost(async () => ({
      url: "https://example-space.backlog.com/git/PROJ/web-app/pullRequests/50",
      linked: true,
    }));
    const req = request();
    const got = await createChangeRequest(host)(req);
    expect(got).toEqual({
      url: "https://example-space.backlog.com/git/PROJ/web-app/pullRequests/50",
      provider: "nulab-backlog",
      linked: true,
    });
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "git.prs.create",
      {
        workspaceId: "ws-1",
        taskId: "task-17",
        sessionId: "session-1",
        repositoryId: "repo-k1",
        body: { title: "Add search", body: "Body", baseBranch: "main" },
      },
      { signal: req.signal },
    );
  });

  it("offers to link an already open pull request, and links it on Confirm (AC5.3.4)", async () => {
    const host = fakeHost(async (key) => {
      if (key === "git.prs.create") throw actionError(409, { code: "conflict", pullRequestNumber: 7 });
      return {
        taskId: "task-17",
        spaceHost: "example-space.backlog.com",
        projectKey: "PROJ",
        repoName: "web-app",
        number: 7,
      };
    });
    const pending = createChangeRequest(host)(request());
    await vi.waitFor(() => expect(host.openModal).toHaveBeenCalled());
    const options = vi.mocked(host.openModal).mock.calls[0]![0] as ModalOptions;
    expect(options.dismissible).toBe(false);
    const c = await mount(options.content, {});
    expect(c.textContent).toContain("Pull request #7 is already open for this branch. Link it?");
    expect(document.activeElement).toBe(byTestId(c, "backlog-pr-exists-dialog-cancel"));
    await act(async () => byTestId(c, "backlog-pr-exists-dialog-confirm")!.click());
    await expect(pending).resolves.toEqual({
      url: "https://example-space.backlog.com/git/PROJ/web-app/pullRequests/7",
      provider: "nulab-backlog",
      linked: true,
    });
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "git.prs.link",
      { workspaceId: "ws-1", taskId: "task-17", repositoryId: "repo-k1", body: { reference: "7" } },
      expect.anything(),
    );
  });

  it("creates nothing on Cancel", async () => {
    const host = fakeHost(async () => {
      throw actionError(409, { code: "conflict", pullRequestNumber: 7 });
    });
    const pending = createChangeRequest(host)(request());
    await vi.waitFor(() => expect(host.openModal).toHaveBeenCalled());
    const c = await mount((vi.mocked(host.openModal).mock.calls[0]![0] as ModalOptions).content, {});
    const outcome = expect(pending).rejects.toThrow(en.prNotCreated);
    await act(async () => byTestId(c, "backlog-pr-exists-dialog-cancel")!.click());
    await outcome;
    expect(vi.mocked(host.api.invokeAction).mock.calls.map(([k]) => k)).toEqual(["git.prs.create"]);
  });

  it("rejects a missing title with Title is required", async () => {
    const invoke: Invoke = async () => {
      throw actionError(400, { code: "validation", field: "title" });
    };
    await expect(createChangeRequest(fakeHost(invoke))(request({ title: "" }))).rejects.toThrow(
      "Title is required",
    );
  });

  it("rejects other errors with their notice", async () => {
    const invoke: Invoke = async () => {
      throw actionError(400, { code: "validation", field: "branch" });
    };
    await expect(createChangeRequest(fakeHost(invoke))(request())).rejects.toThrow(en.pushFirst);
  });
});
