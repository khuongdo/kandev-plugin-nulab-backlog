import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";
import { gitNotice, loadRepoOptions, noticeText, splitRepo, type RepoOption } from "./git-state";

/** A saved watch as git.watches.* returns it. */
export interface Watch {
  id: string;
  name: string;
  projectKey: string;
  repoName: string;
  statuses: string[];
  assignee: string;
  creator: string;
  issueKey?: string;
  state: string;
  createdCount: number;
  pendingCount: number;
  lastRunAt?: string;
}

export interface WatchFormProps {
  workspaceId: string;
  watch?: Partial<Watch>;
  onSaved: (watch: Watch) => void;
  onCancel: () => void;
}

type FormField = "name" | "repo" | "statuses" | "issue";

const IDS: Record<FormField, string> = {
  name: "backlog-watch-name",
  repo: "backlog-watch-repo",
  statuses: "backlog-watch-statuses",
  issue: "backlog-watch-issue",
};
const ERRORS: Record<FormField, MessageKey> = {
  name: "errorName",
  repo: "errorRepository",
  statuses: "errorStatuses",
  issue: "errorIssueKey",
};
const SERVER_FIELDS: Record<string, FormField> = {
  name: "name",
  repository: "repo",
  statuses: "statuses",
  issueKey: "issue",
};
const STATUSES = ["open", "closed", "merged"] as const;
const STATUS_KEYS: Record<string, MessageKey> = {
  open: "stateOpen",
  closed: "stateClosed",
  merged: "stateMerged",
};

