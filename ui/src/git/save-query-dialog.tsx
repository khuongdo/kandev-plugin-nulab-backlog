import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { BUTTON, FIELD, STACK } from "../layout";
import type { IssueQuery } from "../issues/issues-state";
import { en, format, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";
import { gitNotice, noticeText, providerName, stateText } from "./git-state";

/** A saved PR query (git.queries.*). */
export interface Query {
  id?: string;
  name: string;
  projectKey: string;
  repoName: string;
  statuses: string[];
  assignee: string;
  creator?: string;
  isDefault?: boolean;
}

/** A saved query of GitHub, GitLab or Bitbucket (scm.queries.*, FR4.2). */
export interface ScmQuery {
  id?: string;
  name: string;
  provider: string;
  projectKey: string;
  repo: string;
  statuses: string[];
  author: string;
  isDefault?: boolean;
  /** The repository is no longer mapped (FR3.4). */
  unmapped?: boolean;
}

/** "GitHub · PROJ · acme/web · Open · author Me". */
export function scmQueryFilters(q: ScmQuery, messages: Messages = en): string {
  return format(messages.scmQueryFilters, {
    provider: providerName(q.provider, messages),
    project: q.projectKey,
    repo: q.repo,
    statuses: q.statuses.map((s) => stateText(s, messages)).join(", "),
    author: q.author === "me" ? messages.whoMe : messages.whoAnyone,
  });
}

/** "PROJ / web-app · Open, Merged · assignee Me · creator Anyone". */
export function queryFilters(q: Query, messages: Messages = en): string {
  const who = (v?: string) => (v === "me" ? messages.whoMe : messages.whoAnyone);
  return format(messages.queryFilters, {
    repo: format(messages.watchRepo, { project: q.projectKey, repo: q.repoName }),
    statuses: q.statuses.length
      ? q.statuses.map((s) => stateText(s, messages)).join(", ")
      : messages.statusesAll,
    assignee: who(q.assignee),
    creator: who(q.creator),
  });
}

export interface SaveQueryDialogProps {
  workspaceId: string;
  /** The filters to save; with an id the query is renamed (Settings, WF6). */
  query: Query | IssueQuery | ScmQuery;
  /** The save action: git.queries.save (default), issues.queries.save (FR4.3) or scm.queries.save. */
  action?: string;
  /** The filter summary; a PR query's by default. */
  description?: string;
  // Method syntax: callers handle the query type of their own action.
  onSaved(query: Query | IssueQuery | ScmQuery): void;
  onClose: () => void;
}

/** Saves a list's filters as a query, or renames one (BR2.5, FR4.3). */
export function createSaveQueryDialog(
  host: PluginHostApi,
  messages: Messages = en,
): Component<SaveQueryDialogProps> {
  const h = host.jsx;
  const { useState } = host.React;
  const {
    Button,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Label,
  } = hostUi(host);
  const ID = "backlog-save-query";

  return function SaveQueryDialog({
    workspaceId,
    query,
    action = "git.queries.save",
    description,
    onSaved,
    onClose,
  }: SaveQueryDialogProps) {
    const [name, setName] = useState(query.name);
    const [saving, setSaving] = useState(false);
    const [nameError, setNameError] = useState(false);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);

    const save = async () => {
      const trimmed = name.trim();
      setNameError(trimmed === "" || trimmed.length > 100);
      setNotice(undefined);
      if (trimmed === "" || trimmed.length > 100 || saving) return;
      setSaving(true);
      try {
        onSaved(
          await host.api.invokeAction<Query | IssueQuery | ScmQuery>(action, {
            workspaceId,
            body: { ...query, name: trimmed },
          }),
        );
      } catch (e) {
        const f = readFailure(e);
        setNameError(f.code === "validation" && f.field === "name");
        if (!(f.code === "validation" && f.field === "name")) setNotice(gitNotice(e));
        setSaving(false);
      }
    };

    return (
      <Dialog open onOpenChange={(open: boolean) => !open && !saving && onClose()}>
        <DialogContent
          data-testid={`${ID}-dialog`}
          aria-labelledby={`${ID}-title`}
          aria-describedby={`${ID}-filters`}
        >
          <DialogHeader>
            <DialogTitle id={`${ID}-title`}>
              {query.id ? messages.editQueryTitle : messages.saveQueryTitle}
            </DialogTitle>
            <DialogDescription id={`${ID}-filters`}>
              {description ?? queryFilters(query as Query, messages)}
            </DialogDescription>
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
                aria-invalid={nameError ? "true" : undefined}
                aria-describedby={nameError ? `${ID}-name-error` : undefined}
                onChange={(e: { target: { value: string } }) => setName(e.target.value)}
              />
              {nameError ? (
                <p id={`${ID}-name-error`} data-testid={`${ID}-name-error`}>
                  {messages.errorName}
                </p>
              ) : null}
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
