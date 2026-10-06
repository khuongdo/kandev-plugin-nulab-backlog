import { describe, expect, it, vi } from "vitest";
import type { PluginHostApi, PluginHostRepository, TaskContext } from "@kandev/plugin-sdk";

import { en } from "../messages/en";
import { actionError, fakeHost, type Invoke } from "../testing/harness";
import { createPRLinkAction } from "./pr-link";

const BACKLOG_REPO = {
  id: "repo-k1",
  workspace_id: "ws-1",
  name: "web-app",
  provider: "nulab-backlog",
  provider_scope: "example-space.backlog.com",
} as PluginHostRepository;

function task(repositories: PluginHostRepository[] = [BACKLOG_REPO]): TaskContext {
  return {
    workspaceId: "ws-1",
    taskId: "task-17",
    repositories,
    pathname: "/t/task-17",
    presentation: "desktop",
  };
}

type DialogOptions = Parameters<PluginHostApi["openTaskLinkDialog"]>[0];

async function open(invoke: Invoke, repositories?: PluginHostRepository[]) {
  const host = fakeHost(invoke);
  const action = createPRLinkAction(host);
  await action.run(task(repositories));
  const options = vi.mocked(host.openTaskLinkDialog).mock.calls[0]![0] as DialogOptions;
  return { host, action, options };
}

describe("Link a pull request (M7, US5.2)", () => {
  it("is a link-menu task action that opens Kandev's link dialog with localized copy", async () => {
    const { action, options } = await open(async () => ({}));
    expect(action).toMatchObject({ id: "nulab-backlog-link-pr", label: en.linkPRLabel, placement: "link" });
    expect(options).toMatchObject({
      title: en.linkPRTitle,
      description: en.linkPRDescription,
      inputLabel: en.linkPRInput,
      placeholder: en.linkPRPlaceholder,
      emptyError: en.linkPREmpty,
      failureMessage: en.linkPRFailed,
      successMessage: en.linkPRSuccess,
      inputTestId: "backlog-link-pr-input",
      errorTestId: "backlog-link-pr-error",
      submitTestId: "backlog-link-pr-submit",
    });
  });

  it("links with the task id and its one Backlog repository (AC5.2.1)", async () => {
    const { host, options } = await open(async () => ({ taskId: "task-17" }));
    const signal = new AbortController().signal;
    await options.onSubmit(" 42 ", signal);
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "git.prs.link",
      { workspaceId: "ws-1", taskId: "task-17", repositoryId: "repo-k1", body: { reference: "42" } },
      { signal },
    );
  });

  it("sends no repository when the task has none or several Backlog repositories", async () => {
    const { host, options } = await open(
      async () => ({}),
      [BACKLOG_REPO, { ...BACKLOG_REPO, id: "repo-k2" }],
    );
    await options.onSubmit("web-app#3", new AbortController().signal);
    expect(vi.mocked(host.api.invokeAction).mock.calls[0]![1]).toEqual({
      workspaceId: "ws-1",
      taskId: "task-17",
      body: { reference: "web-app#3" },
    });
  });

  it("rejects with Pull request #999 not found", async () => {
    const { options } = await open(async () => {
      throw actionError(404, { code: "not_found" });
    });
    await expect(options.onSubmit("999", new AbortController().signal)).rejects.toThrow(
      "Pull request #999 not found",
    );
  });

  it("rejects a reference that is not a pull request in the space", async () => {
    const { options } = await open(async () => {
      throw actionError(400, { code: "validation", field: "reference" });
    });
    await expect(
      options.onSubmit("https://github.com/a/b/pull/1", new AbortController().signal),
    ).rejects.toThrow("This link is not a pull request in example-space.backlog.com");
    const other = await open(async () => {
      throw actionError(503, { code: "unreachable" });
    }, []);
    await expect(other.options.onSubmit("1", new AbortController().signal)).rejects.toThrow(en.unreachable);
  });
});
