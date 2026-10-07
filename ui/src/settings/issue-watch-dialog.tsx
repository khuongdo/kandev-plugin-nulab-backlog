import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { gitNotice, noticeText } from "../git/git-state";
import { hostUi } from "../host-ui";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type MessageKey, type Messages } from "../messages/en";
import { readFailure, type Notice } from "./state";

/** A saved issue watch as issues.watches.* returns it (FR3). */
export interface IssueWatch {
  id: string;
  name: string;
  projectKey: string;
  statusIds: number[];
  assignee: string;
  creator: string;
  workflowId: string;
  workflowStepId?: string;
  intervalMinutes: number;
  state: string;
  createdCount: number;
  pendingCount: number;
  lastRunAt?: string;
  lastError?: string;
}

/** One project or status choice from issues.filters. */
export interface FilterOption {
  id?: number;
  key?: string;
  name: string;
}

export interface IssueWatchDialogProps {
  workspaceId: string;
  /** Undefined adds a watch. */
  initial?: IssueWatch;
  projects: FilterOption[];
  statuses: FilterOption[];
  onSaved: (watch: IssueWatch) => void;
  onClose: () => void;
}

type Field = "name" | "project" | "statuses" | "interval";

const ID = "backlog-issue-watch";
const ERRORS: Record<Field, MessageKey> = {
  name: "errorName",
  project: "errorProject",
  statuses: "errorStatuses",
  interval: "errorInterval",
};
const SERVER_FIELDS: Record<string, Field> = {
  name: "name",
  projectKey: "project",
  statusIds: "statuses",
  intervalMinutes: "interval",
};

/** A whole number of minutes from 1 to 1440, or undefined (BR3.1). */
function parseMinutes(raw: string): number | undefined {
  if (!/^\d+$/.test(raw.trim())) return undefined;
  const n = Number(raw);
  return n >= 1 && n <= 1440 ? n : undefined;
}

