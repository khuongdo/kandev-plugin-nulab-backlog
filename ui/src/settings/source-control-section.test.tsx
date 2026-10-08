import { act, createElement as h } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en, format } from "../messages/en";
import {
  actionError,
  axeViolations,
  byTestId,
  choose,
  expectOnlyCatalogueText,
  expectTestIds,
  fakeHost,
  mount,
  optionValues,
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

/**
 * A host whose scm.providers.list returns the providers and the active
 * service; "" (pending, FR3.3) shows every card, as before the upgrade.
 */
function setup(handlers: Record<string, Handler> = {}, providers: unknown[] = [GITHUB], active = "") {
  const list = [GITHUB, NOT_CONFIGURED("gitlab"), NOT_CONFIGURED("bitbucket")].map(
    (d) => (providers as { provider: string }[]).find((p) => p.provider === d.provider) ?? d,
  );
  return fakeHost(async (key, input) => {
    const body = (input?.body ?? {}) as Record<string, unknown>;
    if (handlers[key]) return handlers[key](body);
    if (key === "scm.providers.list") return { providers: list, active };
    if (key === "scm.active.set") return { providers: list, active: body.service };
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

  it("offers the CLI login on GitHub and GitLab only, to admins only (FR1.1, FR1.4, FR2.1, FR2.3)", async () => {
    const c = await render(setup());
    expect(byTestId(c, "backlog-scm-github-use-cli")!.textContent).toBe("Use gh CLI login");
    expect(byTestId(c, "backlog-scm-gitlab-use-cli")!.textContent).toBe("Use glab CLI login");
    expect(byTestId(c, "backlog-scm-bitbucket-use-cli")).toBeNull();
    expect(byTestId(c, "backlog-scm-github-token"), "the token field stays").not.toBeNull();
    unmount();
    const m = await render(setup(), { readOnly: true });
    expect(byTestId(m, "backlog-scm-github-use-cli")).toBeNull();
    expect(byTestId(m, "backlog-scm-gitlab-use-cli")).toBeNull();
  });

  it("connects with the CLI login and shows the method (FR1.2, FR5.2)", async () => {
    const viaCli = { provider: "gitlab", state: "connected", method: "cli", account: "Minh", mappings: [] };
    const host = setup({ "scm.providers.use_cli": async () => viaCli });
    const c = await render(host);
    await act(async () => byTestId(c, "backlog-scm-gitlab-use-cli")!.click());
    expect(calls(host, "scm.providers.use_cli")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { provider: "gitlab" },
    });
    expect(byTestId(c, "backlog-scm-gitlab-state")!.textContent).toBe(en.scmStateConnected);
    expect(byTestId(c, "backlog-scm-gitlab-account")!.textContent).toBe("Connected via glab CLI as Minh");
    expect(byTestId(c, "backlog-scm-gitlab-message")!.textContent).toBe("Connected with the glab CLI login.");
    expect(byTestId(c, "backlog-scm-gitlab-save")!.textContent, "a typed token can replace it").toBe(
      en.scmReplaceToken,
    );
  });

  it("shows a typed token connection as before (FR5.2, FR6.1)", async () => {
    const c = await render(setup({}, [{ ...GITHUB, method: "token" }]));
    expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe("Connected as Lan");
  });

  it("explains an unusable CLI, from the action and from Test (FR4.2)", async () => {
    const cliError = "The gh CLI is not available or not logged in on the Kandev server.";
    const host = setup(
      {
        "scm.providers.cli_accounts": async () => {
          throw actionError(503, { code: "cli_unavailable" });
        },
        "scm.providers.test": async () => ({
          ...GITHUB,
          method: "cli",
          state: "error",
          lastError: "cli_unavailable",
        }),
      },
      [{ ...GITHUB, method: "cli" }],
    );
    const c = await render(host);
    expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe("Connected via gh CLI as Lan");
    await act(async () => byTestId(c, "backlog-scm-github-use-cli")!.click());
    expect(byTestId(c, "backlog-scm-github-message")!.textContent).toBe(cliError);
    await act(async () => byTestId(c, "backlog-scm-github-test")!.click());
    expect(byTestId(c, "backlog-scm-github-state")!.textContent).toBe(en.scmStateError);
    expect(byTestId(c, "backlog-scm-github-message")!.textContent).toBe(cliError);
  });

  describe("gh account per workspace (intent 261008-gh-cli-profile)", () => {
    const TWO = {
      accounts: [
        { login: "alice", active: true },
        { login: "bob", active: false },
      ],
    };
    const VIA_GH = { ...GITHUB, method: "cli", account: "bob", login: "bob" };
    const MISSING = "bob is not logged in to gh on the Kandev server — log in again or pick another account";
    const NOTE =
      "Agents working in task worktrees get their GitHub login from Kandev's own GitHub integration (or the executor profile), not from this plugin. Set it to the same account (@bob) for this workspace.";
    const pending = () => {
      let release: (v: unknown) => void = () => {};
      const promise = new Promise((r) => (release = r));
      return { promise, release };
    };

    it("loads the gh accounts, then shows a labelled picker with the active one preselected (AC1.1.1)", async () => {
      const wait = pending();
      const host = setup({
        "scm.providers.cli_accounts": () => wait.promise,
        "scm.providers.use_cli": async () => VIA_GH,
      });
      const c = await render(host);
      await act(async () => byTestId(c, "backlog-scm-github-use-cli")!.click());
      expect((byTestId(c, "backlog-scm-github-use-cli") as HTMLButtonElement).disabled).toBe(true);
      expect(byTestId(c, "backlog-scm-github-message")!.textContent).toBe("Loading gh accounts…");
      expect(byTestId(c, "backlog-scm-github-message")!.getAttribute("role")).toBe("status");
      await act(async () => wait.release(TWO));
      const select = byTestId(c, "backlog-scm-github-account-select")!;
      expect(c.querySelector('label[for="backlog-scm-github-account-select"]')!.textContent).toBe(
        "GitHub account (gh)",
      );
      expect(select.closest("[data-host=Select]")!.getAttribute("data-value")).toBe("alice");
      expect(byTestId(c, "backlog-scm-github-account-alice")!.textContent).toBe("alice (active in gh)");
      expect(byTestId(c, "backlog-scm-github-account-bob")!.textContent).toBe("bob");
      expect(document.activeElement).toBe(select);
      expect(byTestId(c, "backlog-scm-github-account-connect")).not.toBeNull();
      expect(byTestId(c, "backlog-scm-github-account-cancel")).not.toBeNull();
      expect(calls(host, "scm.providers.use_cli")).toHaveLength(0);
    });

    it("connects the picked account and shows its login (AC1.1.2, AC3.2.2)", async () => {
      const host = setup({
        "scm.providers.cli_accounts": async () => TWO,
        "scm.providers.use_cli": async () => VIA_GH,
      });
      const c = await render(host);
      await act(async () => byTestId(c, "backlog-scm-github-use-cli")!.click());
      await act(async () => byTestId(c, "backlog-scm-github-account-bob")!.click());
      await act(async () => byTestId(c, "backlog-scm-github-account-connect")!.click());
      expect(calls(host, "scm.providers.use_cli")[0]![1]).toEqual({
        workspaceId: "ws-1",
        body: { provider: "github", login: "bob" },
      });
      expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe("Connected via gh CLI as @bob");
      expect(byTestId(c, "backlog-scm-github-account-select")).toBeNull();
      unmount();
      const named = await render(setup({}, [{ ...VIA_GH, account: "Bob Smith" }]));
      expect(byTestId(named, "backlog-scm-github-account")!.textContent).toBe(
        "Connected via gh CLI as @bob (Bob Smith)",
      );
    });

    it("connects at once when gh has one account, also for a gh without --json (AC1.1.4, AC1.1.9)", async () => {
      const host = setup({
        "scm.providers.cli_accounts": async () => ({ accounts: [{ login: "alice", active: true }] }),
        "scm.providers.use_cli": async () => ({ ...VIA_GH, login: "alice", account: "alice" }),
      });
      const c = await render(host);
      await act(async () => byTestId(c, "backlog-scm-github-use-cli")!.click());
      expect(byTestId(c, "backlog-scm-github-account-select")).toBeNull();
      expect(calls(host, "scm.providers.use_cli")[0]![1]).toMatchObject({
        body: { provider: "github", login: "alice" },
      });
      expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe("Connected via gh CLI as @alice");
    });

    it("shows an unusable gh as an alert and saves nothing (AC1.1.3)", async () => {
      const host = setup({
        "scm.providers.cli_accounts": async () => {
          throw actionError(503, { code: "cli_unavailable" });
        },
      });
      const c = await render(host);
      await act(async () => byTestId(c, "backlog-scm-github-use-cli")!.click());
      const msg = byTestId(c, "backlog-scm-github-message")!;
      expect(msg.textContent).toBe("The gh CLI is not available or not logged in on the Kandev server.");
      expect(msg.getAttribute("role")).toBe("alert");
      expect(calls(host, "scm.providers.use_cli")).toHaveLength(0);
    });

    it("changes the account with the current login preselected; Cancel saves nothing (AC1.2.1, AC1.2.4)", async () => {
      const host = setup(
        {
          "scm.providers.cli_accounts": async () => TWO,
          "scm.providers.use_cli": async () => ({ ...VIA_GH, login: "alice", account: "Alice" }),
        },
        [VIA_GH],
      );
      const c = await render(host);
      await act(async () => byTestId(c, "backlog-scm-github-change-account")!.click());
      const select = byTestId(c, "backlog-scm-github-account-select")!;
      expect(select.closest("[data-host=Select]")!.getAttribute("data-value")).toBe("bob");
      await act(async () => byTestId(c, "backlog-scm-github-account-cancel")!.click());
      expect(byTestId(c, "backlog-scm-github-account-select")).toBeNull();
      expect(calls(host, "scm.providers.use_cli")).toHaveLength(0);
      expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe("Connected via gh CLI as @bob");

      await act(async () => byTestId(c, "backlog-scm-github-change-account")!.click());
      await act(async () => byTestId(c, "backlog-scm-github-account-alice")!.click());
      await act(async () => byTestId(c, "backlog-scm-github-account-connect")!.click());
      expect(calls(host, "scm.providers.use_cli")[0]![1]).toMatchObject({
        body: { provider: "github", login: "alice" },
      });
      expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe(
        "Connected via gh CLI as @alice (Alice)",
      );
    });

    it("names the missing account in an alert and offers only gh's accounts (AC3.1.1, AC3.1.4)", async () => {
      const host = setup(
        {
          "scm.providers.test": async () => ({ ...VIA_GH, state: "error", lastError: "cli_account_missing" }),
          "scm.providers.cli_accounts": async () => ({ accounts: [{ login: "alice", active: true }] }),
        },
        [VIA_GH],
      );
      const c = await render(host);
      await act(async () => byTestId(c, "backlog-scm-github-test")!.click());
      const msg = byTestId(c, "backlog-scm-github-message")!;
      expect(msg.textContent).toBe(MISSING);
      expect(msg.getAttribute("role")).toBe("alert");
      await act(async () => byTestId(c, "backlog-scm-github-change-account")!.click());
      const select = byTestId(c, "backlog-scm-github-account-select")!;
      expect(select.closest("[data-host=Select]")!.getAttribute("data-value")).toBe("");
      expect(byTestId(c, "backlog-scm-github-account-bob")).toBeNull();
      expect((byTestId(c, "backlog-scm-github-account-connect") as HTMLButtonElement).disabled).toBe(true);
    });

    it("shows a refused change as the account-missing alert (AC1.2.3, AC3.1.1)", async () => {
      const host = setup(
        {
          "scm.providers.cli_accounts": async () => TWO,
          "scm.providers.use_cli": async () => {
            throw actionError(409, { code: "cli_account_missing" });
          },
        },
        [{ ...VIA_GH, login: "alice", account: "Alice" }],
      );
      const c = await render(host);
      await act(async () => byTestId(c, "backlog-scm-github-change-account")!.click());
      await act(async () => byTestId(c, "backlog-scm-github-account-bob")!.click());
      await act(async () => byTestId(c, "backlog-scm-github-account-connect")!.click());
      expect(byTestId(c, "backlog-scm-github-message")!.textContent).toBe(MISSING);
      expect(byTestId(c, "backlog-scm-github-account")!.textContent).toBe(
        "Connected via gh CLI as @alice (Alice)",
      );
    });

    it("asks to pick an account when a gh connection has no login (AC3.2.4)", async () => {
      const c = await render(
        setup({}, [{ ...GITHUB, method: "cli", state: "error", lastError: "cli_account_missing" }]),
      );
      const msg = byTestId(c, "backlog-scm-github-message")!;
      expect(msg.textContent).toBe("Pick the gh account this workspace uses.");
      expect(msg.getAttribute("role")).toBe("alert");
      expect(byTestId(c, "backlog-scm-github-change-account")).not.toBeNull();
    });

    it("explains where agents in task worktrees get their GitHub login (AC4.1.1)", async () => {
      const c = await render(setup({}, [VIA_GH]));
      expect(byTestId(c, "backlog-scm-github-worktree-note")!.textContent).toBe(NOTE);
      unmount();
      const token = await render(setup({}, [{ ...GITHUB, method: "token" }]));
      expect(byTestId(token, "backlog-scm-github-worktree-note")).toBeNull();
      unmount();
      const member = await render(setup({}, [VIA_GH]), { readOnly: true });
      expect(byTestId(member, "backlog-scm-github-worktree-note")!.textContent).toBe(NOTE);
      expect(byTestId(member, "backlog-scm-github-account")!.textContent).toBe(
        "Connected via gh CLI as @bob",
      );
      expect(byTestId(member, "backlog-scm-github-change-account")).toBeNull();
    });

    it("keeps host controls, test ids and axe with the picker open", async () => {
      const c = await render(setup({ "scm.providers.cli_accounts": async () => TWO }, [VIA_GH]));
      await act(async () => byTestId(c, "backlog-scm-github-change-account")!.click());
      expect(rawControls(c)).toEqual([]);
      expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
      expect(await axeViolations(c)).toEqual([]);
    });
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

describe("One active source control service (intent 261008-source-control-settings)", () => {
  const SERVICES = ["backlog_git", "github", "gitlab", "bitbucket"];
  const cards = (c: HTMLElement) =>
    ["backlog", "github", "gitlab", "bitbucket"].filter((p) => byTestId(c, `backlog-scm-${p}`) !== null);

  it("shows a labelled Service selector and only the active service's card (FR2.1, FR1.1)", async () => {
    const c = await render(setup({}, [GITHUB], "github"));
    expect(optionValues(c, "backlog-scm-service")).toEqual(SERVICES);
    expect(c.querySelector('label[for="backlog-scm-service"]')!.textContent).toBe(en.scmServiceLabel);
    expect(
      byTestId(c, "backlog-scm-service")!.closest("[data-host=Select]")!.getAttribute("data-value"),
    ).toBe("github");
    expect(cards(c)).toEqual(["github"]);
    expect(byTestId(c, "backlog-scm-pick-notice")).toBeNull();
    unmount();
    const g = await render(setup({}, [GITHUB], "backlog_git"));
    expect(cards(g)).toEqual(["backlog"]);
    expect(byTestId(g, "fake-git-access"), "Backlog Git keeps its Git access form").not.toBeNull();
  });

  it("shows members the active service read-only, with no selector (FR2.2)", async () => {
    const c = await render(setup({}, [GITHUB], "gitlab"), { readOnly: true });
    expect(byTestId(c, "backlog-scm-service")).toBeNull();
    expect(byTestId(c, "backlog-scm-service-readonly")!.textContent).toBe(
      format(en.scmServiceReadOnly, { service: "GitLab" }),
    );
    expect(cards(c)).toEqual(["gitlab"]);
  });

  it("confirms a switch naming both services; Cancel sends nothing (FR2.3, FR2.4)", async () => {
    const host = setup({}, [GITHUB], "github");
    const c = await render(host);
    await act(async () => choose(c, "backlog-scm-service", "gitlab"));
    const dialog = byTestId(c, "backlog-scm-switch-dialog")!;
    expect(dialog.textContent).toContain(format(en.scmSwitchTitle, { service: "GitLab" }));
    expect(dialog.textContent).toContain(format(en.scmSwitchBody, { from: "GitHub", to: "GitLab" }));
    expect(en.scmSwitchBody).toContain("kept");
    await act(async () => byTestId(c, "backlog-scm-switch-dialog-cancel")!.click());
    expect(byTestId(c, "backlog-scm-switch-dialog")).toBeNull();
    expect(calls(host, "scm.active.set")).toHaveLength(0);
    expect(cards(c)).toEqual(["github"]);

    await act(async () => choose(c, "backlog-scm-service", "gitlab"));
    await act(async () => byTestId(c, "backlog-scm-switch-dialog-confirm")!.click());
    expect(calls(host, "scm.active.set")[0]![1]).toEqual({
      workspaceId: "ws-1",
      body: { service: "gitlab" },
    });
    expect(byTestId(c, "backlog-scm-switch-dialog")).toBeNull();
    expect(cards(c)).toEqual(["gitlab"]);
  });

  it("asks to pick one service while several are connected, and keeps them all working (FR3.3)", async () => {
    const BITBUCKET = { provider: "bitbucket", state: "connected", account: "Lan", mappings: [] };
    const c = await render(setup({}, [GITHUB, BITBUCKET], ""));
    expect(byTestId(c, "backlog-scm-pick-notice")!.textContent).toBe(en.scmPickNotice);
    expect(cards(c)).toEqual(["backlog", "github", "gitlab", "bitbucket"]);
    expect(
      byTestId(c, "backlog-scm-service")!.closest("[data-host=Select]")!.getAttribute("data-value"),
    ).toBe("");
  });

  it("frames the active service: logo, a heading above its sub-headings, a state badge (FR4.1, FR4.2)", async () => {
    const c = await render(setup({}, [GITHUB], "github"));
    const card = byTestId(c, "backlog-scm-github")!;
    expect(card.className).toContain("border");
    expect(card.className).toContain("rounded");
    expect(card.querySelector("svg")).not.toBeNull();
    const heading = card.querySelector("h3")!;
    expect(heading.textContent).toBe(en.providerGithub);
    expect(heading.className).toContain("text-base");
    expect(card.querySelector("h4")!.className).toContain("text-sm");
    expect(byTestId(c, "backlog-scm-github-state")!.textContent).toBe(en.scmStateConnected);
    unmount();
    const g = await render(setup({}, [GITHUB], "gitlab"));
    expect(byTestId(g, "backlog-scm-gitlab-state")!.textContent).toBe("Not connected");
    expect(byTestId(g, "backlog-scm-gitlab")!.querySelector("h3")!.textContent).toBe(en.providerGitlab);
    unmount();
    const b = await render(setup({}, [GITHUB], "backlog_git"));
    expect(byTestId(b, "backlog-scm-backlog")!.className).toContain("border");
    expect(byTestId(b, "backlog-scm-backlog")!.querySelector("h3")!.textContent).toBe(en.providerBacklog);
  });

  it("names the service in every repository label and line (FR5.1-FR5.3)", async () => {
    const c = await render(setup({}, [GITHUB], "github"));
    const card = byTestId(c, "backlog-scm-github")!;
    expect(card.querySelector("h4")!.textContent).toBe("GitHub repositories linked to Backlog projects");
    expect(c.querySelector('label[for="backlog-scm-github-PROJ-search"]')!.textContent).toBe(
      "Search GitHub repositories for PROJ",
    );
    expect(c.querySelector('label[for="backlog-scm-github-PROJ-manual"]')!.textContent).toBe(
      "GitHub repository for PROJ",
    );
    expect(byTestId(c, "backlog-scm-github-map-PROJ")!.textContent).toContain("PROJ: [GitHub] acme/web");
    expect(byTestId(c, "backlog-scm-github-map-DEMO")!.textContent).toContain("DEMO: no GitHub repositories");
    for (const key of [
      "scmMappingsHeading",
      "scmSearchLabel",
      "scmManualLabel",
      "scmNoRepos",
      "scmRepoName",
      "scmServiceLabel",
      "scmServiceReadOnly",
      "scmSwitchTitle",
      "scmSwitchBody",
      "scmPickNotice",
      "scmServiceInactive",
    ] as const) {
      expect(en[key].toLowerCase(), key).not.toContain("scope");
    }
  });

  it("passes axe with host controls, catalogue text and test ids, dialog open (NFR3, NFR4)", async () => {
    const c = await render(setup({}, [GITHUB], "github"), {}, pseudoCatalogue(en));
    expect(rawControls(c)).toEqual([]);
    expectTestIds(c, (ok, msg) => expect(ok, msg).toBe(true));
    expect(await axeViolations(c)).toEqual([]);
    await act(async () => choose(c, "backlog-scm-service", "bitbucket"));
    expect(await axeViolations(c)).toEqual([]);
    unmount();
    const m = await render(setup({}, [GITHUB], "github"), { readOnly: true }, pseudoCatalogue(en));
    expectOnlyCatalogueText(m, (ok, msg) => expect(ok, msg).toBe(true));
  });
});
