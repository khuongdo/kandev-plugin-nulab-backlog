import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  deferred,
  expectTestIds,
  fakeHost,
  mount,
  setValue,
  text,
  unmount,
} from "../testing/harness";
import { createGitAccess } from "./git-access";

afterEach(unmount);

const PW = ["pw", "for", "test"].join("-");

function setup(save: (body: unknown) => Promise<unknown> = async () => ({ hasGitCredential: true })) {
  const host = fakeHost(async (key, input) => {
    if (key === "connection.set_git_credential") return save(input?.body);
    throw new Error(`unexpected ${key}`);
  });
  return host;
}

async function render(host: ReturnType<typeof setup>, props: Record<string, unknown> = {}) {
  const announce = vi.fn();
  const c = await mount(createGitAccess(host), {
    workspaceId: "ws-1",
    hasGitCredential: false,
    announce,
    ...props,
  });
  return { c, announce };
}

async function fill(c: HTMLElement) {
  await act(async () => {
    setValue(byTestId(c, "backlog-git-username") as HTMLInputElement, "lan");
    setValue(byTestId(c, "backlog-git-password") as HTMLInputElement, PW);
  });
}

describe("Git access block (M1, US5.5)", () => {
  it("is a collapsible block with labelled fields", async () => {
    const { c } = await render(setup());
    const details = c.querySelector("details")!;
    expect(details.querySelector("summary")!.textContent).toBe(en.gitAccessHeading);
    expect(c.querySelector('label[for="backlog-git-username"]')!.textContent).toBe(en.gitUsernameLabel);
    expect(c.querySelector('label[for="backlog-git-password"]')!.textContent).toBe(en.gitPasswordLabel);
    expect((byTestId(c, "backlog-git-password") as HTMLInputElement).type).toBe("password");
    expect(byTestId(c, "backlog-git-status")!.textContent).toBe(en.gitNotStored);
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("saves, clears the password and shows Saved and Stored (AC5.5.1)", async () => {
    const save = vi.fn(async () => ({ connected: true, hasGitCredential: true }));
    const host = setup(save);
    const { c, announce } = await render(host);
    await fill(c);
    await act(async () => byTestId(c, "backlog-git-save")!.click());
    expect(save).toHaveBeenCalledWith({ gitUsername: "lan", gitPassword: PW });
    expect(vi.mocked(host.api.invokeAction).mock.calls[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { gitUsername: "lan", gitPassword: PW },
    });
    expect((byTestId(c, "backlog-git-password") as HTMLInputElement).value).toBe("");
    expect(byTestId(c, "backlog-git-message")!.textContent).toBe(en.gitSaved);
    expect(byTestId(c, "backlog-git-status")!.textContent).toBe(en.gitStored);
    expect(announce).toHaveBeenCalledWith({ key: "gitSaved" });
    expect(text(c)).not.toContain(PW);
  });

  it("shows Stored for a stored credential, never the password", async () => {
    const { c } = await render(setup(), { hasGitCredential: true });
    expect(byTestId(c, "backlog-git-status")!.textContent).toBe(en.gitStored);
    expect((byTestId(c, "backlog-git-password") as HTMLInputElement).value).toBe("");
  });

  it("disables Save while saving", async () => {
    const pending = deferred<unknown>();
    const { c } = await render(setup(() => pending.promise));
    await fill(c);
    const button = byTestId(c, "backlog-git-save") as HTMLButtonElement;
    await act(async () => button.click());
    expect(button.textContent).toBe(en.saving);
    expect(button.disabled).toBe(true);
    await act(async () => pending.resolve({ hasGitCredential: true }));
    expect(button.disabled).toBe(false);
  });

  it("shows a field error for a rejected field", async () => {
    const { c } = await render(
      setup(async () => {
        throw actionError(400, { code: "validation", field: "gitPassword" });
      }),
    );
    await fill(c);
    await act(async () => byTestId(c, "backlog-git-save")!.click());
    expect(byTestId(c, "backlog-git-password-error")!.textContent).toBe(en.errorGitPassword);
    expect(byTestId(c, "backlog-git-password")!.getAttribute("aria-invalid")).toBe("true");
  });

  it("announces a Git authentication error from Test connection once (AC5.5.2)", async () => {
    const host = setup();
    const announce = vi.fn();
    const props = { workspaceId: "ws-1", hasGitCredential: true, announce, gitCheck: "invalid" };
    const c = await mount(createGitAccess(host), props);
    expect(byTestId(c, "backlog-git-invalid")!.textContent).toBe(en.gitInvalid);
    expect(announce).toHaveBeenCalledTimes(1);
    expect(announce).toHaveBeenCalledWith({ key: "gitInvalid" });
  });
});
