import { act, createElement as h } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  expectOnlyCatalogueText,
  expectTestIds,
  fakeHost,
  mount,
  pseudoCatalogue,
  rawControls,
  setValue,
  text,
  unmount,
} from "../testing/harness";
import { createSourceControlSection } from "./source-control-section";

afterEach(unmount);

const NOT_CONFIGURED = (provider: string) => ({ provider, state: "not_configured", mappings: [] });
const GITHUB = {
  provider: "github",
  state: "connected",
  account: "Lan",
  mappings: [{ projectKey: "PROJ", repos: ["acme/web"] }],
};

type Handler = (body: Record<string, unknown>) => Promise<unknown>;

function setup(handlers: Record<string, Handler> = {}, providers: unknown[] = [GITHUB]) {
  const list = [GITHUB, NOT_CONFIGURED("gitlab"), NOT_CONFIGURED("bitbucket")].map(
    (d) => (providers as { provider: string }[]).find((p) => p.provider === d.provider) ?? d,
  );
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    if (key === "scm.providers.list") return { providers: list };
    throw new Error(`unexpected ${key}`);
  });
}

const calls = (host: ReturnType<typeof setup>, key: string) =>
  vi.mocked(host.api.invokeAction).mock.calls.filter(([k]) => k === key);

async function render(
  host: ReturnType<typeof setup>,
  props: Partial<{ readOnly: boolean; selectedProjects: string[] }> = {},
  messages = en,
) {
  return mount(createSourceControlSection(host, messages), {
    workspaceId: "ws-1",
    selectedProjects: ["PROJ", "DEMO"],
    readOnly: false,
    backlogGit: <p data-testid="fake-git-access">git</p>,
    ...props,
  });
}

