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
  pseudoCatalogue,
  setValue,
  unmount,
} from "../testing/harness";
import { createPollInterval } from "./poll-interval";

afterEach(unmount);

function setup(
  save: (body: unknown) => Promise<unknown> = async (b) => ({
    pollMinutes: (b as { minutes: number }).minutes,
  }),
) {
  return fakeHost(async (key, input) => {
    if (key === "issues.settings.get") return { pollMinutes: 5 };
    if (key === "issues.set_poll_interval") return save(input?.body);
    throw new Error(`unexpected ${key}`);
  });
}

const field = (c: HTMLElement) => byTestId(c, "backlog-poll-minutes") as HTMLInputElement;

describe("Sync interval in the settings (M1, US4.2)", () => {
  it("is a labelled minutes field with the default 5", async () => {
    const c = await mount(createPollInterval(setup()), { workspaceId: "ws-1" });
    expect(field(c).value).toBe("5");
    expect(c.querySelector(`label[for="${field(c).id}"]`)!.textContent).toBe(en.pollLabel);
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("refuses invalid minutes under the field (AC4.2.2)", async () => {
    const host = setup();
    const c = await mount(createPollInterval(host), { workspaceId: "ws-1" });
    for (const bad of ["0.5", "0", "-1", "abc", "", "1441"]) {
      await act(async () => setValue(field(c), bad));
      await act(async () => byTestId(c, "backlog-poll-save")!.click());
      const error = byTestId(c, "backlog-poll-error")!;
      expect(error.textContent).toBe(en.pollMinError);
      expect(field(c).getAttribute("aria-describedby")).toContain(error.id);
      expect(field(c).getAttribute("aria-invalid")).toBe("true");
    }
    expect(
      vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === "issues.set_poll_interval"),
    ).toHaveLength(0);
  });

  it("saves with Saving..., then says Saved", async () => {
    const pending = deferred<unknown>();
    const host = setup(() => pending.promise);
    const c = await mount(createPollInterval(host), { workspaceId: "ws-1" });
    await act(async () => setValue(field(c), "2"));
    await act(async () => byTestId(c, "backlog-poll-save")!.click());
    expect(byTestId(c, "backlog-poll-save")!.textContent).toBe(en.saving);
    expect(vi.mocked(host.api.invokeAction)).toHaveBeenLastCalledWith("issues.set_poll_interval", {
      workspaceId: "ws-1",
      body: { minutes: 2 },
    });
    await act(async () => pending.resolve({ pollMinutes: 2 }));
    expect(byTestId(c, "backlog-poll-message")!.textContent).toBe(en.pollSaved);
  });

  it("shows the server's refusal and other failures", async () => {
    let error: unknown = actionError(400, { code: "validation", field: "minutes" });
    const c = await mount(
      createPollInterval(
        setup(async () => {
          throw error;
        }),
      ),
      { workspaceId: "ws-1" },
    );
    await act(async () => byTestId(c, "backlog-poll-save")!.click());
    expect(byTestId(c, "backlog-poll-error")!.textContent).toBe(en.pollMinError);
    error = actionError(403, {});
    await act(async () => byTestId(c, "backlog-poll-save")!.click());
    expect(byTestId(c, "backlog-poll-message")!.textContent).toBe(en.adminOnly);
  });

  it("uses only catalogue text", async () => {
    const c = await mount(createPollInterval(setup(), pseudoCatalogue(en)), { workspaceId: "ws-1" });
    expectOnlyCatalogueText(c, (ok, msg) => expect(ok, msg).toBe(true));
  });
});
