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
  choose,
  mount,
  optionValues,
  setValue,
  unmount,
} from "../testing/harness";
import { createWatchForm } from "./watch-form";

afterEach(unmount);

const REPOS = {
  repositories: [
    { ownerOrProject: "PROJ", repositoryName: "web-app", repositoryId: "11" },
    { ownerOrProject: "PROJ", repositoryName: "api", repositoryId: "12" },
  ],
};

function setup(save: (body: unknown) => Promise<unknown> = async (b) => ({ id: "w1", ...(b as object) })) {
  const host = fakeHost(async (key, input) => {
    if (key === "git.repositories.list") return REPOS;
    if (key === "git.watches.save") return save(input?.body);
    throw new Error(`unexpected ${key}`);
  });
  return host;
}

async function render(host: ReturnType<typeof setup>, watch?: Record<string, unknown>) {
  const onSaved = vi.fn();
  const onCancel = vi.fn();
  const c = await mount(createWatchForm(host), { workspaceId: "ws-1", watch, onSaved, onCancel });
  return { c, onSaved, onCancel };
}

async function fill(c: HTMLElement) {
  await act(async () => {
    setValue(byTestId(c, "backlog-watch-name") as HTMLInputElement, "Reviews");
    choose(c, "backlog-watch-repo", "PROJ/web-app");
  });
}

