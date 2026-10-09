import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en, format } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  choose,
  expectErrorAlert,
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

function setup(handlers: Record<string, Handler> = {}, active = "") {
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    switch (key) {
      case "scm.providers.list":
        return { providers: PROVIDERS, active };
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
    const options = [...repo.querySelectorAll('[data-host="IntegrationRepositoryFilterOption"]')].map((b) =>
      b.getAttribute("data-testid"),
    );
    expect(options).toEqual([
      "backlog-scm-prs-repo-option-all",
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
    // FR5.1, BR1.1-BR1.4: the host task indicator, with the task id as fallback title.
    const task = byTestId(row, "backlog-scm-pr-task-42-single")!;
    expect(task.getAttribute("data-host")).toBe("TaskRowIndicator");
    expect(task.textContent).toBe("task-9");
    await act(async () => task.click());
    expect(window.location.pathname).toBe("/t/task-9");
  });

  it("filters by status and author and pages", async () => {
    const host = setup({
      "scm.prs.list": async (b) => ({ items: [ROW], page: b.page, pageSize: 20, hasNext: b.page === 1 }),
    });
    const c = await render(host);
    await pickGitHub(c);
    await act(async () => byTestId(c, "backlog-scm-prs-status")!.click());
    await act(async () => byTestId(c, "backlog-scm-prs-status-merged")!.click());
    await act(async () => byTestId(c, "backlog-scm-prs-author-option-me")!.click());
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
    const saved = byTestId(c, "backlog-scm-prs-saved")!;
    expect(saved.getAttribute("data-value")).toBe("q1");
    expect(byTestId(saved, "backlog-scm-prs-saved-option-q1")!.textContent).toBe("Mine");
    expect(byTestId(saved, "backlog-scm-prs-saved-option-q2")).toBeNull(); // another provider's query
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
    expect(byTestId(c, "backlog-scm-pr-task-42-multi")!.textContent).toBe("Tasks 2");
    expect(byTestId(c, "backlog-scm-pr-task-42-item-t-1")).not.toBeNull();
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
    const alert = expectErrorAlert(byTestId(c, "backlog-scm-prs-error")); // intent 261009, FR2.2
    expect(alert.textContent).toContain("The provider is limiting requests. Try again in 9 s");
    expect(alert.contains(byTestId(c, "backlog-scm-prs-retry"))).toBe(true);
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

  it("lays the provider list out like GitHub: label-less dropdowns, Status (n), provider first (FR5.2, BR5.1-BR5.4, NFR3)", async () => {
    const host = setup();
    const c = await render(host);
    // The provider selector: label-less, named, first in the Backlog list's filter slot.
    const backlogSlot = byTestId(c, "backlog-prs-filters")!;
    expect(
      backlogSlot.firstElementChild!.querySelector('[data-testid="backlog-prs-provider"]'),
    ).not.toBeNull();
    expect(byTestId(c, "backlog-prs-toolbar")!.querySelector("label")).toBeNull();
    const provider = byTestId(c, "backlog-prs-provider")!;
    expect(provider.getAttribute("aria-label")).toBe(en.scmProviderLabel);
    expect(provider.className).toContain("md:w-[220px]");
    await pickGitHub(c);
    const toolbar = byTestId(c, "backlog-scm-prs-toolbar")!;
    expect(toolbar.querySelector("input")).toBeNull(); // no query box
    expect(toolbar.querySelector("label")).toBeNull();
    expect(toolbar.textContent).not.toContain(en.colRepository);
    const slot = byTestId(c, "backlog-scm-prs-filters")!;
    expect(slot.parentElement).toBe(toolbar);
    expect(slot.firstElementChild!.querySelector('[data-testid="backlog-prs-provider"]')).not.toBeNull();
    for (const [id, label] of [
      ["repo", en.colRepository],
      ["saved", en.scmSavedLabel],
      ["author", en.scmAuthorLabel],
    ] as const) {
      const filter = byTestId(c, `backlog-scm-prs-${id}`)!;
      expect(filter.getAttribute("data-host")).toBe("IntegrationRepositoryFilter");
      expect(filter.getAttribute("aria-label")).toBe(label);
      expect(filter.getAttribute("data-trigger-class")).toContain("md:w-[220px]");
      expect(filter.getAttribute("data-popover-class")).toContain("md:min-w-[360px]");
    }
    expect(byTestId(c, "backlog-scm-prs-saved-option-all")!.textContent).toBe(en.scmSavedNone);
    expect(byTestId(c, "backlog-scm-prs-author-option-all")!.textContent).toBe(en.whoAnyone);
    expect(byTestId(c, "backlog-scm-prs-author")!.getAttribute("data-value")).toBe(""); // anyone
    // Status: the same "Status (n)" popover; the last picked status cannot be cleared.
    const status = byTestId(c, "backlog-scm-prs-status")!;
    expect(status.textContent).toBe("Status (1)");
    expect(status.getAttribute("aria-haspopup")).toBe("dialog");
    expect(byTestId(c, "backlog-scm-prs-status-open")).toBeNull();
    await act(async () => status.click());
    const open = byTestId(c, "backlog-scm-prs-status-open") as HTMLButtonElement;
    expect(open.getAttribute("aria-disabled")).toBe("true");
    expect(open.disabled).toBe(false);
    const listed = calls(host, "scm.prs.list").length;
    await act(async () => open.click());
    expect(open.getAttribute("aria-checked")).toBe("true");
    expect(calls(host, "scm.prs.list")).toHaveLength(listed);
    await act(async () => byTestId(c, "backlog-scm-prs-author-option-me")!.click());
    await act(async () => byTestId(c, "backlog-scm-prs-author-option-all")!.click()); // "" is Anyone
    expect(calls(host, "scm.prs.list").at(-1)![1]).toMatchObject({ body: { author: "anyone", page: 1 } });
    expect(await axeViolations(c)).toEqual([]);
  });

  it("uses host controls, has test ids and passes axe", async () => {
    const c = await render(setup());
    await pickGitHub(c);
    expect(rawControls(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
    expect(await axeViolations(c)).toEqual([]);
  });
});

describe("Only the active source control service (intent 261008-source-control-settings)", () => {
  it("offers no provider selector while Backlog Git is active (FR1.5)", async () => {
    const host = setup({}, "backlog_git");
    const c = await render(host);
    expect(byTestId(c, "backlog-prs-provider")).toBeNull();
    expect(byTestId(c, "backlog-prs-filters")).not.toBeNull();
    expect(calls(host, "scm.prs.list")).toHaveLength(0);
  });

  it("lists only the active provider's pull requests, with no selector (FR1.5)", async () => {
    const host = setup({}, "github");
    const c = await render(host);
    expect(byTestId(c, "backlog-prs-provider")).toBeNull();
    expect(byTestId(c, "backlog-scm-prs-toolbar")).not.toBeNull();
  });

  it("names the active service when Backlog Git is refused (FR1.4)", async () => {
    const host = setup({
      "git.prs.list": async () => {
        throw actionError(409, { code: "service_inactive", activeService: "github" });
      },
    });
    const c = await render(host);
    const alert = expectErrorAlert(byTestId(c, "backlog-prs-error")); // intent 261009, FR2.2
    expect(alert.textContent).toContain(format(en.scmServiceInactive, { service: "GitHub" }));
    expect(alert.contains(byTestId(c, "backlog-prs-retry"))).toBe(true);
  });
});
