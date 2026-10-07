import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  choose,
  expectTestIds,
  fakeHost,
  mount,
  optionValues,
  rawControls,
  text,
  unmount,
} from "../testing/harness";
import { createPrList } from "./pr-list";

afterEach(unmount);

const PROVIDERS = [
  {
    provider: "github",
    state: "connected",
    account: "Lan",
    mappings: [
      { projectKey: "PROJ", repos: ["acme/web", "acme/api"] },
      { projectKey: "DEMO", repos: ["acme/demo"] },
    ],
  },
  { provider: "gitlab", state: "not_configured", mappings: [] },
  { provider: "bitbucket", state: "error", lastError: "invalid_token", mappings: [] },
];

const ROW = {
  provider: "github",
  repo: "acme/web",
  number: 42,
  title: "PROJ-1 Add login",
  state: "draft",
  author: "lan-dev",
  sourceBranch: "feature/PROJ-1",
  targetBranch: "main",
  updatedAt: "2026-10-06T09:00:00Z",
  url: "https://github.com/acme/web/pull/42",
  linkedTaskIds: ["task-9"],
};

type Handler = (body: Record<string, unknown>) => Promise<unknown>;

function setup(handlers: Record<string, Handler> = {}) {
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    switch (key) {
      case "scm.providers.list":
        return { providers: PROVIDERS };
      case "scm.queries.list":
        return { queries: [] };
      case "scm.prs.list":
        return { items: [ROW], page: body.page, pageSize: 20, hasNext: false };
      case "git.repositories.list":
        return { repositories: [{ ownerOrProject: "PROJ", repositoryName: "web-app", repositoryId: "11" }] };
      case "git.prs.list":
        return { items: [], page: 1, pageSize: 20, total: 0, hasNext: false };
    }
    throw new Error(`unexpected ${key}`);
  });
}

const calls = (host: ReturnType<typeof setup>, key: string) =>
  vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);

const QUICK = [
  { id: "review", label: "Review", hint: "Read", icon: "eye", promptTemplate: "Review {{url}}" },
];

async function render(host: ReturnType<typeof setup>) {
  return mount(createPrList(host, en), {
    workspaceId: "ws-1",
    selectedProjects: ["PROJ"],
    selection: { key: "preset:0" },
    quickActions: QUICK,
  });
}

async function pickGitHub(c: HTMLElement) {
  await act(async () => choose(c, "backlog-prs-provider", "github"));
}

