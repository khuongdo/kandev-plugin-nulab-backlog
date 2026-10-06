import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  expectTestIds,
  fakeHost,
  mount,
  selectValue,
  setValue,
  unmount,
} from "../testing/harness";
import { createDashboardPage } from "./dashboard-page";

afterEach(unmount);

const QUERY = {
  id: "q1",
  name: "Open PRs",
  projectKey: "PROJ",
  repoName: "web-app",
  statuses: ["open"],
  assignee: "anyone",
};
const ROW = {
  number: 42,
  title: "Add login page",
  state: "open",
  assignee: "Lan",
  repoName: "web-app",
  url: "https://example-space.backlog.com/git/PROJ/web-app/pullRequests/42",
  linkedTaskIds: ["task-17"],
};

type Handlers = Record<string, (body: unknown) => Promise<unknown>>;

function setup(handlers: Handlers = {}) {
  return fakeHost(async (key, input) => {
    if (key === "git.queries.list") return { queries: [QUERY] };
    if (key === "git.repositories.list")
      return { repositories: [{ ownerOrProject: "PROJ", repositoryName: "web-app" }] };
    const h = handlers[key];
    if (!h) throw new Error(`unexpected ${key}`);
    return h(input?.body);
  });
}

async function choose(c: HTMLElement) {
  await act(async () => selectValue(byTestId(c, "backlog-query-select") as HTMLSelectElement, "q1"));
}

describe("PR dashboard (M5, US6.3)", () => {
  it("runs the chosen query and lists PRs with state and linked tasks (AC6.3.1)", async () => {
    const run = vi.fn(async () => ({ rows: [ROW] }));
    const c = await mount(createDashboardPage(setup({ "git.queries.run": run })), {});
    expect(
      [...(byTestId(c, "backlog-query-select") as HTMLSelectElement).options].map((o) => o.text),
    ).toEqual([en.chooseQuery, "Open PRs"]);
    await choose(c);
    expect(run).toHaveBeenCalledWith({ id: "q1" });
    const row = byTestId(c, "backlog-query-row-42")!;
    expect(row.textContent).toContain("Add login page");
    expect(row.textContent).toContain(en.stateOpen);
    expect(byTestId(c, "backlog-query-task-task-17")!.getAttribute("href")).toBe("/t/task-17");
    expect(byTestId(c, "backlog-query-pr-42")!.getAttribute("href")).toBe(ROW.url);
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("shows a hint for no results", async () => {
    const c = await mount(createDashboardPage(setup({ "git.queries.run": async () => ({ rows: [] }) })), {});
    await choose(c);
    expect(byTestId(c, "backlog-query-empty")!.textContent).toBe(en.queryEmpty);
  });

  it("shows an error or a rate limit with Retry (AC6.3.2)", async () => {
    let n = 0;
    const run = vi.fn(async () => {
      if (n++ === 0) throw actionError(429, { code: "rate_limited", retryAfterSeconds: 5 });
      return { rows: [ROW] };
    });
    const c = await mount(createDashboardPage(setup({ "git.queries.run": run })), {});
    await choose(c);
    expect(byTestId(c, "backlog-query-error")!.textContent).toContain("Try again in 5 s");
    await act(async () => byTestId(c, "backlog-query-retry")!.click());
    expect(byTestId(c, "backlog-query-row-42")).not.toBeNull();
  });

  it("validates the save-query form and saves", async () => {
    const save = vi.fn(async (b: unknown) => ({ id: "q2", ...(b as object) }));
    const c = await mount(createDashboardPage(setup({ "git.queries.save": save })), {});
    await act(async () => byTestId(c, "backlog-query-new")!.click());
    await act(async () => byTestId(c, "backlog-query-save")!.click());
    expect(byTestId(c, "backlog-query-name-error")!.textContent).toBe(en.errorName);
    await act(async () => setValue(byTestId(c, "backlog-query-name") as HTMLInputElement, "Mine"));
    await act(async () => byTestId(c, "backlog-query-save")!.click());
    expect(byTestId(c, "backlog-query-repo-error")!.textContent).toBe(en.errorRepository);
    await act(async () =>
      selectValue(byTestId(c, "backlog-query-repo") as HTMLSelectElement, "PROJ/web-app"),
    );
    await act(async () => byTestId(c, "backlog-query-save")!.click());
    expect(save).toHaveBeenCalledWith({
      name: "Mine",
      projectKey: "PROJ",
      repoName: "web-app",
      statuses: ["open"],
      assignee: "anyone",
    });
  });

  it("asks before deleting a query", async () => {
    const del = vi.fn(async () => ({ ok: true }));
    const c = await mount(
      createDashboardPage(
        setup({ "git.queries.run": async () => ({ rows: [] }), "git.queries.delete": del }),
      ),
      {},
    );
    await choose(c);
    await act(async () => byTestId(c, "backlog-query-delete")!.click());
    expect(byTestId(c, "backlog-query-delete-dialog")!.textContent).toContain(en.deleteQueryTitle);
    await act(async () => byTestId(c, "backlog-query-delete-dialog-confirm")!.click());
    expect(del).toHaveBeenCalledWith({ id: "q1" });
  });
});
