import { describe, expect, it, vi } from "vitest";

import { actionError, fakeHost, type Invoke } from "../testing/harness";
import { createRepositoryProvider } from "./repository-provider";

const DESCRIPTOR = {
  provider_id: "nulab-backlog",
  provider_host: "https://example-space.backlog.com",
  provider_scope: "example-space.backlog.com",
  provider_repository_id: "11",
  owner_or_project: "PROJ",
  name: "web-app",
  clone_url: "https://example-space.backlog.com/git/PROJ/web-app.git",
  default_branch: "main",
};

const INSPECTION = {
  providerId: "nulab-backlog",
  providerHost: "https://example-space.backlog.com",
  providerScope: "example-space.backlog.com",
  ownerOrProject: "PROJ",
  repositoryId: "11",
  repositoryName: "web-app",
  cloneUrl: "https://example-space.backlog.com/git/PROJ/web-app.git",
  defaultBranch: "main",
};

function setup(invoke: Invoke) {
  const host = fakeHost(invoke);
  return { host, provider: createRepositoryProvider(host) };
}

describe("Repository provider (M9, US5.1)", () => {
  it("is the nulab-backlog provider without drafts", () => {
    const { provider } = setup(async () => ({}));
    expect(provider.id).toBe("nulab-backlog");
    expect(provider.label).toBe("Backlog");
    expect(provider.supportsDraft).toBe(false);
    expect(provider.matchesURL?.("https://x.backlog.jp/git/PROJ/a.git")).toBe(true);
    expect(provider.matchesURL?.("https://github.com/acme/a.git")).toBe(false);
  });

  it("forwards query, cursor and signal to git.repositories.list", async () => {
    const page = { repositories: [INSPECTION], nextCursor: "c2" };
    const { host, provider } = setup(async () => page);
    const signal = new AbortController().signal;
    const got = await provider.listRepositories({ workspaceId: "ws-1", query: "web", cursor: "c1", signal });
    expect(got).toEqual(page);
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "git.repositories.list",
      { workspaceId: "ws-1", body: { query: "web", cursor: "c1" } },
      { signal },
    );
  });

  it("rejects on an action error so Kandev shows its error with Retry (AC5.1.2)", async () => {
    const { provider } = setup(async () => {
      throw actionError(503, { code: "unreachable" });
    });
    await expect(
      provider.listRepositories({ workspaceId: "ws-1", signal: new AbortController().signal }),
    ).rejects.toThrow("Could not reach Backlog.");
  });

  it("inspects a URL and maps the descriptor, or returns null", async () => {
    let reply: unknown = { repository: DESCRIPTOR };
    const { host, provider } = setup(async () => reply);
    const signal = new AbortController().signal;
    const url = "https://example-space.backlog.com/git/PROJ/web-app.git";
    expect(await provider.inspectURL({ workspaceId: "ws-1", url, signal })).toEqual(INSPECTION);
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "repositories.inspect",
      { workspaceId: "ws-1", body: { url } },
      { signal },
    );
    reply = { matched: false };
    expect(await provider.inspectURL({ workspaceId: "ws-1", url, signal })).toBeNull();
  });

  it("lists branches from repositories.branches", async () => {
    const { host, provider } = setup(async () => ({
      branches: [{ name: "main", is_default: true }, { name: "dev" }],
    }));
    const signal = new AbortController().signal;
    const got = await provider.listBranches({ workspaceId: "ws-1", repository: INSPECTION, signal });
    expect(got).toEqual([{ name: "main" }, { name: "dev" }]);
    expect(host.api.invokeAction).toHaveBeenCalledWith(
      "repositories.branches",
      { workspaceId: "ws-1", body: { repository: DESCRIPTOR } },
      { signal },
    );
  });

  it("publishes nothing for an aborted request", async () => {
    const controller = new AbortController();
    const { provider } = setup(async () => {
      controller.abort();
      return { repositories: [INSPECTION] };
    });
    const onDone = vi.fn();
    await expect(
      provider.listRepositories({ workspaceId: "ws-1", signal: controller.signal }).then(onDone),
    ).rejects.toThrow();
    expect(onDone).not.toHaveBeenCalled();
  });
});
