import type { PluginHostApi, RepositoryInspection, RepositoryProviderRegistration } from "@kandev/plugin-sdk";

import { en, type Messages } from "../messages/en";
import { PLUGIN_ID } from "../switch/enabled-events";
import { createChangeRequest } from "./create-pr";
import { gitNotice, noticeError } from "./git-state";

/** The repositories.inspect / repositories.branches descriptor (snake_case, C8). */
interface Descriptor {
  provider_id: string;
  provider_host: string;
  provider_scope: string;
  provider_repository_id: string;
  owner_or_project: string;
  name: string;
  clone_url: string;
  default_branch: string;
}

const BACKLOG_GIT_URL = /^https:\/\/[a-z0-9-]+\.(backlog\.com|backlog\.jp|backlogtool\.com)\/git\//i;

function toInspection(d: Descriptor): RepositoryInspection {
  return {
    providerId: d.provider_id,
    providerHost: d.provider_host,
    providerScope: d.provider_scope,
    ownerOrProject: d.owner_or_project,
    repositoryId: d.provider_repository_id,
    repositoryName: d.name,
    cloneUrl: d.clone_url,
    defaultBranch: d.default_branch,
  };
}

function toDescriptor(r: RepositoryInspection): Descriptor {
  return {
    provider_id: r.providerId,
    provider_host: r.providerHost,
    provider_scope: r.providerScope ?? "",
    provider_repository_id: r.repositoryId,
    owner_or_project: r.ownerOrProject,
    name: r.repositoryName,
    clone_url: r.cloneUrl,
    default_branch: r.defaultBranch ?? "",
  };
}

/** Throws when signal was aborted, so a late reply publishes nothing. */
function live<T>(signal: AbortSignal, value: T): T {
  if (signal.aborted) throw new DOMException("Aborted", "AbortError");
  return value;
}

/**
 * Backlog Git as a Kandev repository provider (M9, US5.1). Kandev renders
 * the picker, the error with Retry and the Create PR dialog.
 */
export function createRepositoryProvider(
  host: PluginHostApi,
  messages: Messages = en,
): RepositoryProviderRegistration {
  const call = async <T>(
    key: string,
    workspaceId: string,
    body: unknown,
    signal: AbortSignal,
  ): Promise<T> => {
    try {
      return live(signal, await host.api.invokeAction<T>(key, { workspaceId, body }, { signal }));
    } catch (error) {
      if (signal.aborted) throw error;
      throw noticeError(gitNotice(error), messages, error);
    }
  };
  return {
    id: PLUGIN_ID,
    label: messages.integrationLabel,
    supportsDraft: false,
    matchesURL: (url) => BACKLOG_GIT_URL.test(url),
    listRepositories: ({ workspaceId, query, cursor, signal }) =>
      call("git.repositories.list", workspaceId, { query, cursor }, signal),
    inspectURL: async ({ workspaceId, url, signal }) => {
      const reply = await call<{ repository?: Descriptor }>(
        "repositories.inspect",
        workspaceId,
        { url },
        signal,
      );
      return reply?.repository ? toInspection(reply.repository) : null;
    },
    listBranches: async ({ workspaceId, repository, signal }) => {
      const reply = await call<{ branches?: { name: string }[] }>(
        "repositories.branches",
        workspaceId,
        { repository: toDescriptor(repository) },
        signal,
      );
      return (reply?.branches ?? []).map((b) => ({ name: b.name }));
    },
    createChangeRequest: createChangeRequest(host, messages),
  };
}
