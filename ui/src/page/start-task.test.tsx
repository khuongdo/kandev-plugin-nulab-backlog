import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  fakeHost,
  mount,
  unmount,
  type Invoke,
} from "../testing/harness";
import type { QuickAction } from "./quick-actions";
import { createStartTask } from "./start-task";

afterEach(unmount);

const ISSUE_ACTIONS: QuickAction[] = [
  {
    id: "implement",
    label: "Implement",
    hint: "Build and open a PR",
    icon: "code",
    promptTemplate: "Build {{url}}",
  },
  {
    id: "investigate",
    label: "Investigate",
    hint: "Find the root cause",
    icon: "search",
    promptTemplate: 'Investigate the Backlog issue at {{url}} (title: "{{title}}").',
  },
];
const URL = "https://example-space.backlog.com/view/PROJ-1";
const PR_URL = "https://example-space.backlog.com/git/PROJ/web-app/pullRequests/7";

async function render(
  invoke: Invoke,
  props: Record<string, unknown> = {},
  overrides: Record<string, unknown> = {},
) {
  const host = fakeHost(invoke, overrides);
  const onLinked = vi.fn();
  const c = await mount(createStartTask(host, en), {
    workspaceId: "ws-1",
    kind: "issue",
    title: "Fix login",
    url: URL,
    issueKey: "PROJ-1",
    actions: ISSUE_ACTIONS,
    testId: "backlog-issue-PROJ-1",
    onLinked,
    ...props,
  });
  return { c, host, onLinked };
}

const pick = (c: HTMLElement, id: string) =>
  act(async () => c.querySelector<HTMLElement>(`[data-preset-id="${id}"]`)!.click());

describe("Start task menu (FR1.1-FR1.4)", () => {
  it("lists the quick actions with icons behind a labelled + Task trigger", async () => {
    const { c } = await render(async () => ({}));
    const trigger = byTestId(c, "backlog-issue-PROJ-1-start")!;
    expect(trigger.getAttribute("aria-label")).toBe("Start a task from PROJ-1");
    const items = c.querySelectorAll('[data-testid="backlog-issue-PROJ-1-start-item"]');
    expect([...items].map((i) => i.getAttribute("data-icon"))).toEqual(["code", "search"]);
    expect(await axeViolations(c)).toEqual([]);
  });

  it("opens Kandev's dialog prefilled, then links the created task to the issue", async () => {
    const { c, host, onLinked } = await render(async () => ({ taskId: "t-1", taskKey: "T-1" }));
    expect(byTestId(c, "fake-task-create-dialog")).toBeNull();
    await pick(c, "investigate");
    expect(byTestId(c, "fake-task-create-dialog")!.getAttribute("data-workflow")).toBe("wf-1/step-1");
    expect(byTestId(c, "fake-task-create-title")!.textContent).toBe("Investigate: Fix login");
    expect(byTestId(c, "fake-task-create-description")!.textContent).toBe(
      `Investigate the Backlog issue at ${URL} (title: "Fix login").`,
    );
    expect(host.api.invokeAction).not.toHaveBeenCalled();
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(host.api.invokeAction).toHaveBeenCalledWith("issues.link", {
      workspaceId: "ws-1",
      taskId: "t-1",
      body: { issueKey: "PROJ-1" },
    });
    expect(onLinked).toHaveBeenCalledWith("t-1", "T-1");
    expect(byTestId(c, "fake-task-create-dialog")).toBeNull();
    expect(byTestId(c, "backlog-issue-PROJ-1-start-notice")).toBeNull();
  });

  it("creates nothing when the dialog is cancelled", async () => {
    const { c, host } = await render(async () => ({}));
    await pick(c, "implement");
    await act(async () => byTestId(c, "fake-task-create-cancel")!.click());
    expect(byTestId(c, "fake-task-create-dialog")).toBeNull();
    expect(host.api.invokeAction).not.toHaveBeenCalled();
  });

  it("keeps the task and toasts that it was not linked when the link fails (FR2.1)", async () => {
    const { c, host, onLinked } = await render(async () => {
      throw actionError(500, { code: "internal", message: "SECRET" });
    });
    await pick(c, "implement");
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(host.toast.error).toHaveBeenCalledWith(en.taskNotLinked);
    expect(byTestId(c, "backlog-issue-PROJ-1-start-notice")).toBeNull();
    expect(c.querySelector('[role="alert"]')).toBeNull();
    expect(onLinked).not.toHaveBeenCalled();
  });

  it("links a pull request by its URL", async () => {
    const { c, host } = await render(async () => ({}), {
      kind: "pr",
      title: "Change 7",
      url: PR_URL,
      issueKey: undefined,
      testId: "backlog-pr-7",
      actions: [{ id: "review", label: "Review", hint: "", icon: "eye", promptTemplate: "Review {{url}}" }],
    });
    await pick(c, "review");
    expect(byTestId(c, "fake-task-create-description")!.textContent).toBe(`Review ${PR_URL}`);
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(host.api.invokeAction).toHaveBeenCalledWith("git.prs.link", {
      workspaceId: "ws-1",
      taskId: "t-1",
      body: { reference: PR_URL },
    });
  });

  it("renders nothing without quick actions", async () => {
    const { c } = await render(async () => ({}), { actions: [] });
    expect(byTestId(c, "backlog-issue-PROJ-1-start")).toBeNull();
  });
});

