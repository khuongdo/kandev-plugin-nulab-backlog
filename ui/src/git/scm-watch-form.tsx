import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";
import {
  noticeText,
  scmNotice,
  scmRepoOptions,
  splitScmRepo,
  type ProviderView,
  type ScmProvider,
} from "./git-state";

/** A watch of GitHub, GitLab or Bitbucket as scm.watches.* returns it (FR4.3). */
export interface ScmWatch {
  id: string;
  name: string;
  provider: ScmProvider;
  projectKey: string;
  repo: string;
  statuses: string[];
  author: string;
  intervalMinutes: number;
  state: string;
  createdCount: number;
  lastRunAt?: string;
  unmapped?: boolean;
}

export interface ScmWatchFormProps {
  /** React key: one form per provider. */
  key?: string;
  workspaceId: string;
  provider: ScmProvider;
  view?: ProviderView;
  selectedProjects: string[];
  watch?: Partial<ScmWatch>;
  /** The provider selector, shown first. */
  providerControl?: unknown;
  onSaved: (watch: ScmWatch) => void;
  onCancel: () => void;
}

type FormField = "name" | "repo" | "statuses" | "interval";

const IDS: Record<FormField, string> = {
  name: "backlog-watch-name",
  repo: "backlog-watch-repo",
  statuses: "backlog-watch-statuses",
  interval: "backlog-watch-interval",
};
const ERRORS: Record<FormField, MessageKey> = {
  name: "errorName",
  repo: "errorRepository",
  statuses: "errorStatuses",
  interval: "errorInterval",
};
const SERVER_FIELDS: Record<string, FormField> = {
  name: "name",
  repository: "repo",
  statuses: "statuses",
  intervalMinutes: "interval",
};
const STATUSES = ["open", "closed", "merged"] as const;
const STATUS_KEYS: Record<string, MessageKey> = {
  open: "stateOpen",
  closed: "stateClosed",
  merged: "stateMerged",
};

/** Create or edit a watch of another provider: at most one task per run, every N minutes (FR4.3). */
export function createScmWatchForm(
  host: PluginHostApi,
  messages: Messages = en,
): Component<ScmWatchFormProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const { Button, Checkbox, Input, Label, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } =
    hostUi(host);

  return function ScmWatchForm(props: ScmWatchFormProps) {
    const { workspaceId, provider, view, selectedProjects, watch, providerControl, onSaved, onCancel } =
      props;
    const [name, setName] = useState(watch?.name ?? "");
    const [repo, setRepo] = useState(watch?.projectKey ? `${watch.projectKey}:${watch.repo}` : "");
    const [statuses, setStatuses] = useState<string[]>(watch?.statuses ?? ["open"]);
    const [author, setAuthor] = useState(watch?.author ?? "anyone");
    const [minutesText, setMinutesText] = useState(String(watch?.intervalMinutes ?? 5));
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<FormField | undefined>(undefined);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);
    const repos = scmRepoOptions(view, selectedProjects, messages);
    if (repo && !repos.some((r) => r.value === repo)) repos.push({ value: repo, label: repo });

    useEffect(() => {
      if (error) document.getElementById(error === "statuses" ? `${IDS.statuses}-open` : IDS[error])?.focus();
    }, [error]);

    const minutes = Number(minutesText);
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
              : !Number.isInteger(minutes) || minutes < 1 || minutes > 1440
                ? "interval"
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
        const saved = await host.api.invokeAction<ScmWatch>("scm.watches.save", {
          workspaceId,
          body: {
            ...(watch?.id ? { id: watch.id } : {}),
            name: trimmed,
            provider,
            ...splitScmRepo(repo),
            statuses,
            author,
            intervalMinutes: minutes,
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
        if (!field) setNotice(f.field === "workflowId" ? { key: "errorWorkflow" } : scmNotice(err));
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
    const choice = (
      id: string,
      label: string,
      value: string,
      set: (v: string) => void,
      items: [string, string][],
    ) => (
      <div className={FIELD}>
        <Label htmlFor={id}>{label}</Label>
        <Select value={value} onValueChange={set}>
          <SelectTrigger id={id} data-testid={id} {...(id === IDS.repo ? described("repo") : {})}>
            <SelectValue placeholder={messages.chooseRepository} />
          </SelectTrigger>
          <SelectContent>
            {items.map(([v, l]) => (
              <SelectItem key={v} value={v} data-testid={`${id}-${v}`}>
                {l}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {id === IDS.repo ? errorText("repo") : null}
      </div>
    );

    return (
      <form data-testid="backlog-watch-form" className={STACK} onSubmit={save} noValidate>
        {providerControl}
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
        {choice(
          IDS.repo,
          format(messages.required, { label: messages.watchRepoLabel }),
          repo,
          setRepo,
          repos.map((r) => [r.value, r.label]),
        )}
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
        {choice("backlog-watch-author", messages.scmAuthorLabel, author, setAuthor, [
          ["anyone", messages.whoAnyone],
          ["me", messages.whoMe],
        ])}
        <div className={FIELD}>
          <Label htmlFor={IDS.interval}>{messages.watchIntervalLabel}</Label>
          <Input
            id={IDS.interval}
            data-testid={IDS.interval}
            type="number"
            min={1}
            max={1440}
            value={minutesText}
            onChange={(e: { target: { value: string } }) => setMinutesText(e.target.value)}
            {...described("interval")}
          />
          {errorText("interval")}
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
