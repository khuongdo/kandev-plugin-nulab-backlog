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

  it("keeps the task and says it was not linked when the link fails", async () => {
    const { c, onLinked } = await render(async () => {
      throw actionError(500, { code: "internal", message: "SECRET" });
    });
    await pick(c, "implement");
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    const notice = byTestId(c, "backlog-issue-PROJ-1-start-notice")!;
    expect(notice.getAttribute("role")).toBe("alert");
    expect(notice.textContent).toBe(en.taskNotLinked);
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

  it("explains a missing workflow instead of opening the dialog", async () => {
    const { c } = await render(
      async () => ({}),
      {},
      {
        context: {
          getActiveWorkspaceId: () => "ws-1",
          subscribeActiveWorkspace: () => () => undefined,
          getTaskCreationContext: () => null,
        },
      },
    );
    await pick(c, "implement");
    expect(byTestId(c, "fake-task-create-dialog")).toBeNull();
    expect(byTestId(c, "backlog-issue-PROJ-1-start-notice")!.textContent).toBe(en.errorWorkflow);
  });

  it("renders nothing without quick actions", async () => {
    const { c } = await render(async () => ({}), { actions: [] });
    expect(byTestId(c, "backlog-issue-PROJ-1-start")).toBeNull();
  });
});
