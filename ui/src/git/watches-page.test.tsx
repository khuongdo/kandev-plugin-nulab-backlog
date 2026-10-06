import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  connected,
  deferred,
  expectOnlyCatalogueText,
  expectTestIds,
  fakeHost,
  mount,
  notConnected,
  pseudoCatalogue,
  unmount,
} from "../testing/harness";
import { createWatchesPage } from "./watches-page";

afterEach(unmount);

const WATCH = {
  id: "w1",
  name: "Reviews",
  projectKey: "PROJ",
  repoName: "web-app",
  statuses: ["open"],
  assignee: "anyone",
  creator: "anyone",
  workflowId: "wf-1",
  state: "active",
  createdCount: 10,
  pendingCount: 15,
};

type Handlers = Record<string, (body: unknown) => Promise<unknown>>;

function setup(watches: () => Promise<unknown>, handlers: Handlers = {}, view: unknown = connected) {
  return fakeHost(async (key, input) => {
    if (key === "connection.get") return view;
    if (key === "git.watches.list") return watches();
    if (key === "git.repositories.list") return { repositories: [] };
    const h = handlers[key];
    if (!h) throw new Error(`unexpected ${key}`);
    return h(input?.body);
  });
}

const calls = (host: ReturnType<typeof setup>, key: string) =>
  vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);

describe("PR watches page (M4, US6.1)", () => {
  it("shows the empty state with New watch", async () => {
    const c = await mount(createWatchesPage(setup(async () => ({ watches: [] }))), {});
    expect(byTestId(c, "backlog-watches-empty")!.textContent).toContain(en.watchesEmpty);
    expect(byTestId(c, "backlog-watch-new")!.textContent).toBe(en.newWatch);
    await act(async () => byTestId(c, "backlog-watch-new")!.click());
    expect(byTestId(c, "backlog-watch-form")).not.toBeNull();
  });

  it("shows loading, then a load error with Retry", async () => {
    const first = deferred<unknown>();
    let n = 0;
    const host = setup(() => (n++ === 0 ? first.promise : Promise.resolve({ watches: [] })));
    const c = await mount(createWatchesPage(host), {});
    expect(byTestId(c, "backlog-watches-loading")).not.toBeNull();
    await act(async () => first.reject(actionError(503, { code: "unreachable" })));
    expect(byTestId(c, "backlog-watches-error")!.textContent).toContain(en.unreachable);
    await act(async () => byTestId(c, "backlog-watches-retry")!.click());
    expect(byTestId(c, "backlog-watches-empty")).not.toBeNull();
  });

  it("shows the rate-limited notice with its wait", async () => {
    const c = await mount(
      createWatchesPage(
        setup(async () => {
          throw actionError(429, { code: "rate_limited", retryAfterSeconds: 12 });
        }),
      ),
      {},
    );
    expect(byTestId(c, "backlog-watches-error")!.textContent).toContain("Try again in 12 s");
  });

  it("shows Backlog is not connected", async () => {
    const c = await mount(createWatchesPage(setup(async () => ({ watches: [WATCH] }), {}, notConnected)), {});
    expect(byTestId(c, "backlog-watches-not-connected")!.textContent).toContain(en.pageNotConnected);
  });

  it("lists watches with status and progress, and runs, pauses and resumes them", async () => {
    const run = deferred<unknown>();
    const host = setup(
      async () => ({
        watches: [WATCH, { ...WATCH, id: "w2", name: "Paused one", state: "paused", pendingCount: 0 }],
      }),
      {
        "git.watches.run": () => run.promise,
        "git.watches.pause": async () => ({ ...WATCH, state: "paused" }),
        "git.watches.resume": async () => ({ ...WATCH, id: "w2", state: "active" }),
      },
    );
    const c = await mount(createWatchesPage(host), {});
    expect(byTestId(c, "backlog-watch-status-w1")!.textContent).toBe(en.watchActive);
    expect(byTestId(c, "backlog-watch-progress-w1")!.textContent).toBe(
      "Created 10/25 tasks, the rest in later cycles",
    );
    expect(byTestId(c, "backlog-watch-status-w2")!.textContent).toBe(en.watchPaused);
    const runButton = byTestId(c, "backlog-watch-run-w1") as HTMLButtonElement;
    await act(async () => runButton.click());
    expect(runButton.disabled).toBe(true);
    expect(runButton.textContent).toBe(en.running);
    await act(async () => run.resolve({ queued: true }));
    expect(runButton.disabled).toBe(false);
    expect(calls(host, "git.watches.run")[0]![1]).toEqual({ workspaceId: "ws-1", body: { id: "w1" } });
    await act(async () => byTestId(c, "backlog-watch-pause-w1")!.click());
    expect(byTestId(c, "backlog-watch-status-w1")!.textContent).toBe(en.watchPaused);
    await act(async () => byTestId(c, "backlog-watch-resume-w2")!.click());
    expect(byTestId(c, "backlog-watch-status-w2")!.textContent).toBe(en.watchActive);
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("offers no Resume or Run for a not-connected watch", async () => {
    const c = await mount(
      createWatchesPage(setup(async () => ({ watches: [{ ...WATCH, state: "not_connected" }] }))),
      {},
    );
    expect(byTestId(c, "backlog-watch-status-w1")!.textContent).toBe(en.watchNotConnected);
    expect(byTestId(c, "backlog-watch-resume-w1")).toBeNull();
    expect(byTestId(c, "backlog-watch-run-w1")).toBeNull();
  });

  it("asks before deleting, says tasks stay, with Cancel focused (AC6.1.5)", async () => {
    const del = vi.fn(async () => ({ ok: true }));
    const host = setup(async () => ({ watches: [WATCH] }), { "git.watches.delete": del });
    const c = await mount(createWatchesPage(host), {});
    await act(async () => byTestId(c, "backlog-watch-delete-w1")!.click());
    const dialog = byTestId(c, "backlog-watch-delete-dialog")!;
    expect(dialog.textContent).toContain(en.deleteWatchBody);
    expect(document.activeElement).toBe(byTestId(c, "backlog-watch-delete-dialog-cancel"));
    await act(async () => byTestId(c, "backlog-watch-delete-dialog-confirm")!.click());
    expect(del).toHaveBeenCalledWith({ id: "w1" });
    expect(byTestId(c, "backlog-watch-row-w1")).toBeNull();
  });

  it("renders only catalogue text", async () => {
    // The watch name is user data; it is wrapped so only plugin text is checked.
    const page = createWatchesPage(
      setup(async () => ({ watches: [{ ...WATCH, name: "⟦Reviews⟧" }] })),
      pseudoCatalogue(en),
    );
    const c = await mount(page, {});
    expectOnlyCatalogueText(c, (ok, msg) => expect(ok, msg).toBe(true));
  });
});