// Intent 261009: Kandev's browser context is null when /backlog is opened
// directly, although the workspace has a workflow (FR1.2-FR1.6).
const NO_CONTEXT = {
  context: {
    getActiveWorkspaceId: () => "ws-1",
    subscribeActiveWorkspace: () => () => undefined,
    getTaskCreationContext: () => null,
  },
};

/** Answers workflows.status with status, and every link action with linkReply. */
function byAction(status: () => Promise<unknown>, linkReply: unknown = { taskKey: "T-1" }): Invoke {
  return async (key) => (key === "workflows.status" ? status() : linkReply);
}

describe("Start task without a Kandev context (FR1.2-FR1.6)", () => {
  it("regression: opens Kandev's dialog with no workflow when the workspace has one", async () => {
    const { c, host } = await render(
      byAction(async () => ({ hasWorkflow: true })),
      {},
      NO_CONTEXT,
    );
    await pick(c, "investigate");
    expect(host.api.invokeAction).toHaveBeenCalledWith("workflows.status", { workspaceId: "ws-1" });
    const dialog = byTestId(c, "fake-task-create-dialog")!;
    expect(dialog.getAttribute("data-workflow")).toBe("null/null");
    expect(dialog.getAttribute("data-steps")).toBe("0");
    expect(byTestId(c, "fake-task-create-title")!.textContent).toBe("Investigate: Fix login");
    expect(byTestId(c, "fake-task-create-description")!.textContent).toBe(
      `Investigate the Backlog issue at ${URL} (title: "Fix login").`,
    );
    expect(host.toast.error).not.toHaveBeenCalled();
  });

  it("toasts that Kandev has no workflow when the workspace has none (FR1.3)", async () => {
    const { c, host } = await render(
      byAction(async () => ({ hasWorkflow: false })),
      {},
      NO_CONTEXT,
    );
    await pick(c, "implement");
    expect(byTestId(c, "fake-task-create-dialog")).toBeNull();
    expect(host.toast.error).toHaveBeenCalledWith(en.errorWorkflow);
    expect(c.querySelector('[role="alert"]')).toBeNull();
  });

  it("still opens the dialog when the workflow check fails (FR1.4)", async () => {
    const { c, host } = await render(
      byAction(async () => {
        throw actionError(500, { code: "internal" });
      }),
      {},
      NO_CONTEXT,
    );
    await pick(c, "implement");
    expect(byTestId(c, "fake-task-create-dialog")!.getAttribute("data-workflow")).toBe("null/null");
    expect(host.toast.error).not.toHaveBeenCalled();
  });

  it("uses Kandev's context when it has one and never asks the backend (FR1.1, FR3.4)", async () => {
    const step = { id: "step-1", title: "Todo" };
    const { c, host } = await render(
      async () => ({}),
      {},
      {
        context: {
          ...NO_CONTEXT.context,
          getTaskCreationContext: (workspaceId: string) => ({
            workspaceId,
            workflowId: "wf-1",
            defaultStepId: "step-1",
            steps: [step],
            repositories: [],
          }),
        },
      },
    );
    await pick(c, "implement");
    const dialog = byTestId(c, "fake-task-create-dialog")!;
    expect(dialog.getAttribute("data-workflow")).toBe("wf-1/step-1");
    expect(dialog.getAttribute("data-steps")).toBe("1");
    expect(host.api.invokeAction).not.toHaveBeenCalled();
  });

  it("links the task created from the no-workflow dialog (FR1.5)", async () => {
    const { c, host, onLinked } = await render(
      byAction(async () => ({ hasWorkflow: true })),
      {},
      NO_CONTEXT,
    );
    await pick(c, "implement");
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(host.api.invokeAction).toHaveBeenLastCalledWith("issues.link", {
      workspaceId: "ws-1",
      taskId: "t-1",
      body: { issueKey: "PROJ-1" },
    });
    expect(onLinked).toHaveBeenCalledWith("t-1", "T-1");
  });

  it("toasts a failed link after the no-workflow dialog (FR1.5, FR2.1)", async () => {
    const { c, host, onLinked } = await render(
      async (key) => {
        if (key === "workflows.status") return { hasWorkflow: true };
        throw actionError(500, { code: "internal", message: "SECRET" });
      },
      {},
      NO_CONTEXT,
    );
    await pick(c, "implement");
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(host.toast.error).toHaveBeenCalledWith(en.taskNotLinked);
    expect(byTestId(c, "backlog-issue-PROJ-1-start-notice")).toBeNull();
    expect(onLinked).not.toHaveBeenCalled();
  });

  it("follows the same path on a pull request row of another provider (FR1.6)", async () => {
    const { c, host } = await render(
      byAction(async () => ({ hasWorkflow: true }), {}),
      {
        kind: "pr",
        title: "Change 7",
        url: PR_URL,
        issueKey: undefined,
        testId: "backlog-pr-7",
        linkAction: "scm.prs.link",
        actions: [{ id: "review", label: "Review", hint: "", icon: "eye", promptTemplate: "Review {{url}}" }],
      },
      NO_CONTEXT,
    );
    await pick(c, "review");
    expect(byTestId(c, "fake-task-create-dialog")!.getAttribute("data-workflow")).toBe("null/null");
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(host.api.invokeAction).toHaveBeenLastCalledWith("scm.prs.link", {
      workspaceId: "ws-1",
      taskId: "t-1",
      body: { url: PR_URL },
    });
  });
});
