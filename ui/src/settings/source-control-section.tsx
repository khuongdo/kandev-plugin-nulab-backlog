import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText, providerName, scmNotice, type ProviderView, type ScmProvider } from "../git/git-state";
import { hostUi } from "../host-ui";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import { createSectionParts } from "./section-parts";
import { readFailure, type Notice } from "./state";

export interface SourceControlProps {
  workspaceId: string;
  /** The Backlog projects a mapping can use (FR3.1). */
  selectedProjects: string[];
  /** Members see states and mappings only (FR2.6). */
  readOnly: boolean;
  /** The Backlog Git access form, shown to admins (FR2.5). */
  backlogGit?: unknown;
}

const SCOPES: Record<ScmProvider, MessageKey> = {
  github: "scmScopesGithub",
  gitlab: "scmScopesGitlab",
  bitbucket: "scmScopesBitbucket",
};
const STATES: Record<string, MessageKey> = {
  not_configured: "scmStateNotConfigured",
  connected: "scmStateConnected",
  error: "scmStateError",
};
const TEST_ERRORS: Record<string, MessageKey> = {
  invalid_token: "scmErrorInvalidToken",
  missing_scope: "scmErrorMissingScope",
  rate_limited: "scmErrorRateLimited",
  unreachable: "scmErrorUnreachable",
};

type OnView = (view: ProviderView) => void;

/**
 * The "Source control" settings section (FR2, FR3): Backlog Git with its
 * Git access form, then GitHub, GitLab and Bitbucket, each with its state,
 * the read scopes it needs, a token form (admins) and the repositories of
 * each selected Backlog project. A token is never shown or refilled (NFR1).
 */
