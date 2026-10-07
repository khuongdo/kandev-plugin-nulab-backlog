import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { gitNotice, noticeText } from "../git/git-state";
import { hostUi } from "../host-ui";
import { BUTTON, FIELD, ROW, STACK } from "../layout";
import { en, format, type Messages } from "../messages/en";
import {
  NO_ACTIONS,
  QUICK_ACTION_ICONS,
  type QuickAction,
  type QuickActionKind,
  type QuickActions,
} from "../page/quick-actions";
import { createSectionParts } from "./section-parts";
import type { Notice } from "./state";

const MAX_ACTIONS = 20;
const MAX_LABEL = 100;
const MAX_HINT = 100;
const MAX_PROMPT = 4000;
const ID = "backlog-quick-action";

/**
 * The "Quick actions" settings section (FR2.3, FR2.4): one tab per kind with
 * an editor per action (icon, label, hint, prompt template), Add, Delete,
 * its own Save, and Reset, which saves an empty list so the server's
 * defaults come back.
 */
export function createQuickActionsSection(
  host: PluginHostApi,
  messages: Messages = en,
): Component<{ workspaceId: string }> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const ui = hostUi(host);
  const { Button, Card, CardContent, Input, Label, SettingsSection, Tabs, TabsList, TabsTrigger, Textarea } =
    ui;
  const { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } = ui;
  const { ListError } = createSectionParts(host, messages);

  return function QuickActionsSection({ workspaceId }: { workspaceId: string }) {
    const [actions, setActions] = useState<QuickActions>(NO_ACTIONS);
    const [kind, setKind] = useState<QuickActionKind>("issue");
    const [loadError, setLoadError] = useState<Notice | undefined>(undefined);
    const [nonce, setNonce] = useState(0);
    const [labelErrors, setLabelErrors] = useState<number[]>([]);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);
    const [saved, setSaved] = useState(false);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
      setLoadError(undefined);
      host.api
        .invokeAction<QuickActions>("issues.quick_actions.get", { workspaceId })
        .then((r) => setActions({ issue: r?.issue ?? [], pr: r?.pr ?? [] }))
        .catch((e: unknown) => setLoadError(gitNotice(e)));
    }, [workspaceId, nonce]);

    const list = actions[kind];
    const setList = (next: QuickAction[]) => {
      setActions((a) => ({ ...a, [kind]: next }));
      setSaved(false);
    };
    const edit = (i: number, patch: Partial<QuickAction>) =>
      setList(list.map((a, j) => (j === i ? { ...a, ...patch } : a)));

    const send = async (next: QuickAction[]) => {
      setNotice(undefined);
      setSaved(false);
      setSaving(true);
      try {
        const r = await host.api.invokeAction<QuickActions>("issues.quick_actions.save", {
          workspaceId,
          body: { kind, actions: next },
        });
        setActions({ issue: r?.issue ?? [], pr: r?.pr ?? [] });
        setSaved(true);
      } catch (e) {
        setNotice(gitNotice(e));
      }
      setSaving(false);
    };

    const save = () => {
      const bad = list.flatMap((a, i) => {
        const n = Array.from(a.label.trim()).length;
        return n === 0 || n > MAX_LABEL ? [i] : [];
      });
      setLabelErrors(bad);
      if (bad.length === 0) void send(list);
    };
    const reset = () => {
      setLabelErrors([]);
      void send([]);
    };

    const editor = (a: QuickAction, i: number) => (
      <Card key={i} data-testid={`${ID}-${i}`}>
        <CardContent className={`${STACK} p-4`}>
          <div className={ROW}>
            <div className={FIELD}>
              <Label htmlFor={`${ID}-icon-${i}`}>{messages.quickActionIcon}</Label>
              <Select value={a.icon} onValueChange={(v: string) => edit(i, { icon: v })}>
                <SelectTrigger id={`${ID}-icon-${i}`} data-testid={`${ID}-icon-${i}`} className="min-w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {QUICK_ACTION_ICONS.map((icon) => (
                    <SelectItem key={icon} value={icon} data-testid={`${ID}-icon-${i}-${icon}`}>
                      {icon}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className={`${FIELD} flex-1`}>
              <Label htmlFor={`${ID}-label-${i}`}>
                {format(messages.required, { label: messages.quickActionLabel })}
              </Label>
              <Input
                id={`${ID}-label-${i}`}
                data-testid={`${ID}-label-${i}`}
                value={a.label}
                aria-invalid={labelErrors.includes(i) ? "true" : undefined}
                aria-describedby={labelErrors.includes(i) ? `${ID}-label-error-${i}` : undefined}
                onChange={(e: { target: { value: string } }) => edit(i, { label: e.target.value })}
              />
              {labelErrors.includes(i) ? (
                <p id={`${ID}-label-error-${i}`} data-testid={`${ID}-label-error-${i}`}>
                  {messages.errorQuickActionLabel}
                </p>
              ) : null}
            </div>
            <div className={`${FIELD} flex-1`}>
              <Label htmlFor={`${ID}-hint-${i}`}>{messages.quickActionHint}</Label>
              <Input
                id={`${ID}-hint-${i}`}
                data-testid={`${ID}-hint-${i}`}
                value={a.hint}
                maxLength={MAX_HINT}
                onChange={(e: { target: { value: string } }) => edit(i, { hint: e.target.value })}
              />
            </div>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className={`${BUTTON} self-end`}
              data-testid={`${ID}-delete-${i}`}
              aria-label={format(messages.deleteQuickAction, { name: a.label })}
              onClick={() => setList(list.filter((_, j) => j !== i))}
            >
              {messages.delete}
            </Button>
          </div>
          <div className={FIELD}>
            <Label htmlFor={`${ID}-prompt-${i}`}>{messages.quickActionPrompt}</Label>
            <Textarea
              id={`${ID}-prompt-${i}`}
              data-testid={`${ID}-prompt-${i}`}
              value={a.promptTemplate}
              maxLength={MAX_PROMPT}
              rows={3}
              onChange={(e: { target: { value: string } }) => edit(i, { promptTemplate: e.target.value })}
            />
          </div>
        </CardContent>
      </Card>
    );

    const body = loadError ? (
      <ListError testId="backlog-quick-actions" notice={loadError} onRetry={() => setNonce((n) => n + 1)} />
    ) : (
      <div className={STACK}>
        <Tabs value={kind} onValueChange={(v: QuickActionKind) => setKind(v)}>
          <TabsList aria-label={messages.sectionQuickActions}>
            <TabsTrigger value="issue" className={BUTTON} data-testid="backlog-quick-actions-tab-issue">
              {messages.scopeIssues}
            </TabsTrigger>
            <TabsTrigger value="pr" className={BUTTON} data-testid="backlog-quick-actions-tab-pr">
              {messages.scopePRs}
            </TabsTrigger>
          </TabsList>
        </Tabs>
        {list.map(editor)}
        <div className={ROW}>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className={BUTTON}
            data-testid="backlog-quick-actions-add"
            disabled={list.length >= MAX_ACTIONS}
            onClick={() =>
              setList([
                ...list,
                { id: "", label: messages.newQuickAction, hint: "", icon: "sparkle", promptTemplate: "" },
              ])
            }
          >
            {messages.addQuickAction}
          </Button>
          <Button
            type="button"
            size="sm"
            className={BUTTON}
            data-testid="backlog-quick-actions-save"
            disabled={saving}
            onClick={save}
          >
            {saving ? messages.saving : messages.save}
          </Button>
          {saved ? (
            <span role="status" data-testid="backlog-quick-actions-status">
              {messages.quickActionsSaved}
            </span>
          ) : null}
        </div>
        {notice ? (
          <p role="alert" data-testid="backlog-quick-actions-notice">
            {noticeText(notice, messages)}
          </p>
        ) : null}
      </div>
    );

    return (
      <div data-testid="backlog-section-quick-actions">
        <SettingsSection
          title={messages.sectionQuickActions}
          description={format(messages.quickActionsDescription, { url: "{{url}}", title: "{{title}}" })}
          action={
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={BUTTON}
              data-testid="backlog-quick-actions-reset"
              disabled={saving || Boolean(loadError)}
              onClick={reset}
            >
              {messages.resetQuickActions}
            </Button>
          }
        >
          {body}
        </SettingsSection>
      </div>
    );
  };
}
