import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  deferred,
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
import type { LinksStore } from "./links-store";

afterEach(unmount);

const TASKS = [
  { taskId: "task-17", taskKey: "T-17", title: "Login work" },
  { taskId: "task-18", taskKey: "T-18", title: "Login old", linkedIssueKey: "PROJ-118" },
];

/** A fake host; `link` replaces the default issues.link reply. */
function setup(tasks = TASKS, link?: () => Promise<unknown>) {
  return fakeHost(async (key, input) => {
    if (key === "issues.tasks.search") return { tasks };
    if (key === "issues.link" && link) return link();
    if (key === "issues.link")
      return {
        taskId: input?.taskId,
        taskKey: "T-17",
        issueKey: (input?.body as { issueKey: string }).issueKey,
      };
    throw new Error(`unexpected ${key}`);
  });
}

function fakeStore() {
  return {
    get: vi.fn(() => undefined),
    load: vi.fn(async () => undefined),
    refresh: vi.fn(async () => undefined),
    subscribe: vi.fn(() => () => undefined),
  } satisfies LinksStore;
}

async function open(host: ReturnType<typeof setup>, messages = en, store = fakeStore()) {
  const opener = document.createElement("button");
  opener.setAttribute("data-testid", "opener");
  document.body.appendChild(opener);
  opener.focus();
  const onLinked = vi.fn();
  const onClose = vi.fn();
  const c = await mount(createLinkTaskDialog(host, messages, store), {
    workspaceId: "ws-1",
    issueKey: "PROJ-120",
    onLinked,
    onClose,
  });
  return { c, onLinked, onClose, opener, store };
}

describe("Link to task dialog (M3, US3.3)", () => {
  it("is a labelled dialog with focus on the search field", async () => {
    const { c } = await open(setup());
    const dialog = byTestId(c, "backlog-link-task-dialog")!;
    expect(dialog.getAttribute("role")).toBe("dialog");
    expect(dialog.querySelector("h2")!.textContent).toBe("Link PROJ-120 to a task");
    // FR4.1, FR4.2: the GitHub link dialog's width and description line.
    expect(dialog.className).toContain("w-[calc(100vw-2rem)] sm:max-w-lg");
    const description = dialog.querySelector('[data-host="DialogDescription"]')!;
    expect(description.textContent).toBe(en.linkTaskDescription);
    expect(dialog.getAttribute("aria-describedby")).toBe(description.id);
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

  it("links the chosen task with Save, toasts and refreshes the links (AC3.3.1, AC3.3.2, FR4.4, FR4.5)", async () => {
    const host = setup();
    const { c, onLinked, onClose, store } = await open(host);
    await act(async () => byTestId(c, "backlog-link-task-option-task-17")!.click());
    expect(byTestId(c, "backlog-link-task-option-task-17")!.getAttribute("aria-pressed")).toBe("true");
    const submit = byTestId(c, "backlog-link-task-submit") as HTMLButtonElement;
    expect(submit.textContent).toBe(en.save);
    expect(submit.type).toBe("submit");
    await act(async () => submit.click());
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenLastCalledWith("issues.link", {
      workspaceId: "ws-1",
      taskId: "task-17",
      body: { issueKey: "PROJ-120" },
    });
    // The regression: the task's badge read stale links until the 60 s refresh.
    expect(host.toast.success).toHaveBeenCalledWith(en.linkIssueSuccess);
    expect(store.refresh).toHaveBeenCalledWith("ws-1");
    expect(onLinked).toHaveBeenCalledWith({ taskId: "task-17", taskKey: "T-17" });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("shows Saving... and cannot close while saving (FR4.4, FR4.6)", async () => {
    const pending = deferred<unknown>();
    const host = setup(TASKS, () => pending.promise);
    const { c, onClose } = await open(host);
    await act(async () => byTestId(c, "backlog-link-task-option-task-17")!.click());
    await act(async () => byTestId(c, "backlog-link-task-submit")!.click());
    const submit = byTestId(c, "backlog-link-task-submit") as HTMLButtonElement;
    expect(submit.textContent).toBe(en.saving);
    expect(submit.disabled).toBe(true);
    await act(async () => press(byTestId(c, "backlog-link-task-search")!, "Escape"));
    expect(onClose).not.toHaveBeenCalled();
    await act(async () => pending.resolve({ taskKey: "T-17" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("submits on Enter in the search field only once a task is chosen (FR4.4)", async () => {
    const host = setup();
    const { c, onClose } = await open(host);
    const search = byTestId(c, "backlog-link-task-search") as HTMLInputElement;
    await act(async () => search.form!.requestSubmit());
    expect(vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "issues.link")).toHaveLength(0);
    await act(async () => byTestId(c, "backlog-link-task-option-task-17")!.click());
    await act(async () => search.form!.requestSubmit());
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenLastCalledWith("issues.link", expect.anything());
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("shows a failed link as an inline red alert and keeps Save enabled (FR4.3)", async () => {
    const host = setup(TASKS, async () => {
      throw actionError(409, { code: "conflict" });
    });
    const { c, onClose, store } = await open(host);
    await act(async () => byTestId(c, "backlog-link-task-option-task-17")!.click());
    await act(async () => byTestId(c, "backlog-link-task-submit")!.click());
    const error = byTestId(c, "backlog-link-task-error")!;
    expect(error.getAttribute("role")).toBe("alert");
    expect(error.className).toBe("text-xs text-destructive");
    expect(error.textContent).toBe(en.gitConflict);
    const submit = byTestId(c, "backlog-link-task-submit") as HTMLButtonElement;
    expect(submit.textContent).toBe(en.save);
    expect(submit.disabled).toBe(false);
    expect(onClose).not.toHaveBeenCalled();
    expect(store.refresh).not.toHaveBeenCalled();
    expect(host.toast.success).not.toHaveBeenCalled();
  });

  it("says No tasks found with Save disabled", async () => {
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