/** Add or edit an issue watch (WF2, BR3.1, BR3.2). */
export function createIssueWatchDialog(
  host: PluginHostApi,
  messages: Messages = en,
): Component<IssueWatchDialogProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const ui = hostUi(host);
  const { Button, Checkbox, Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, Input, Label } =
    ui;
  const { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } = ui;

  return function IssueWatchDialog({
    workspaceId,
    initial,
    projects,
    statuses,
    onSaved,
    onClose,
  }: IssueWatchDialogProps) {
    const [name, setName] = useState(initial?.name ?? "");
    const [project, setProject] = useState(initial?.projectKey ?? "");
    const [statusIds, setStatusIds] = useState<number[]>(initial?.statusIds ?? []);
    const [assignee, setAssignee] = useState(initial?.assignee ?? "anyone");
    const [creator, setCreator] = useState(initial?.creator ?? "anyone");
    const [interval, setMinutes] = useState(String(initial?.intervalMinutes ?? 5));
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<Field | undefined>(undefined);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);

    useEffect(() => {
      if (error) document.getElementById(error === "statuses" ? `${ID}-statuses` : `${ID}-${error}`)?.focus();
    }, [error]);

    const save = async () => {
      if (saving) return;
      const minutes = parseMinutes(interval);
      const trimmed = name.trim();
      const local: Field | undefined =
        trimmed === "" || trimmed.length > 100
          ? "name"
          : !project
            ? "project"
            : statusIds.length === 0 || statusIds.length > 20
              ? "statuses"
              : minutes === undefined
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
        const saved = await host.api.invokeAction<IssueWatch>("issues.watches.save", {
          workspaceId,
          body: {
            ...(initial?.id ? { id: initial.id } : {}),
            name: trimmed,
            projectKey: project,
            statusIds,
            assignee,
            creator,
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
        if (!field) setNotice(f.field === "workflowId" ? { key: "errorWorkflow" } : gitNotice(err));
      }
    };

    const described = (field: Field) =>
      error === field ? { "aria-invalid": "true", "aria-describedby": `${ID}-${field}-error` } : {};
    const errorText = (field: Field) =>
      error === field ? (
        <p id={`${ID}-${field}-error`} data-testid={`${ID}-${field}-error`}>
          {messages[ERRORS[field]]}
        </p>
      ) : null;
    const select = (
      field: string,
      label: string,
      value: string,
      set: (v: string) => void,
      options: FilterOption[],
    ) => (
      <div className={FIELD}>
        <Label htmlFor={`${ID}-${field}`}>{label}</Label>
        <Select value={value} onValueChange={set}>
          <SelectTrigger
            id={`${ID}-${field}`}
            data-testid={`${ID}-${field}`}
            {...(field === "project" ? described("project") : {})}
          >
            <SelectValue placeholder={field === "project" ? messages.chooseProject : undefined} />
          </SelectTrigger>
          <SelectContent>
            {options.map((o) => (
              <SelectItem key={o.key} value={o.key} data-testid={`${ID}-${field}-${o.key}`}>
                {o.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {field === "project" ? errorText("project") : null}
      </div>
    );
    const who = [
      { key: "anyone", name: messages.whoAnyone },
      { key: "me", name: messages.whoMe },
    ];

    return (
      <Dialog open onOpenChange={(open: boolean) => !open && !saving && onClose()}>
        <DialogContent data-testid={`${ID}-dialog`} aria-labelledby={`${ID}-dialog-title`}>
          <DialogHeader>
            <DialogTitle id={`${ID}-dialog-title`}>
              {initial ? messages.editIssueWatch : messages.addIssueWatch}
            </DialogTitle>
          </DialogHeader>
          <div className={STACK}>
            <div className={FIELD}>
              <Label htmlFor={`${ID}-name`}>
                {format(messages.required, { label: messages.watchNameLabel })}
              </Label>
              <Input
                id={`${ID}-name`}
                data-testid={`${ID}-name`}
                value={name}
                maxLength={100}
                onChange={(e: { target: { value: string } }) => setName(e.target.value)}
                {...described("name")}
              />
              {errorText("name")}
            </div>
            {select(
              "project",
              format(messages.required, { label: messages.colProject }),
              project,
              setProject,
              projects,
            )}
            <fieldset id={`${ID}-statuses`} tabIndex={-1} className={FIELD} {...described("statuses")}>
              <legend>{messages.watchStatusLegend}</legend>
              {statuses.map((s) => (
                <div key={s.id} className={ROW}>
                  <Checkbox
                    id={`${ID}-status-${s.id}`}
                    data-testid={`${ID}-status-${s.id}`}
                    checked={statusIds.includes(s.id!)}
                    onCheckedChange={() =>
                      setStatusIds((l) => (l.includes(s.id!) ? l.filter((x) => x !== s.id) : [...l, s.id!]))
                    }
                  />
                  <Label htmlFor={`${ID}-status-${s.id}`}>{s.name}</Label>
                </div>
              ))}
              {errorText("statuses")}
            </fieldset>
            {select("assignee", messages.assigneeLabel, assignee, setAssignee, who)}
            {select("creator", messages.creatorLabel, creator, setCreator, who)}
            <div className={FIELD}>
              <Label htmlFor={`${ID}-interval`}>{messages.watchIntervalLabel}</Label>
              <Input
                id={`${ID}-interval`}
                data-testid={`${ID}-interval`}
                type="number"
                min={1}
                max={1440}
                step={1}
                inputMode="numeric"
                value={interval}
                onChange={(e: { target: { value: string } }) => setMinutes(e.target.value)}
                {...described("interval")}
              />
              {errorText("interval")}
            </div>
            {notice ? <p data-testid={`${ID}-notice`}>{noticeText(notice, messages)}</p> : null}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              className={BUTTON}
              data-testid={`${ID}-cancel`}
              onClick={onClose}
            >
              {messages.cancel}
            </Button>
            <Button
              type="button"
              className={BUTTON}
              data-testid={`${ID}-save`}
              disabled={saving}
              onClick={() => void save()}
            >
              {saving ? messages.saving : messages.save}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  };
}