export function createSourceControlSection(
  host: PluginHostApi,
  messages: Messages = en,
): Component<SourceControlProps> {
  const h = host.jsx;
  const { useCallback, useEffect, useState } = host.React;
  const { Badge, Button, Input, Label, SettingsSection } = hostUi(host);
  const { ListError } = createSectionParts(host, messages);
  const t = (n: Notice) => noticeText(n, messages);
  const invoke = <T,>(key: string, workspaceId: string, body?: unknown) =>
    host.api.invokeAction<T>(key, body === undefined ? { workspaceId } : { workspaceId, body });

  function ProjectRepos(props: {
    key?: string;
    workspaceId: string;
    view: ProviderView;
    project: string;
    readOnly: boolean;
    onView: OnView;
  }) {
    const { workspaceId, view, project, readOnly, onView } = props;
    const p = view.provider;
    const id = `backlog-scm-${p}-${project}`;
    const repos = view.mappings.find((m) => m.projectKey === project)?.repos ?? [];
    const [query, setQuery] = useState("");
    const [manual, setManual] = useState("");
    const [results, setResults] = useState<string[] | undefined>(undefined);
    const [error, setError] = useState<Notice | undefined>(undefined);

    const save = async (next: string[]) => {
      setError(undefined);
      try {
        onView(
          await invoke<ProviderView>("scm.mappings.set", workspaceId, {
            provider: p,
            projectKey: project,
            repos: next,
          }),
        );
        return true;
      } catch (e) {
        const f = readFailure(e);
        setError(f.code === "validation" && f.field === "repos" ? { key: "scmErrorRepos" } : scmNotice(e));
        return false;
      }
    };
    const search = async () => {
      setError(undefined);
      try {
        const r = await invoke<{ repos?: { fullName: string }[] }>("scm.repos.search", workspaceId, {
          provider: p,
          query,
        });
        setResults((r?.repos ?? []).map((x) => x.fullName));
      } catch (e) {
        setError(scmNotice(e));
      }
    };
    const addManual = async () => {
      if (manual.trim() && (await save([...repos, manual.trim()]))) setManual("");
    };

    const line = repos.length
      ? format(messages.scmMappingLine, { project, repos: repos.join(", ") })
      : format(messages.scmNoRepos, { project });
    if (readOnly) return <p data-testid={`backlog-scm-${p}-map-${project}`}>{line}</p>;
    const field = (suffix: string, label: string, value: string, set: (v: string) => void, extra = {}) => (
      <div className={FIELD}>
        <Label htmlFor={`${id}-${suffix}`}>{label}</Label>
        <Input
          id={`${id}-${suffix}`}
          data-testid={`${id}-${suffix}`}
          value={value}
          onChange={(e: { target: { value: string } }) => set(e.target.value)}
          {...extra}
        />
      </div>
    );
    const button = (testId: string, label: string, onClick: () => void, disabled = false) => (
      <Button
        type="button"
        variant="outline"
        size="sm"
        className={BUTTON}
        data-testid={testId}
        disabled={disabled}
        onClick={onClick}
      >
        {label}
      </Button>
    );
    return (
      <div data-testid={`backlog-scm-${p}-map-${project}`} className={STACK}>
        <p>{line}</p>
        <ul className={STACK}>
          {repos.map((repo) => (
            <li key={repo} className={ROW}>
              <span>{format(messages.scmRepoName, { repo })}</span>
              {button(
                `${id}-remove-${repo}`,
                format(messages.scmRemoveRepo, { repo }),
                () => void save(repos.filter((r) => r !== repo)),
              )}
            </li>
          ))}
        </ul>
        <div className={ROW}>
          {field("search", format(messages.scmSearchLabel, { project }), query, setQuery)}
          {button(`${id}-search-go`, messages.scmSearch, () => void search())}
        </div>
        {results ? (
          <ul className={ROW}>
            {results.map((repo) => (
              <li key={repo}>
                {button(
                  `${id}-add-${repo}`,
                  format(messages.scmAddRepo, { repo }),
                  () => void save([...repos, repo]),
                  repos.includes(repo),
                )}
              </li>
            ))}
          </ul>
        ) : null}
        <div className={ROW}>
          {field("manual", format(messages.scmManualLabel, { project }), manual, setManual, {
            placeholder: messages.scmManualPlaceholder,
          })}
          {button(`${id}-manual-add`, messages.scmAdd, () => void addManual())}
        </div>
        {error ? (
          <p role="alert" data-testid={`${id}-error`}>
            {t(error)}
          </p>
        ) : null}
      </div>
    );
  }

  function ProviderCard(props: {
    key?: string;
    workspaceId: string;
    view: ProviderView;
    selectedProjects: string[];
    readOnly: boolean;
    onView: OnView;
  }) {
    const { workspaceId, view, selectedProjects, readOnly, onView } = props;
    const p = view.provider;
    const id = `backlog-scm-${p}`;
    const [token, setToken] = useState("");
    const [username, setUsername] = useState("");
    const [busy, setBusy] = useState(false);
    const [message, setMessage] = useState<Notice | undefined>(undefined);
    const hasToken = view.state !== "not_configured";

    const run = async (key: string, body: Record<string, string>, ok: (v: ProviderView) => Notice) => {
      if (busy) return;
      setBusy(true);
      setMessage(undefined);
      try {
        const v = await invoke<ProviderView>(key, workspaceId, body);
        onView(v);
        setMessage(ok(v));
      } catch (e) {
        setMessage(scmNotice(e));
      }
      setToken(""); // the token never stays in the page (NFR1)
      setBusy(false);
    };
    const saveToken = () =>
      run(
        "scm.providers.set_token",
        { provider: p, token, ...(p === "bitbucket" ? { username } : {}) },
        () => ({
          key: "scmTokenSaved",
        }),
      );
    const test = () =>
      run("scm.providers.test", { provider: p }, (v) => ({
        key: v.lastError ? (TEST_ERRORS[v.lastError] ?? "scmErrorUnreachable") : "scmTestOk",
      }));
    const remove = () => run("scm.providers.remove", { provider: p }, () => ({ key: "scmTokenRemoved" }));
    const button = (suffix: string, label: string, onClick: () => void, variant = "outline") => (
      <Button
        type="button"
        variant={variant}
        size="sm"
        className={BUTTON}
        data-testid={`${id}-${suffix}`}
        disabled={busy}
        onClick={onClick}
      >
        {label}
      </Button>
    );
    const input = (suffix: string, label: string, value: string, set: (v: string) => void, extra = {}) => (
      <div className={FIELD}>
        <Label htmlFor={`${id}-${suffix}`}>{label}</Label>
        <Input
          id={`${id}-${suffix}`}
          data-testid={`${id}-${suffix}`}
          value={value}
          onChange={(e: { target: { value: string } }) => set(e.target.value)}
          {...extra}
        />
      </div>
    );

    return (
      <section data-testid={id} className={STACK}>
        <div className={ROW}>
          <h4 className="text-sm font-medium">{providerName(p, messages)}</h4>
          <Badge variant="outline" data-testid={`${id}-state`}>
            {messages[STATES[view.state] ?? "scmStateError"]}
          </Badge>
        </div>
        {hasToken && view.account ? (
          <p data-testid={`${id}-account`}>{format(messages.scmAccount, { name: view.account })}</p>
        ) : null}
        <p data-testid={`${id}-scopes`} className="text-sm text-muted-foreground">
          {messages[SCOPES[p]]}
        </p>
        {readOnly ? null : (
          <div className={STACK}>
            {p === "bitbucket"
              ? input("username", messages.scmUsernameLabel, username, setUsername, {
                  autoComplete: "username",
                })
              : null}
            {input(
              "token",
              format(messages.scmTokenLabel, { provider: providerName(p, messages) }),
              token,
              setToken,
              {
                type: "password",
                autoComplete: "off",
              },
            )}
            <div className={ROW}>
              {button(
                "save",
                hasToken ? messages.scmReplaceToken : messages.scmSaveToken,
                () => void saveToken(),
                "default",
              )}
              {hasToken ? button("test", messages.scmTest, () => void test()) : null}
              {hasToken ? button("remove", messages.scmRemoveToken, () => void remove()) : null}
            </div>
          </div>
        )}
        {message ? (
          <p role="status" data-testid={`${id}-message`}>
            {t(message)}
          </p>
        ) : null}
        {hasToken || view.mappings.length > 0 ? (
          <div className={STACK}>
            <h5 className="text-sm font-medium">{messages.scmMappingsHeading}</h5>
            {selectedProjects.map((project) => (
              <ProjectRepos
                key={project}
                workspaceId={workspaceId}
                view={view}
                project={project}
                readOnly={readOnly}
                onView={onView}
              />
            ))}
          </div>
        ) : null}
      </section>
    );
  }

  return function SourceControlSection({
    workspaceId,
    selectedProjects,
    readOnly,
    backlogGit,
  }: SourceControlProps) {
    const [views, setViews] = useState<ProviderView[] | undefined>(undefined);
    const [error, setError] = useState<Notice | undefined>(undefined);
    const load = useCallback(() => {
      setError(undefined);
      invoke<{ providers?: ProviderView[] }>("scm.providers.list", workspaceId).then(
        (r) => setViews(r?.providers ?? []),
        (e: unknown) => setError(scmNotice(e)),
      );
    }, [workspaceId]);
    useEffect(load, [load]);
    const onView = (v: ProviderView) =>
      setViews((list) => list?.map((x) => (x.provider === v.provider ? v : x)));

    return (
      <div data-testid="backlog-section-source-control">
        <SettingsSection
          title={messages.sectionSourceControl}
          description={messages.sourceControlDescription}
        >
          <div className="flex flex-col gap-6">
            <section data-testid="backlog-scm-backlog" className={STACK}>
              <h4 className="text-sm font-medium">{messages.providerBacklog}</h4>
              <p className="text-sm text-muted-foreground">{messages.scmBacklogGitHelp}</p>
              {readOnly ? null : backlogGit}
            </section>
            {error ? <ListError testId="backlog-scm" notice={error} onRetry={load} /> : null}
            {(views ?? []).map((v) => (
              <ProviderCard
                key={v.provider}
                workspaceId={workspaceId}
                view={v}
                selectedProjects={selectedProjects}
                readOnly={readOnly}
                onView={onView}
              />
            ))}
          </div>
        </SettingsSection>
      </div>
    );
  };
}
