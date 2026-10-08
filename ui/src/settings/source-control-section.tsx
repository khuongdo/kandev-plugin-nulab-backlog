import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import {
  noticeText,
  providerName,
  scmNotice,
  type CliAccount,
  type ProviderView,
  type ScmProvider,
} from "../git/git-state";
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
/** The CLI whose login a provider can use (FR1.1, FR2.1); Bitbucket has none. */
const CLIS: Partial<Record<ScmProvider, string>> = { github: "gh", gitlab: "glab" };
const TEST_ERRORS: Record<string, MessageKey> = {
  cli_unavailable: "scmCliUnavailable",
  invalid_token: "scmErrorInvalidToken",
  missing_scope: "scmErrorMissingScope",
  rate_limited: "scmErrorRateLimited",
  unreachable: "scmErrorUnreachable",
};

type OnView = (view: ProviderView) => void;
/** A card message; failures are alerts (AC1.1.3). */
type Shown = { notice: Notice; alert?: boolean };

/** The account-missing notice, naming the chosen login when there is one (AC3.1.1, AC3.2.4). */
const missing = (login?: string): Notice =>
  login ? { key: "scmCliAccountMissing", params: { login } } : { key: "scmCliPickAccount" };

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
  const {
    Badge,
    Button,
    Input,
    Label,
    SettingsSection,
    Select,
    SelectTrigger,
    SelectValue,
    SelectContent,
    SelectItem,
  } = hostUi(host);
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
    const hasToken = view.state !== "not_configured";
    const cli = CLIS[p];
    /** GitHub connected through gh: it has a chosen login (intent 261008-gh-cli-profile). */
    const viaGh = p === "github" && view.method === "cli" && hasToken;
    const [shown, setShown] = useState<Shown | undefined>(() =>
      viaGh && view.lastError === "cli_account_missing"
        ? { notice: missing(view.login), alert: true }
        : undefined,
    );
    /** The gh accounts of the open picker, and the picked login (FR2.2). */
    const [accounts, setAccounts] = useState<CliAccount[] | undefined>(undefined);
    const [pick, setPick] = useState("");
    const withCli = (n: Notice): Notice => (cli ? { ...n, params: { ...n.params, cli } } : n);

    const run = async (key: string, body: Record<string, string>, ok: (v: ProviderView) => Notice) => {
      if (busy) return;
      setBusy(true);
      setShown(undefined);
      try {
        const v = await invoke<ProviderView>(key, workspaceId, body);
        onView(v);
        setShown({ notice: withCli(ok(v)), alert: Boolean(v.lastError) });
      } catch (e) {
        const n =
          readFailure(e).code === "cli_account_missing" ? missing(body.login ?? view.login) : scmNotice(e);
        setShown({ notice: withCli(n), alert: true });
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
      run("scm.providers.test", { provider: p }, (v) =>
        v.lastError === "cli_account_missing"
          ? missing(v.login)
          : { key: v.lastError ? (TEST_ERRORS[v.lastError] ?? "scmErrorUnreachable") : "scmTestOk" },
      );
    const remove = () => run("scm.providers.remove", { provider: p }, () => ({ key: "scmTokenRemoved" }));
    const connect = (login?: string) => {
      setAccounts(undefined);
      return run("scm.providers.use_cli", login ? { provider: p, login } : { provider: p }, () => ({
        key: "scmCliConnected",
      }));
    };
    /**
     * Loads gh's accounts into the picker (AC1.1.1). A first connect with one
     * account connects at once (AC1.1.4); a change preselects the current
     * login only while gh still has it (AC1.2.1, AC3.1.4).
     */
    const openPicker = async (changing: boolean) => {
      if (busy) return;
      setBusy(true);
      setShown({ notice: { key: "scmLoadingAccounts" } });
      let list: CliAccount[] | undefined;
      try {
        list =
          (await invoke<{ accounts?: CliAccount[] }>("scm.providers.cli_accounts", workspaceId))?.accounts ??
          [];
        setShown(undefined);
      } catch (e) {
        setShown({ notice: withCli(scmNotice(e)), alert: true });
      }
      setBusy(false);
      if (!list) return;
      if (!changing && list.length === 1) return void connect(list[0]!.login);
      const current = changing ? view.login : list.find((a) => a.active)?.login;
      setPick(list.some((a) => a.login === current) ? current! : "");
      setAccounts(list);
    };
    const useCli = () => (p === "github" ? openPicker(false) : connect());
    const button = (
      suffix: string,
      label: string,
      onClick: () => void,
      variant = "outline",
      disabled = false,
    ) => (
      <Button
        type="button"
        variant={variant}
        size="sm"
        className={BUTTON}
        data-testid={`${id}-${suffix}`}
        disabled={busy || disabled}
        onClick={onClick}
      >
        {label}
      </Button>
    );
    const accountLine = () => {
      const name = view.account ?? "";
      if (view.method !== "cli" || !cli) return format(messages.scmAccount, { name });
      if (!view.login) return format(messages.scmAccountCli, { cli, name });
      const login = view.login;
      return name && name !== login
        ? format(messages.scmAccountCliLoginName, { cli, login, name })
        : format(messages.scmAccountCliLogin, { cli, login });
    };
    const picker = accounts ? (
      <div className={FIELD} data-testid={`${id}-account-picker`}>
        <Label htmlFor={`${id}-account-select`}>{messages.scmGhAccountLabel}</Label>
        <Select value={pick} onValueChange={setPick}>
          <SelectTrigger id={`${id}-account-select`} data-testid={`${id}-account-select`} autoFocus>
            <SelectValue placeholder={messages.scmGhAccountPlaceholder} />
          </SelectTrigger>
          <SelectContent>
            {accounts.map((a) => (
              <SelectItem key={a.login} value={a.login} data-testid={`${id}-account-${a.login}`}>
                {a.active ? format(messages.scmGhAccountActive, { login: a.login }) : a.login}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <div className={ROW}>
          {button("account-connect", messages.connect, () => void connect(pick), "default", !pick)}
          {button("account-cancel", messages.cancel, () => setAccounts(undefined))}
        </div>
      </div>
    ) : null;
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
        {hasToken && (view.account || view.login) ? (
          <p data-testid={`${id}-account`}>{accountLine()}</p>
        ) : null}
        {viaGh && view.login ? (
          <p data-testid={`${id}-worktree-note`} className="text-sm text-muted-foreground">
            {format(messages.scmWorktreeNote, { login: view.login })}
          </p>
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
              {cli ? button("use-cli", format(messages.scmUseCli, { cli }), () => void useCli()) : null}
              {viaGh
                ? button("change-account", messages.scmChangeAccount, () => void openPicker(true))
                : null}
              {hasToken ? button("test", messages.scmTest, () => void test()) : null}
              {hasToken ? button("remove", messages.scmRemoveToken, () => void remove()) : null}
            </div>
            {picker}
          </div>
        )}
        {shown ? (
          <p role={shown.alert ? "alert" : "status"} data-testid={`${id}-message`}>
            {t(shown.notice)}
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