/** Create or edit a PR watch (M4, AC6.1.1, AC6.1.2), shown in the PR watches dialog. */
export function createWatchForm(host: PluginHostApi, messages: Messages = en): Component<WatchFormProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const { Button, Input, Label, Checkbox, Select, SelectTrigger, SelectValue, SelectContent, SelectItem } =
    hostUi(host);

  return function WatchForm({ workspaceId, watch, onSaved, onCancel }: WatchFormProps) {
    const [repos, setRepos] = useState<RepoOption[]>([]);
    const [name, setName] = useState(watch?.name ?? "");
    const [repo, setRepo] = useState(watch?.projectKey ? `${watch.projectKey}/${watch.repoName}` : "");
    const [statuses, setStatuses] = useState<string[]>(watch?.statuses ?? ["open"]);
    const [assignee, setAssignee] = useState(watch?.assignee ?? "anyone");
    const [creator, setCreator] = useState(watch?.creator ?? "anyone");
    const [issue, setIssue] = useState(watch?.issueKey ?? "");
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<FormField | undefined>(undefined);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);

    useEffect(() => {
      loadRepoOptions(host, workspaceId, messages).then(setRepos, (e: unknown) => setNotice(gitNotice(e)));
    }, [workspaceId]);
    useEffect(() => {
      if (error) document.getElementById(error === "statuses" ? `${IDS.statuses}-open` : IDS[error])?.focus();
    }, [error]);

    const save = async (e?: { preventDefault(): void }) => {
      e?.preventDefault();
      if (saving) return;
      const trimmed = name.trim();
      const local: FormField | undefined =
        trimmed === "" || trimmed.length > 100
          ? "name"
          : !repo
            ? "repo"
            : statuses.length === 0
              ? "statuses"
              : undefined;
      setError(local);
      setNotice(undefined);
      if (local) return;
      const flow = host.context.getTaskCreationContext(workspaceId);
      if (!flow?.workflowId) {
        setNotice({ key: "errorWorkflow" });
        return;
      }
      setSaving(true);
      try {
        const saved = await host.api.invokeAction<Watch>("git.watches.save", {
          workspaceId,
          body: {
            ...(watch?.id ? { id: watch.id } : {}),
            name: trimmed,
            ...splitRepo(repo),
            statuses,
            assignee,
            creator,
            ...(issue.trim() ? { issueKey: issue.trim() } : {}),
            workflowId: flow.workflowId,
            workflowStepId: flow.defaultStepId,
          },
        });
        setSaving(false);
        onSaved(saved);
      } catch (err) {
        setSaving(false);
        const f = readFailure(err);
        const field = f.code === "validation" && f.field ? SERVER_FIELDS[f.field] : undefined;
        setError(field);
        if (!field) setNotice(f.field === "workflowId" ? { key: "errorWorkflow" } : gitNotice(err));
      }
    };

    const described = (field: FormField) =>
      error === field ? { "aria-invalid": "true", "aria-describedby": `${IDS[field]}-error` } : {};
    const errorText = (field: FormField) =>
      error === field ? (
        <p id={`${IDS[field]}-error`} data-testid={`${IDS[field]}-error`}>
          {messages[ERRORS[field]]}
        </p>
      ) : null;
    const who = (id: string, label: string, value: string, set: (v: string) => void) => (
      <div className={FIELD}>
        <Label htmlFor={id}>{label}</Label>
        <Select value={value} onValueChange={set}>
          <SelectTrigger id={id} data-testid={id}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="anyone" data-testid={`${id}-anyone`}>
              {messages.whoAnyone}
            </SelectItem>
            <SelectItem value="me" data-testid={`${id}-me`}>
              {messages.whoMe}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    );

    return (
      <form data-testid="backlog-watch-form" className={STACK} onSubmit={save} noValidate>
        <div className={FIELD}>
          <Label htmlFor={IDS.name}>{format(messages.required, { label: messages.watchNameLabel })}</Label>
          <Input
            id={IDS.name}
            data-testid={IDS.name}
            value={name}
            maxLength={100}
            onChange={(e: { target: { value: string } }) => setName(e.target.value)}
            {...described("name")}
          />
          {errorText("name")}
        </div>
        <div className={FIELD}>
          <Label htmlFor={IDS.repo}>{format(messages.required, { label: messages.watchRepoLabel })}</Label>
          <Select value={repo} onValueChange={setRepo}>
            <SelectTrigger id={IDS.repo} data-testid={IDS.repo} {...described("repo")}>
              <SelectValue placeholder={messages.chooseRepository} />
            </SelectTrigger>
            <SelectContent>
              {repos.map((r) => (
                <SelectItem key={r.value} value={r.value} data-testid={`${IDS.repo}-${r.value}`}>
                  {r.label}
                </SelectItem>
              ))}
              {repo && !repos.some((r) => r.value === repo) ? (
                <SelectItem value={repo} data-testid={`${IDS.repo}-current`}>
                  {repo}
                </SelectItem>
              ) : null}
            </SelectContent>
          </Select>
          {errorText("repo")}
        </div>
        <fieldset
          className={FIELD}
          {...(error === "statuses" ? { "aria-describedby": `${IDS.statuses}-error` } : {})}
        >
          <legend>{messages.watchStatusLegend}</legend>
          {STATUSES.map((s) => (
            <div key={s} className={ROW}>
              <Checkbox
                id={`${IDS.statuses}-${s}`}
                data-testid={`backlog-watch-status-${s}`}
                checked={statuses.includes(s)}
                onCheckedChange={() =>
                  setStatuses((l) => (l.includes(s) ? l.filter((x) => x !== s) : [...l, s]))
                }
              />
              <Label htmlFor={`${IDS.statuses}-${s}`}>{messages[STATUS_KEYS[s]!]}</Label>
            </div>
          ))}
          {errorText("statuses")}
        </fieldset>
        {who("backlog-watch-assignee", messages.assigneeLabel, assignee, setAssignee)}
        {who("backlog-watch-creator", messages.creatorLabel, creator, setCreator)}
        <div className={FIELD}>
          <Label htmlFor={IDS.issue}>{messages.linkedIssueLabel}</Label>
          <Input
            id={IDS.issue}
            data-testid={IDS.issue}
            value={issue}
            placeholder={messages.issueKeyPlaceholder}
            onChange={(e: { target: { value: string } }) => setIssue(e.target.value)}
            {...described("issue")}
          />
          {errorText("issue")}
        </div>
        {notice ? <p data-testid="backlog-watch-form-notice">{noticeText(notice, messages)}</p> : null}
        <div className={ROW}>
          <Button
            type="button"
            variant="outline"
            className={BUTTON}
            data-testid="backlog-watch-cancel"
            onClick={onCancel}
          >
            {messages.cancel}
          </Button>
          <Button type="submit" className={BUTTON} data-testid="backlog-watch-save" disabled={saving}>
            {saving ? messages.saving : messages.save}
          </Button>
        </div>
      </form>
    );
  };
}
