import type { Component, PluginHostApi, TaskCreationContext } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { en, format, type Messages } from "../messages/en";
import { interpolate, taskTitle, type QuickAction, type QuickActionKind } from "./quick-actions";

export interface StartTaskProps {
  workspaceId: string;
  kind: QuickActionKind;
  /** The issue summary or PR title. */
  title: string;
  /** The issue or PR page on Backlog: {{url}}, and the PR link reference. */
  url: string;
  /** The issue to link (kind "issue"). */
  issueKey?: string;
  actions: QuickAction[];
  /** Prefix of the trigger and item test ids. */
  testId: string;
  /** The task is linked; taskKey comes from the link reply when Kandev gave one. */
  onLinked?: (taskId: string, taskKey?: string) => void;
  /** A PR of another provider links with scm.prs.link and its URL (FR5.1). */
  linkAction?: "scm.prs.link";
}

/**
 * The row's "+ Task" menu (FR1): a quick action opens Kandev's own
 * TaskCreateDialog prefilled from its prompt template; once Kandev creates
 * the task it is linked to the issue or PR. A failed link keeps the task.
 *
 * Kandev's browser context is null unless its store already holds the
 * workflow's steps (e.g. /backlog opened directly), so the plugin then asks
 * its backend whether the workspace has a workflow and lets the dialog pick
 * it (intent 261009, FR1.2). A failed check still opens the dialog (FR1.4).
 */
export function createStartTask(host: PluginHostApi, messages: Messages = en): Component<StartTaskProps> {
  const h = host.jsx;
  const { useState, useRef } = host.React;
  const { IntegrationStartTaskMenu, TaskCreateDialog } = hostUi(host);

  return function StartTask({
    workspaceId,
    kind,
    title,
    url,
    issueKey,
    actions,
    testId,
    onLinked,
    linkAction,
  }: StartTaskProps) {
    const [open, setOpen] = useState<{ action: QuickAction; ctx: TaskCreationContext | null } | undefined>(
      undefined,
    );
    const checking = useRef(false);

    const hasWorkflow = async () => {
      try {
        const reply = await host.api.invokeAction<{ hasWorkflow?: boolean }>("workflows.status", {
          workspaceId,
        });
        return reply?.hasWorkflow !== false;
      } catch {
        return true; // FR1.4: a failed check must not bring the original bug back
      }
    };

    const select = async (preset: { id: string }) => {
      const action = actions.find((a) => a.id === preset.id);
      if (!action || checking.current) return;
      const ctx = host.context.getTaskCreationContext?.(workspaceId) ?? null;
      if (!ctx) {
        checking.current = true;
        const ok = await hasWorkflow();
        checking.current = false;
        if (!ok) {
          host.toast.error(messages.errorWorkflow);
          return;
        }
      }
      setOpen({ action, ctx });
    };

    const link = async (taskId: string) => {
      try {
        const [action, body]: [string, unknown] =
          kind === "issue"
            ? ["issues.link", { issueKey }]
            : linkAction
              ? [linkAction, { url }]
              : ["git.prs.link", { reference: url }];
        const reply = await host.api.invokeAction<{ taskKey?: string }>(action, {
          workspaceId,
          taskId,
          body,
        });
        onLinked?.(taskId, reply?.taskKey);
      } catch {
        host.toast.error(messages.taskNotLinked);
      }
    };

    return (
      <span className="flex items-center gap-2">
        <IntegrationStartTaskMenu
          presets={actions.map((a) => ({ id: a.id, label: a.label, hint: a.hint, iconName: a.icon }))}
          onSelect={(preset: { id: string }) => void select(preset)}
          triggerLabel={messages.startTask}
          triggerAriaLabel={format(messages.startTaskLabel, { name: issueKey ?? title })}
          triggerTestId={`${testId}-start`}
          itemTestId={`${testId}-start-item`}
        />
        {open ? (
          <TaskCreateDialog
            open
            onOpenChange={(o: boolean) => !o && setOpen(undefined)}
            workspaceId={workspaceId}
            workflowId={open.ctx?.workflowId ?? null}
            defaultStepId={open.ctx?.defaultStepId ?? null}
            steps={open.ctx?.steps ?? []}
            initialValues={{
              title: taskTitle(open.action.label, title),
              description: interpolate(open.action.promptTemplate, { url, title }),
            }}
            onSuccess={(task: { id: string }) => {
              setOpen(undefined);
              void link(task.id);
            }}
          />
        ) : null}
      </span>
    );
  };
}