describe("Source control settings (FR2, FR3)", () => {
  it("lists Backlog Git and the three providers with their state (FR2.1, FR2.5)", async () => {
    const c = await render(setup());
    expect(byTestId(c, "backlog-section-source-control")).not.toBeNull();
    expect(byTestId(c, "backlog-scm-backlog")!.textContent).toContain(en.providerBacklog);
    expect(byTestId(c, "fake-git-access"), "the Git access form stays, for Backlog Git").not.toBeNull();
    expect(byTestId(c, "backlog-scm-github-state")!.textContent).toBe(en.scmStateConnected);
    expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe("Connected as Lan");
    expect(byTestId(c, "backlog-scm-gitlab-state")!.textContent).toBe(en.scmStateNotConfigured);
    expect(byTestId(c, "backlog-scm-bitbucket-state")!.textContent).toBe(en.scmStateNotConfigured);
  });

  it("states the read scopes each provider needs (FR2.3)", async () => {
    const c = await render(setup());
    expect(byTestId(c, "backlog-scm-github-scopes")!.textContent).toBe(en.scmScopesGithub);
    expect(byTestId(c, "backlog-scm-gitlab-scopes")!.textContent).toBe(en.scmScopesGitlab);
    expect(byTestId(c, "backlog-scm-bitbucket-scopes")!.textContent).toBe(en.scmScopesBitbucket);
  });

  it("saves a token, empties the input and never fills it from state (FR2.2, NFR1)", async () => {
    const saved = { provider: "gitlab", state: "connected", account: "Minh", mappings: [] };
    const host = setup({ "scm.providers.set_token": async () => saved });
    const c = await render(host);
    const input = byTestId(c, "backlog-scm-gitlab-token") as HTMLInputElement;
    expect(input.type).toBe("password");
    expect(input.value).toBe("");
    await act(async () => setValue(input, "glpat-FAKE"));
    await act(async () => byTestId(c, "backlog-scm-gitlab-save")!.click());
    expect(calls(host, "scm.providers.set_token")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { provider: "gitlab", token: "glpat-FAKE" },
    });
    expect((byTestId(c, "backlog-scm-gitlab-token") as HTMLInputElement).value).toBe("");
    expect(byTestId(c, "backlog-scm-gitlab-state")!.textContent).toBe(en.scmStateConnected);
    expect(text(c)).not.toContain("glpat-FAKE");
    expect(byTestId(c, "backlog-scm-gitlab-save")!.textContent).toBe(en.scmReplaceToken);
  });

  it("sends the Bitbucket user name with the token (FR2.2)", async () => {
    const host = setup({
      "scm.providers.set_token": async () => ({ provider: "bitbucket", state: "connected", mappings: [] }),
    });
    const c = await render(host);
    await act(async () => {
      setValue(byTestId(c, "backlog-scm-bitbucket-username") as HTMLInputElement, "lan@example.com");
      setValue(byTestId(c, "backlog-scm-bitbucket-token") as HTMLInputElement, "ATBB-FAKE");
    });
    await act(async () => byTestId(c, "backlog-scm-bitbucket-save")!.click());
    expect(calls(host, "scm.providers.set_token")[0]![1]).toMatchObject({
      body: { provider: "bitbucket", token: "ATBB-FAKE", username: "lan@example.com" },
    });
    expect(byTestId(c, "backlog-scm-github-username"), "only Bitbucket asks for a user name").toBeNull();
  });

  it("shows a refused token as a plain error and keeps nothing (FR2.4)", async () => {
    const host = setup({
      "scm.providers.set_token": async () => {
        throw actionError(401, { code: "reconnect_required" });
      },
    });
    const c = await render(host);
    await act(async () => setValue(byTestId(c, "backlog-scm-gitlab-token") as HTMLInputElement, "bad"));
    await act(async () => byTestId(c, "backlog-scm-gitlab-save")!.click());
    expect(byTestId(c, "backlog-scm-gitlab-message")!.textContent).toBe(en.scmTokenRefused);
    expect((byTestId(c, "backlog-scm-gitlab-token") as HTMLInputElement).value).toBe("");
  });

  it("tests the token: the account on success, a plain reason on failure (FR2.4)", async () => {
    let reply: unknown = { ...GITHUB, account: "Lan Nguyen" };
    const host = setup({ "scm.providers.test": async () => reply });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scm-github-test")!.click());
    expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe("Connected as Lan Nguyen");
    expect(byTestId(c, "backlog-scm-github-message")!.textContent).toBe(en.scmTestOk);
    for (const [code, key] of [
      ["invalid_token", "scmErrorInvalidToken"],
      ["missing_scope", "scmErrorMissingScope"],
      ["rate_limited", "scmErrorRateLimited"],
      ["unreachable", "scmErrorUnreachable"],
    ] as const) {
      reply = { ...GITHUB, state: "error", lastError: code };
      await act(async () => byTestId(c, "backlog-scm-github-test")!.click());
      expect(byTestId(c, "backlog-scm-github-state")!.textContent).toBe(en.scmStateError);
      expect(byTestId(c, "backlog-scm-github-message")!.textContent).toBe(en[key]);
    }
  });

  it("removes a token and keeps the mappings (FR2.2)", async () => {
    const host = setup({
      "scm.providers.remove": async () => ({ ...GITHUB, state: "not_configured", account: undefined }),
    });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scm-github-remove")!.click());
    expect(calls(host, "scm.providers.remove")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { provider: "github" },
    });
    expect(byTestId(c, "backlog-scm-github-state")!.textContent).toBe(en.scmStateNotConfigured);
    expect(byTestId(c, "backlog-scm-github-remove")).toBeNull();
    expect(byTestId(c, "backlog-scm-github-map-PROJ")!.textContent).toContain("acme/web");
  });

  it("maps each selected project to repositories: search or type, then save (FR3.1, FR3.2)", async () => {
    const mapped = { ...GITHUB, mappings: [{ projectKey: "DEMO", repos: ["acme/api"] }, GITHUB.mappings[0]] };
    const host = setup({
      "scm.repos.search": async () => ({ repos: [{ fullName: "acme/api" }, { fullName: "acme/web" }] }),
      "scm.mappings.set": async () => mapped,
    });
    const c = await render(host);
    expect(byTestId(c, "backlog-scm-github-map-PROJ")).not.toBeNull();
    expect(byTestId(c, "backlog-scm-github-map-DEMO")).not.toBeNull();
    await act(async () => setValue(byTestId(c, "backlog-scm-github-DEMO-search") as HTMLInputElement, "api"));
    await act(async () => byTestId(c, "backlog-scm-github-DEMO-search-go")!.click());
    expect(calls(host, "scm.repos.search")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { provider: "github", query: "api" },
    });
    await act(async () => byTestId(c, "backlog-scm-github-DEMO-add-acme/api")!.click());
    expect(calls(host, "scm.mappings.set")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { provider: "github", projectKey: "DEMO", repos: ["acme/api"] },
    });
    expect(byTestId(c, "backlog-scm-github-map-DEMO")!.textContent).toContain("acme/api");

    await act(async () =>
      setValue(byTestId(c, "backlog-scm-github-PROJ-manual") as HTMLInputElement, "acme/docs"),
    );
    await act(async () => byTestId(c, "backlog-scm-github-PROJ-manual-add")!.click());
    expect(calls(host, "scm.mappings.set")[1]![1]).toMatchObject({
      body: { projectKey: "PROJ", repos: ["acme/web", "acme/docs"] },
    });
  });

  it("unmaps a repository and shows a refused name as an error (FR3.2, FR3.4)", async () => {
    let fail = false;
    const host = setup({
      "scm.mappings.set": async (body) => {
        if (fail) throw actionError(400, { code: "validation", field: "repos" });
        return { ...GITHUB, mappings: (body.repos as string[]).length ? GITHUB.mappings : [] };
      },
    });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scm-github-PROJ-remove-acme/web")!.click());
    expect(calls(host, "scm.mappings.set")[0]![1]).toMatchObject({ body: { projectKey: "PROJ", repos: [] } });
    fail = true;
    await act(async () =>
      setValue(byTestId(c, "backlog-scm-github-PROJ-manual") as HTMLInputElement, "not a repo"),
    );
    await act(async () => byTestId(c, "backlog-scm-github-PROJ-manual-add")!.click());
    expect(byTestId(c, "backlog-scm-github-PROJ-error")!.textContent).toBe(en.scmErrorRepos);
  });

  it("is read-only for members: states and mappings, no inputs (FR2.6)", async () => {
    const c = await render(setup(), { readOnly: true });
    expect(byTestId(c, "backlog-scm-github-state")!.textContent).toBe(en.scmStateConnected);
    expect(byTestId(c, "backlog-scm-github-map-PROJ")!.textContent).toContain("acme/web");
    expect(c.querySelectorAll("input").length).toBe(0);
    expect(byTestId(c, "backlog-scm-github-save")).toBeNull();
    expect(byTestId(c, "backlog-scm-github-PROJ-remove-acme/web")).toBeNull();
    expect(byTestId(c, "fake-git-access"), "members do not see Git access").toBeNull();
  });

  it("shows a load failure with Retry", async () => {
    let fail = true;
    const host = setup({
      "scm.providers.list": async () => {
        if (fail) throw actionError(503, { code: "unreachable" });
        return { providers: [GITHUB] };
      },
    });
    const c = await render(host);
    expect(byTestId(c, "backlog-scm-error")).not.toBeNull();
    fail = false;
    await act(async () => byTestId(c, "backlog-scm-retry")!.click());
    expect(byTestId(c, "backlog-scm-github-state")).not.toBeNull();
  });

  it("uses host controls, catalogue text, test ids and passes axe", async () => {
    const c = await render(setup(), {}, pseudoCatalogue(en));
    expect(rawControls(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
    expect(await axeViolations(c)).toEqual([]);
    unmount();
    const p = await render(setup(), { readOnly: true }, pseudoCatalogue(en));
    expectOnlyCatalogueText(p, (ok, msg) => expect(ok, msg).toBe(true));
  });
});