describe("Pull requests of every provider (FR4.1, FR1.3)", () => {
  it("offers Backlog Git and the connected providers; Backlog stays the default", async () => {
    const host = setup();
    const c = await render(host);
    expect(optionValues(c, "backlog-prs-provider")).toEqual(["backlog", "github"]);
    expect(calls(host, "git.prs.list")).toHaveLength(1);
    expect(calls(host, "git.prs.list")[0]![1]).toMatchObject({
      body: { projectKey: "PROJ", repoName: "web-app", statuses: ["open"] },
    });
    expect(calls(host, "scm.prs.list")).toHaveLength(0);
  });

  it("lists a provider's mapped repositories of selected projects and its pull requests", async () => {
    const host = setup();
    const c = await render(host);
    await pickGitHub(c);
    const repo = byTestId(c, "backlog-scm-prs-repo")!;
    const options = [...repo.querySelectorAll("button")].map((b) => b.getAttribute("data-testid"));
    expect(options).toEqual([
      "backlog-scm-prs-repo-option-PROJ:acme/web",
      "backlog-scm-prs-repo-option-PROJ:acme/api",
    ]); // FR3.3, FR6.2: DEMO is not selected, so its repository is hidden
    expect(calls(host, "scm.prs.list")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: {
        provider: "github",
        projectKey: "PROJ",
        repo: "acme/web",
        statuses: ["open"],
        author: "anyone",
        page: 1,
      },
    });
    const row = byTestId(c, "backlog-scm-pr-row-42")!;
    expect(byTestId(row, "backlog-scm-pr-row-42-link")!.getAttribute("href")).toBe(ROW.url);
    expect(text(row)).toContain("#42");
    expect(text(row)).toContain("PROJ-1 Add login");
    expect(text(row)).toContain(en.stateDraft);
    expect(text(row)).toContain("by lan-dev");
    expect(text(row)).toContain("feature/PROJ-1 → main");
    expect(text(row)).toContain("updated rel:2026-10-06T09:00:00Z");
    expect(byTestId(row, "backlog-scm-pr-task-42-task-9")).not.toBeNull();
  });

  it("filters by status and author and pages", async () => {
    const host = setup({
      "scm.prs.list": async (b) => ({ items: [ROW], page: b.page, pageSize: 20, hasNext: b.page === 1 }),
    });
    const c = await render(host);
    await pickGitHub(c);
    await act(async () => byTestId(c, "backlog-scm-prs-status-merged")!.click());
    await act(async () => choose(c, "backlog-scm-prs-author", "me"));
    await act(async () => byTestId(c, "backlog-scm-prs-next")!.click());
    expect(calls(host, "scm.prs.list").at(-1)![1]).toMatchObject({
      body: { statuses: ["open", "merged"], author: "me", page: 2 },
    });
    expect((byTestId(c, "backlog-scm-prs-next") as HTMLButtonElement).disabled).toBe(true);
    await act(async () => byTestId(c, "backlog-scm-prs-prev")!.click());
    expect(calls(host, "scm.prs.list").at(-1)![1]).toMatchObject({ body: { page: 1 } });
  });

  it("opens on the provider's default saved query and saves the filters with the provider (FR4.2)", async () => {
    const save = vi.fn(async (b: Record<string, unknown>) => ({ id: "q9", ...b }));
    const host = setup({
      "scm.queries.list": async () => ({
        queries: [
          {
            id: "q1",
            name: "Mine",
            provider: "github",
            projectKey: "PROJ",
            repo: "acme/api",
            statuses: ["merged"],
            author: "me",
            isDefault: true,
          },
          {
            id: "q2",
            name: "GitLab one",
            provider: "gitlab",
            projectKey: "PROJ",
            repo: "g/p",
            statuses: ["open"],
          },
        ],
      }),
      "scm.queries.save": save,
    });
    const c = await render(host);
    await pickGitHub(c);
    expect(calls(host, "scm.prs.list")[0]![1]).toMatchObject({
      body: { repo: "acme/api", statuses: ["merged"], author: "me" },
    });
    expect(optionValues(c, "backlog-scm-prs-saved")).toEqual(["q1"]);
    await act(async () => byTestId(c, "backlog-scm-prs-save-query")!.click());
    await act(async () => {
      const input = byTestId(c, "backlog-save-query-name") as HTMLInputElement;
      const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value")!.set!;
      setter.call(input, "Merged API");
      input.dispatchEvent(new Event("input", { bubbles: true }));
    });
    await act(async () => byTestId(c, "backlog-save-query-save")!.click());
    expect(save).toHaveBeenCalledWith({
      name: "Merged API",
      provider: "github",
      projectKey: "PROJ",
      repo: "acme/api",
      statuses: ["merged"],
      author: "me",
    });
  });

  it("links a task created from a row to the pull request (FR5.1)", async () => {
    const host = setup({ "scm.prs.link": async () => ({ taskId: "t-1" }) });
    const c = await render(host);
    await pickGitHub(c);
    await act(async () => byTestId(c, "backlog-scm-pr-42-start-item")!.click());
    await act(async () => byTestId(c, "fake-task-create-confirm")!.click());
    expect(calls(host, "scm.prs.link")[0]![1]).toEqual({
      workspaceId: "ws-1",
      taskId: "t-1",
      body: { url: ROW.url },
    });
    expect(byTestId(c, "backlog-scm-pr-task-42-t-1")).not.toBeNull();
  });

  it("shows a provider failure with Retry, in provider words", async () => {
    let fail = true;
    const host = setup({
      "scm.prs.list": async () => {
        if (fail) throw actionError(429, { code: "rate_limited", retryAfterSeconds: 9 });
        return { items: [], page: 1, pageSize: 20, hasNext: false };
      },
    });
    const c = await render(host);
    await pickGitHub(c);
    expect(byTestId(c, "backlog-scm-prs-error")!.textContent).toContain(
      "The provider is limiting requests. Try again in 9 s",
    );
    fail = false;
    await act(async () => byTestId(c, "backlog-scm-prs-retry")!.click());
    expect(byTestId(c, "backlog-scm-prs-empty")).not.toBeNull();
  });

  it("falls back to Backlog Git only when the providers cannot be read", async () => {
    const host = setup({
      "scm.providers.list": async () => {
        throw new Error("down");
      },
    });
    const c = await render(host);
    expect(optionValues(c, "backlog-prs-provider")).toEqual(["backlog"]);
  });

  it("uses host controls, has test ids and passes axe", async () => {
    const c = await render(setup());
    await pickGitHub(c);
    expect(rawControls(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
    expect(await axeViolations(c)).toEqual([]);
  });
});
