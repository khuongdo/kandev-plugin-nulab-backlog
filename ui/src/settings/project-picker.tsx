import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { withImpact } from "../git/git-state";
import { loadImpactText } from "../issues/issues-state";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { failureNotice, type ConnectionView, type Notice } from "./state";

type AnyProps = Record<string, unknown>;

/** One row of connection.list_projects. */
interface ProjectItem {
  projectKey: string;
  projectId: number;
  projectName: string;
  selected: boolean;
}

export interface ProjectPickerProps {
  workspaceId: string;
  announce: (notice: Notice) => void;
  onView?: (view: ConnectionView) => void;
}

const STACK = "flex flex-col gap-4";
const FIELD = "flex flex-col gap-2";
const ROW = "flex gap-2";

/**
 * The project checkbox list with a search field (US1.7). Unselecting a
 * project in use asks first (US1.9). A rate-limited load counts down and
 * retries once (AC8.4.4); a save never retries by itself.
 */
export function createProjectPicker(
  host: PluginHostApi,
  messages: Messages = en,
): Component<ProjectPickerProps> {
  const h = host.jsx;
  const { useCallback, useEffect, useRef, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const Input = host.ui.Input as Component<AnyProps>;
  const Label = host.ui.Label as Component<AnyProps>;
  const ConfirmDialog = createConfirmDialog(host, messages);
  const t = (n: Notice) => format(messages[n.key], n.params);

  return function ProjectPicker({ workspaceId, announce, onView }: ProjectPickerProps) {
    const [items, setItems] = useState<ProjectItem[] | undefined>(undefined);
    const [checked, setChecked] = useState<string[]>([]);
    const [filter, setFilter] = useState("");
    const [wait, setWait] = useState<number | undefined>(undefined);
    const [loadFailed, setLoadFailed] = useState(false);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<Notice | undefined>(undefined);
    const [fieldError, setFieldError] = useState(false);
    const [confirming, setConfirming] = useState(false);
    const [impact, setImpact] = useState("");
    const retried = useRef(false);

    const load = useCallback(async () => {
      setLoadFailed(false);
      try {
        const reply = await host.api.invokeAction<{ projects?: ProjectItem[] }>("connection.list_projects", {
          workspaceId,
        });
        const list = Array.isArray(reply?.projects) ? reply.projects : [];
        setItems(list);
        setChecked(list.filter((p) => p.selected).map((p) => p.projectKey));
      } catch (error) {
        const f = failureNotice(error, { retrying: true });
        if (f.retryAfterSeconds !== undefined && f.notice && !retried.current) {
          retried.current = true;
          setWait(f.retryAfterSeconds);
          announce(f.notice); // once; the visible countdown below changes every second
          return;
        }
        setLoadFailed(true);
      }
    }, [workspaceId]);

    useEffect(() => {
      void load();
    }, [load]);

    const waiting = wait !== undefined;
    useEffect(() => {
      if (!waiting) return;
      const id = setInterval(() => setWait((w) => (w === undefined ? w : w - 1)), 1000);
      return () => clearInterval(id);
    }, [waiting]);

    useEffect(() => {
      if (wait !== undefined && wait <= 0) {
        setWait(undefined);
        void load();
      }
    }, [wait, load]);

    const save = async () => {
      setSaving(true);
      setMessage(undefined);
      setFieldError(false);
      let notice: Notice;
      try {
        const view = await host.api.invokeAction<ConnectionView>("connection.set_projects", {
          workspaceId,
          body: { projectKeys: checked },
        });
        setItems((list) => list?.map((p) => ({ ...p, selected: checked.includes(p.projectKey) })));
        if (view?.state) onView?.(view);
        notice = { key: "projectsSaved" };
      } catch (error) {
        const f = failureNotice(error);
        setFieldError(f.fieldError?.field === "projectKeys");
        notice = f.fieldError ? { key: f.fieldError.key } : (f.notice ?? { key: "actionFailed" });
      }
      setSaving(false);
      setMessage(notice);
      announce(notice);
    };

    const onSave = () => {
      const removed = (items ?? [])
        .filter((p) => p.selected && !checked.includes(p.projectKey))
        .map((p) => p.projectKey);
      if (removed.length === 0) return void save();
      setImpact("");
      setConfirming(true);
      // U3/U4: how many issue links, PR links and watches of those projects stop (AC1.9.1).
      void loadImpactText(host, workspaceId, removed, messages).then(setImpact);
    };

    const toggle = (key: string) =>
      setChecked((list) => (list.includes(key) ? list.filter((k) => k !== key) : [...list, key]));

    const needle = filter.trim().toLowerCase();
    const visible = (items ?? []).filter(
      (p) => p.projectName.toLowerCase().includes(needle) || p.projectKey.toLowerCase().includes(needle),
    );

    return (
      <section data-testid="backlog-projects" className={STACK}>
        {wait !== undefined ? (
          <p data-testid="backlog-projects-wait">
            {t({ key: "rateLimitedRetrying", params: { seconds: wait } })}
          </p>
        ) : null}
        {loadFailed ? (
          <div className={STACK}>
            <p data-testid="backlog-projects-load-failed">{messages.projectsLoadFailed}</p>
            <Button type="button" data-testid="backlog-projects-retry" onClick={() => void load()}>
              {messages.retry}
            </Button>
          </div>
        ) : null}
        {items ? (
          <fieldset className={STACK} aria-describedby={fieldError ? "backlog-projects-error" : undefined}>
            <legend>{messages.projectsLegend}</legend>
            {items.length === 0 ? (
              <p data-testid="backlog-projects-empty">{messages.noProjects}</p>
            ) : (
              <div className={STACK}>
                <div className={FIELD}>
                  <Label htmlFor="backlog-project-search">{messages.projectSearchLabel}</Label>
                  <Input
                    id="backlog-project-search"
                    data-testid="backlog-project-search"
                    type="search"
                    value={filter}
                    onChange={(e: { target: { value: string } }) => setFilter(e.target.value)}
                  />
                </div>
                {visible.map((p) => (
                  <div key={p.projectKey} className={ROW}>
                    <input
                      type="checkbox"
                      id={`backlog-project-${p.projectKey}`}
                      data-testid={`backlog-project-${p.projectKey}`}
                      checked={checked.includes(p.projectKey)}
                      onChange={() => toggle(p.projectKey)}
                    />
                    <label htmlFor={`backlog-project-${p.projectKey}`}>
                      {t({ key: "projectLabel", params: { name: p.projectName, key: p.projectKey } })}
                    </label>
                  </div>
                ))}
                <Button type="button" data-testid="backlog-projects-save" disabled={saving} onClick={onSave}>
                  {saving ? messages.saving : messages.saveProjects}
                </Button>
              </div>
            )}
            {fieldError ? (
              <p id="backlog-projects-error" data-testid="backlog-projects-error">
                {messages.errorProjectKeys}
              </p>
            ) : null}
          </fieldset>
        ) : null}
        {message && !fieldError ? <p data-testid="backlog-projects-message">{t(message)}</p> : null}
        {confirming ? (
          <ConfirmDialog
            testId="backlog-deselect-dialog"
            title={messages.deselectTitle}
            body={withImpact(messages.deselectBody, impact)}
            confirmLabel={messages.confirm}
            onConfirm={save}
            onClose={() => setConfirming(false)}
          />
        ) : null}
      </section>
    );
  };
}