describe("Watch form (M4, US6.1)", () => {
  it("has labelled fields, repositories of selected projects and a status fieldset", async () => {
    const { c } = await render(setup());
    expect(c.querySelector('label[for="backlog-watch-name"]')!.textContent).toBe(`${en.watchNameLabel} *`);
    expect(optionValues(c, "backlog-watch-repo")).toEqual(["PROJ/web-app", "PROJ/api"]);
    const fieldset = c.querySelector("fieldset")!;
    expect(fieldset.querySelector("legend")!.textContent).toBe(en.watchStatusLegend);
    expect(byTestId(c, "backlog-watch-status-open")!.getAttribute("aria-checked")).toBe("true");
    expect(byTestId(c, "backlog-watch-save")!.getAttribute("data-variant")).toBe("default");
    expect(byTestId(c, "backlog-watch-cancel")!.getAttribute("data-variant")).toBe("outline");
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("moves focus to an empty name and saves nothing (AC6.1.2)", async () => {
    const save = vi.fn();
    const { c } = await render(setup(save));
    await act(async () => byTestId(c, "backlog-watch-save")!.click());
    expect(byTestId(c, "backlog-watch-name-error")!.textContent).toBe(en.errorName);
    expect(document.activeElement).toBe(byTestId(c, "backlog-watch-name"));
    expect(save).not.toHaveBeenCalled();
  });

  it("shows a server validation error on its field and focuses it", async () => {
    const { c } = await render(
      setup(async () => {
        throw actionError(400, { code: "validation", field: "repository" });
      }),
    );
    await fill(c);
    await act(async () => byTestId(c, "backlog-watch-save")!.click());
    expect(byTestId(c, "backlog-watch-repo-error")!.textContent).toBe(en.errorRepository);
    expect(document.activeElement).toBe(byTestId(c, "backlog-watch-repo"));
  });

  it("saves with the workflow from the task creation context", async () => {
    const pending = deferred<unknown>();
    const save = vi.fn(() => pending.promise);
    const { c, onSaved } = await render(setup(save));
    await fill(c);
    await act(async () => {
      byTestId(c, "backlog-watch-status-merged")!.click();
      choose(c, "backlog-watch-assignee", "me");
      setValue(byTestId(c, "backlog-watch-issue") as HTMLInputElement, " PROJ-120 ");
    });
    const button = byTestId(c, "backlog-watch-save") as HTMLButtonElement;
    await act(async () => button.click());
    expect(button.disabled).toBe(true);
    expect(button.textContent).toBe(en.saving);
    expect(save).toHaveBeenCalledWith({
      name: "Reviews",
      projectKey: "PROJ",
      repoName: "web-app",
      statuses: ["open", "merged"],
      assignee: "me",
      creator: "anyone",
      issueKey: "PROJ-120",
      workflowId: "wf-1",
      workflowStepId: "step-1",
    });
    await act(async () => pending.resolve({ id: "w9" }));
    expect(onSaved).toHaveBeenCalledWith({ id: "w9" });
  });

  it("edits a watch with its id and Cancel calls back", async () => {
    const save = vi.fn(async () => ({ id: "w1" }));
    const watch = {
      id: "w1",
      name: "Old",
      projectKey: "PROJ",
      repoName: "api",
      statuses: ["closed"],
      assignee: "anyone",
      creator: "me",
    };
    const { c, onCancel } = await render(setup(save), watch);
    expect((byTestId(c, "backlog-watch-name") as HTMLInputElement).value).toBe("Old");
    await act(async () => byTestId(c, "backlog-watch-save")!.click());
    expect(save).toHaveBeenCalledWith(
      expect.objectContaining({ id: "w1", repoName: "api", statuses: ["closed"], creator: "me" }),
    );
    await act(async () => byTestId(c, "backlog-watch-cancel")!.click());
    expect(onCancel).toHaveBeenCalled();
  });
});

describe("Watch form for other providers (FR4.3)", () => {
  const PROVIDERS = {
    providers: [
      { provider: "github", state: "connected", mappings: [{ projectKey: "PROJ", repos: ["acme/web"] }] },
      { provider: "gitlab", state: "not_configured", mappings: [] },
      { provider: "bitbucket", state: "connected", mappings: [{ projectKey: "DEMO", repos: ["ws/old"] }] },
    ],
  };

  function scmSetup(save = vi.fn(async (b: unknown) => ({ id: "s1", ...(b as object) }))) {
    const host = fakeHost(async (key, input) => {
      if (key === "git.repositories.list") return REPOS;
      if (key === "scm.providers.list") return PROVIDERS;
      if (key === "scm.watches.save") return save(input?.body);
      throw new Error(`unexpected ${key}`);
    });
    return { host, save };
  }

  const renderScm = (host: ReturnType<typeof fakeHost>, watch?: Record<string, unknown>) =>
    mount(createWatchForm(host), {
      workspaceId: "ws-1",
      watch,
      selectedProjects: ["PROJ"],
      onSaved: vi.fn(),
      onCancel: vi.fn(),
    });

  it("chooses the provider; Backlog Git keeps its own form", async () => {
    const { host } = scmSetup();
    const c = await renderScm(host);
    expect(optionValues(c, "backlog-watch-provider")).toEqual(["backlog", "github", "bitbucket"]);
    expect(byTestId(c, "backlog-watch-issue"), "Backlog Git form by default").not.toBeNull();
    await act(async () => choose(c, "backlog-watch-provider", "github"));
    expect(byTestId(c, "backlog-watch-issue")).toBeNull();
    expect(optionValues(c, "backlog-watch-repo")).toEqual(["PROJ:acme/web"]);
    expect((byTestId(c, "backlog-watch-interval") as HTMLInputElement).value).toBe("5");
    expect(await axeViolations(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
  });

  it("saves with the provider, author, interval and workflow", async () => {
    const { host, save } = scmSetup();
    const c = await renderScm(host);
    await act(async () => choose(c, "backlog-watch-provider", "github"));
    await act(async () => {
      setValue(byTestId(c, "backlog-watch-name") as HTMLInputElement, "GH reviews");
      choose(c, "backlog-watch-repo", "PROJ:acme/web");
      choose(c, "backlog-watch-author", "me");
      setValue(byTestId(c, "backlog-watch-interval") as HTMLInputElement, "15");
    });
    await act(async () => byTestId(c, "backlog-watch-save")!.click());
    expect(save).toHaveBeenCalledWith({
      name: "GH reviews",
      provider: "github",
      projectKey: "PROJ",
      repo: "acme/web",
      statuses: ["open"],
      author: "me",
      intervalMinutes: 15,
      workflowId: "wf-1",
      workflowStepId: "step-1",
    });
  });

  it("checks the interval and shows server field errors", async () => {
    const { host } = scmSetup(
      vi.fn(async () => {
        throw actionError(400, { code: "validation", field: "repository" });
      }),
    );
    const c = await renderScm(host);
    await act(async () => choose(c, "backlog-watch-provider", "github"));
    await act(async () => {
      setValue(byTestId(c, "backlog-watch-name") as HTMLInputElement, "W");
      choose(c, "backlog-watch-repo", "PROJ:acme/web");
      setValue(byTestId(c, "backlog-watch-interval") as HTMLInputElement, "0");
    });
    await act(async () => byTestId(c, "backlog-watch-save")!.click());
    expect(byTestId(c, "backlog-watch-interval-error")!.textContent).toBe(en.errorInterval);
    await act(async () => setValue(byTestId(c, "backlog-watch-interval") as HTMLInputElement, "5"));
    await act(async () => byTestId(c, "backlog-watch-save")!.click());
    expect(byTestId(c, "backlog-watch-repo-error")!.textContent).toBe(en.errorRepository);
  });

  it("edits a provider watch with its provider fixed", async () => {
    const { host, save } = scmSetup();
    const c = await renderScm(host, {
      id: "s1",
      name: "Old",
      provider: "bitbucket",
      projectKey: "DEMO",
      repo: "ws/old",
      statuses: ["merged"],
      author: "anyone",
      intervalMinutes: 30,
    });
    expect(byTestId(c, "backlog-watch-provider")!.hasAttribute("disabled")).toBe(true);
    expect(optionValues(c, "backlog-watch-repo")).toContain("DEMO:ws/old");
    await act(async () => byTestId(c, "backlog-watch-save")!.click());
    expect(save).toHaveBeenCalledWith(
      expect.objectContaining({ id: "s1", provider: "bitbucket", repo: "ws/old", intervalMinutes: 30 }),
    );
  });
});
