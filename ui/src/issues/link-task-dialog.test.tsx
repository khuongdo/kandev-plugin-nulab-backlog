import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  axeViolations,
  byTestId,
  expectOnlyCatalogueText,
  expectTestIds,
  fakeHost,
  mount,
  press,
  pseudoCatalogue,
  setValue,
  unmount,
} from "../testing/harness";
import { createLinkTaskDialog } from "./link-task-dialog";

afterEach(unmount);

const TASKS = [
  { taskId: "task-17", taskKey: "T-17", title: "Login work" },
  { taskId: "task-18", taskKey: "T-18", title: "Login old", linkedIssueKey: "PROJ-118" },
];

function setup(tasks = TASKS) {
  return fakeHost(async (key, input) => {
    if (key === "issues.tasks.search") return { tasks };
    if (key === "issues.link")
      return {
        taskId: input?.taskId,
        taskKey: "T-17",
        issueKey: (input?.body as { issueKey: string }).issueKey,
      };
    throw new Error(`unexpected ${key}`);
  });
}

async function open(host: ReturnType<typeof setup>, messages = en) {
  const opener = document.createElement("button");
  opener.setAttribute("data-testid", "opener");
  document.body.appendChild(opener);
  opener.focus();
  const onLinked = vi.fn();
  const onClose = vi.fn();
  const c = await mount(createLinkTaskDialog(host, messages), {
    workspaceId: "ws-1",
    issueKey: "PROJ-120",
    onLinked,
    onClose,
  });
  return { c, onLinked, onClose, opener };
}

describe("Link to task dialog (M3, US3.3)", () => {
  it("is a labelled dialog with focus on the search field", async () => {
    const { c } = await open(setup());
    const dialog = byTestId(c, "backlog-link-task-dialog")!;
    expect(dialog.getAttribute("role")).toBe("dialog");
    expect(dialog.querySelector("h2")!.textContent).toBe("Link PROJ-120 to a task");
    expect(document.activeElement).toBe(byTestId(c, "backlog-link-task-search"));
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("searches tasks and marks one linked to another issue (AC3.3.3)", async () => {
    const host = setup();
    const { c } = await open(host);
    await act(async () => setValue(byTestId(c, "backlog-link-task-search") as HTMLInputElement, "login"));
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenLastCalledWith("issues.tasks.search", {
      workspaceId: "ws-1",
      body: { query: "login" },
    });
    const other = byTestId(c, "backlog-link-task-option-task-18")!;
    expect(other.getAttribute("aria-disabled")).toBe("true");
    expect(other.textContent).toContain("Linked to PROJ-118");
    await act(async () => other.click());
    expect((byTestId(c, "backlog-link-task-submit") as HTMLButtonElement).disabled).toBe(true);
  });

  it("links the chosen task (AC3.3.1, AC3.3.2)", async () => {
    const host = setup();
    const { c, onLinked, onClose } = await open(host);
    await act(async () => byTestId(c, "backlog-link-task-option-task-17")!.click());
    expect(byTestId(c, "backlog-link-task-option-task-17")!.getAttribute("aria-pressed")).toBe("true");
    await act(async () => byTestId(c, "backlog-link-task-submit")!.click());
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenLastCalledWith("issues.link", {
      workspaceId: "ws-1",
      taskId: "task-17",
      body: { issueKey: "PROJ-120" },
    });
    expect(onLinked).toHaveBeenCalledWith({ taskId: "task-17", taskKey: "T-17" });
    expect(onClose).toHaveBeenCalled();
  });

  it("says No tasks found with Link disabled", async () => {
    const { c } = await open(setup([]));
    expect(byTestId(c, "backlog-link-task-none")!.textContent).toBe(en.noTasksFound);
    expect((byTestId(c, "backlog-link-task-submit") as HTMLButtonElement).disabled).toBe(true);
  });

  it("Esc and Cancel change nothing and return focus to the opener (AC3.3.4, AC8.2.3)", async () => {
    const host = setup();
    const { c, onClose, opener } = await open(host);
    await act(async () => press(document.activeElement as HTMLElement, "Escape"));
    expect(onClose).toHaveBeenCalledTimes(1);
    await act(async () => byTestId(c, "backlog-link-task-cancel")!.click());
    expect(onClose).toHaveBeenCalledTimes(2);
    unmount();
    expect(document.activeElement).toBe(opener);
    expect(vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "issues.link")).toHaveLength(0);
    opener.remove();
  });

  it("uses only catalogue text", async () => {
    const { c } = await open(
      setup([{ taskId: "t1", taskKey: "⟦T-1⟧", title: "⟦A⟧", linkedIssueKey: "PROJ-1" }]),
      pseudoCatalogue(en),
    );
    expectOnlyCatalogueText(c, (ok, msg) => expect(ok, msg).toBe(true));
  });
});
