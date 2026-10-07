import { act } from "react";
import axe from "axe-core";
import { afterEach, describe, expect, it, vi } from "vitest";

import { createConfirmDialog } from "./confirm-dialog";
import { en } from "../messages/en";
import { byTestId, deferred, fakeHost, mount, press, unmount } from "../testing/harness";

afterEach(unmount);

async function open(onConfirm: () => Promise<void>, onClose = vi.fn()) {
  const host = fakeHost(async () => undefined);
  const opener = document.createElement("button");
  document.body.appendChild(opener);
  opener.focus();
  const c = await mount(createConfirmDialog(host), {
    testId: "test-dialog",
    title: "Title",
    body: "Body text",
    confirmLabel: "Do it",
    onConfirm,
    onClose,
  });
  return { c, opener, onClose };
}

describe("ConfirmDialog", () => {
  it("is a labelled host dialog with the default focus on Cancel", async () => {
    const { c } = await open(async () => undefined);
    const dialog = byTestId(c, "test-dialog")!;
    expect(dialog.getAttribute("data-host")).toBe("DialogContent");
    expect(dialog.getAttribute("role")).toBe("dialog");
    expect(dialog.getAttribute("aria-modal")).toBe("true");
    expect(document.getElementById(dialog.getAttribute("aria-labelledby")!)!.textContent).toBe("Title");
    expect(document.getElementById(dialog.getAttribute("aria-describedby")!)!.textContent).toBe("Body text");
    expect(document.activeElement).toBe(byTestId(c, "test-dialog-cancel"));
    const result = await axe.run(c, { rules: { "color-contrast": { enabled: false } } });
    expect(result.violations.map((v) => v.id)).toEqual([]);
  });

  it("styles Cancel as outline and a destructive confirm as destructive (BR5.2)", async () => {
    const host = fakeHost(async () => undefined);
    const c = await mount(createConfirmDialog(host), {
      testId: "test-dialog",
      title: "Title",
      body: "Body text",
      confirmLabel: "Delete",
      destructive: true,
      onConfirm: async () => undefined,
      onClose: vi.fn(),
    });
    expect(byTestId(c, "test-dialog-cancel")!.getAttribute("data-variant")).toBe("outline");
    expect(byTestId(c, "test-dialog-confirm")!.getAttribute("data-variant")).toBe("destructive");
  });

  it("closes on Esc and returns the focus to its opener", async () => {
    const onClose = vi.fn();
    const { c, opener } = await open(async () => undefined, onClose);
    act(() => press(byTestId(c, "test-dialog-cancel")!, "Escape"));
    expect(onClose).toHaveBeenCalledTimes(1);
    unmount();
    expect(document.activeElement).toBe(opener);
  });

  it("shows Working... with a disabled confirm button while working", async () => {
    const pending = deferred<void>();
    const { c, onClose } = await open(() => pending.promise);
    const confirm = byTestId(c, "test-dialog-confirm") as HTMLButtonElement;
    await act(async () => confirm.click());
    expect(confirm.textContent).toBe(en.working);
    expect(confirm.disabled).toBe(true);
    await act(async () => pending.resolve());
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("stays open with an inline error when the action fails", async () => {
    const { c, onClose } = await open(async () => {
      throw new Error("boom");
    });
    await act(async () => (byTestId(c, "test-dialog-confirm") as HTMLButtonElement).click());
    expect(byTestId(c, "test-dialog")).not.toBeNull();
    expect(byTestId(c, "test-dialog-error")!.textContent).toBe(en.actionFailed);
    expect(byTestId(c, "test-dialog-error")!.getAttribute("role")).toBe("alert");
    expect((byTestId(c, "test-dialog-confirm") as HTMLButtonElement).disabled).toBe(false);
    expect(onClose).not.toHaveBeenCalled();
  });
});
