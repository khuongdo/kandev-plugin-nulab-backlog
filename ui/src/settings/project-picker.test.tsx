import { act } from "react";
import axe from "axe-core";
import { afterEach, describe, expect, it, vi } from "vitest";

import { createProjectPicker } from "./project-picker";
import { en } from "../messages/en";
import { actionError, byTestId, fakeHost, mount, setValue, text, unmount } from "../testing/harness";

afterEach(() => {
  unmount();
  vi.useRealTimers();
});

const PROJECTS = [
  { projectKey: "PROJ", projectId: 101, projectName: "Test Project", selected: true },
  { projectKey: "DEMO", projectId: 102, projectName: "Demo Project", selected: false },
];

function scripted(
  list: () => Promise<unknown>,
  save = vi.fn<(body: unknown) => Promise<unknown>>(async () => ({})),
) {
  const host = fakeHost(async (key, input) => {
    if (key === "connection.list_projects") return list();
    if (key === "git.impact") return { prLinks: 3, prWatches: 2 };
    if (key === "issues.impact") return { issueLinks: 4 };
    if (key === "connection.set_projects") return save(input?.body);
    throw new Error(`unexpected ${key}`);
  });
  return { host, save };
}

async function render(host: ReturnType<typeof fakeHost>, announce = vi.fn()) {
  const c = await mount(createProjectPicker(host), { workspaceId: "ws-1", announce });
  return { c, announce };
}

const box = (c: HTMLElement, key: string) => byTestId(c, `backlog-project-${key}`) as HTMLButtonElement;
const checked = (c: HTMLElement, key: string) => box(c, key).getAttribute("aria-checked") === "true";

describe("Project picker (US1.7, US1.9)", () => {
  it("lists projects as checkboxes in a fieldset with a legend, and filters them", async () => {
    const { host } = scripted(async () => ({ projects: PROJECTS }));
    const { c } = await render(host);
    const fieldset = c.querySelector("fieldset")!;
    expect(fieldset.querySelector("legend")!.textContent).toBe(en.projectsLegend);
    expect(box(c, "PROJ").getAttribute("data-host")).toBe("Checkbox");
    expect(checked(c, "PROJ")).toBe(true);
    expect(checked(c, "DEMO")).toBe(false);
    expect(c.querySelector('label[for="backlog-project-PROJ"]')!.textContent).toBe("Test Project (PROJ)");
    await act(async () => setValue(byTestId(c, "backlog-project-search") as HTMLInputElement, "demo"));
    expect(box(c, "PROJ")).toBeNull();
    expect(box(c, "DEMO")).not.toBeNull();
    const result = await axe.run(c, { rules: { "color-contrast": { enabled: false } } });
    expect(result.violations.map((v) => v.id)).toEqual([]);
    c.querySelectorAll("button, input").forEach((el) => expect(el.getAttribute("data-testid")).toBeTruthy());
  });

  it("shows No projects available for an empty list", async () => {
    const { host } = scripted(async () => ({ projects: [] }));
    const { c } = await render(host);
    expect(byTestId(c, "backlog-projects-empty")!.textContent).toBe(en.noProjects);
  });

  it("saves the checked keys", async () => {
    const { host, save } = scripted(async () => ({ projects: PROJECTS }));
    const { c, announce } = await render(host);
    await act(async () => box(c, "DEMO").click());
    await act(async () => byTestId(c, "backlog-projects-save")!.click());
    expect(save).toHaveBeenCalledWith({ projectKeys: ["PROJ", "DEMO"] });
    expect(announce).toHaveBeenCalledWith({ key: "projectsSaved" });
  });

  it("asks before unselecting a project in use, and Cancel saves nothing", async () => {
    const { host, save } = scripted(async () => ({ projects: PROJECTS }));
    const { c } = await render(host);
    await act(async () => box(c, "PROJ").click());
    await act(async () => byTestId(c, "backlog-projects-save")!.click());
    expect(byTestId(c, "backlog-deselect-dialog")!.textContent).toContain(en.deselectTitle);
    await act(async () => byTestId(c, "backlog-deselect-dialog-cancel")!.click());
    expect(save).not.toHaveBeenCalled();
    await act(async () => byTestId(c, "backlog-projects-save")!.click());
    await act(async () => byTestId(c, "backlog-deselect-dialog-confirm")!.click());
    expect(save).toHaveBeenCalledWith({ projectKeys: [] });
  });

  it("shows the field error when Backlog refuses the keys", async () => {
    const { host } = scripted(
      async () => ({ projects: PROJECTS }),
      vi.fn<(body: unknown) => Promise<unknown>>(async () => {
        throw actionError(400, { code: "validation", field: "projectKeys" });
      }),
    );
    const { c } = await render(host);
    await act(async () => box(c, "DEMO").click());
    await act(async () => byTestId(c, "backlog-projects-save")!.click());
    expect(text(c)).toContain(en.errorProjectKeys);
  });

  it("counts down on rate_limited, announces once and retries the load once", async () => {
    vi.useFakeTimers();
    let attempt = 0;
    const { host } = scripted(async () => {
      attempt += 1;
      if (attempt === 1) throw actionError(429, { code: "rate_limited", retryAfterSeconds: 3 });
      return { projects: PROJECTS };
    });
    const { c, announce } = await render(host);
    expect(byTestId(c, "backlog-projects-wait")!.textContent).toBe(
      "Backlog is limiting requests. Retrying in 3 seconds",
    );
    await act(async () => vi.advanceTimersByTime(1000));
    expect(byTestId(c, "backlog-projects-wait")!.textContent).toBe(
      "Backlog is limiting requests. Retrying in 2 seconds",
    );
    await act(async () => vi.advanceTimersByTime(2000));
    expect(attempt).toBe(2);
    expect(box(c, "PROJ")).not.toBeNull();
    expect(announce).toHaveBeenCalledTimes(1);
    expect(announce).toHaveBeenCalledWith({ key: "rateLimitedRetrying", params: { seconds: 3 } });
  });

  it("says how many issue links, PR links and watches of the unselected projects are turned off (AC1.9.1)", async () => {
    const { host } = scripted(async () => ({ projects: PROJECTS }));
    const { c } = await render(host);
    await act(async () => box(c, "PROJ").click());
    await act(async () => byTestId(c, "backlog-projects-save")!.click());
    expect(byTestId(c, "backlog-deselect-dialog")!.textContent).toContain(
      "4 issue links, 3 PR links and 2 PR watches will be turned off.",
    );
    for (const key of ["git.impact", "issues.impact"]) {
      const impact = vi.mocked(host.api.invokeAction).mock.calls.find(([k]) => k === key)!;
      expect(impact[1]).toEqual({ workspaceId: "ws-1", body: { projectKeys: ["PROJ"] } });
    }
  });
});
