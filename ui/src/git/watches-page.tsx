import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { en, type Messages } from "../messages/en";
import { createConfirmDialog } from "../settings/confirm-dialog";
import type { ConnectionView, Notice } from "../settings/state";
import { gitNotice, noticeText, watchProgress, watchStatusKey } from "./git-state";
import { createWatchForm, type Watch } from "./watch-form";

type AnyProps = Record<string, unknown>;

const STACK = "flex flex-col gap-4";
const ROW = "flex gap-2";

type Load = { kind: "loading" } | { kind: "failed"; notice: Notice } | { kind: "ready" };

/** The PR watches page at /backlog/watches (M4, US6.1, US6.2). */
export function createWatchesPage(host: PluginHostApi, messages: Messages = en): Component {
  const h = host.jsx;
  const { useCallback, useEffect, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const ConfirmDialog = createConfirmDialog(host, messages);
  const WatchForm = createWatchForm(host, messages);
  const t = (n: Notice) => noticeText(n, messages);

  return function WatchesPage() {
    const [workspaceId, setWorkspaceId] = useState(host.context.getActiveWorkspaceId());
    const [load, setLoad] = useState<Load>({ kind: "loading" });
    const [watches, setWatches] = useState<Watch[]>([]);
    const [connected, setConnected] = useState(true);
    const [editing, setEditing] = useState<Partial<Watch> | undefined>(undefined);
    const [running, setRunning] = useState<string | undefined>(undefined);
    const [deleting, setDeleting] = useState<string | undefined>(undefined);
    const [notice, setNotice] = useState<Notice | undefined>(undefined);

    useEffect(() => host.context.subscribeActiveWorkspace(setWorkspaceId), []);

    const reload = useCallback(async () => {
      if (!workspaceId) return;
      setLoad({ kind: "loading" });
      try {
        const [view, list] = await Promise.all([
          host.api.invokeAction<ConnectionView>("connection.get", { workspaceId }),
          host.api.invokeAction<{ watches?: Watch[] }>("git.watches.list", { workspaceId }),
        ]);
        setConnected(view?.state === "connected");
        setWatches(list?.watches ?? []);
        setLoad({ kind: "ready" });
      } catch (error) {
        setLoad({ kind: "failed", notice: gitNotice(error) });
      }
    }, [workspaceId]);

    useEffect(() => {
      void reload();
    }, [reload]);

    const replace = (w: Watch) => setWatches((list) => list.map((x) => (x.id === w.id ? w : x)));

    const act = async (key: string, id: string) => {
      setNotice(undefined);
      try {
        replace(await host.api.invokeAction<Watch>(key, { workspaceId, body: { id } }));
      } catch (error) {
        setNotice(gitNotice(error));
      }
    };

    const run = async (id: string) => {
      setRunning(id);
      setNotice(undefined);
      try {
        await host.api.invokeAction("git.watches.run", { workspaceId, body: { id } });
        setNotice({ key: "runQueued" });
      } catch (error) {
        setNotice(gitNotice(error));
      }
      setRunning(undefined);
    };

    const remove = async (id: string) => {
      await host.api.invokeAction("git.watches.delete", { workspaceId, body: { id } });
      setWatches((list) => list.filter((w) => w.id !== id));
    };

    const onSaved = (w: Watch) => {
      setWatches((list) =>
        list.some((x) => x.id === w.id) ? list.map((x) => (x.id === w.id ? w : x)) : [...list, w],
      );
      setEditing(undefined);
    };

    if (load.kind === "loading") return <p data-testid="backlog-watches-loading">{messages.pageLoading}</p>;
    if (load.kind === "failed") {
      return (
        <div data-testid="backlog-watches-error" className={STACK}>
          <p>{messages.watchesLoadFailed}</p>
          <p>{t(load.notice)}</p>
          <Button type="button" data-testid="backlog-watches-retry" onClick={() => void reload()}>
            {messages.retry}
          </Button>
        </div>
      );
    }

    return (
      <div data-testid="backlog-watches" className={STACK}>
        {!connected ? <p data-testid="backlog-watches-not-connected">{messages.pageNotConnected}</p> : null}
        <div className={ROW}>
          <Button type="button" data-testid="backlog-watch-new" onClick={() => setEditing({})}>
            {messages.newWatch}
          </Button>
        </div>
        {editing && workspaceId ? (
          <WatchForm
            workspaceId={workspaceId}
            watch={editing}
            onSaved={onSaved}
            onCancel={() => setEditing(undefined)}
          />
        ) : null}
        {watches.length === 0 ? (
          <div data-testid="backlog-watches-empty">
            <p>{messages.watchesEmpty}</p>
          </div>
        ) : (
          <ul className={STACK}>
            {watches.map((w) => {
              const progress = watchProgress(w);
              return (
                <li key={w.id} data-testid={`backlog-watch-row-${w.id}`} className={STACK}>
                  <div className={ROW}>
                    <span>{w.name}</span>
                    <span data-testid={`backlog-watch-status-${w.id}`}>
                      {messages[watchStatusKey(w.state)]}
                    </span>
                  </div>
                  {progress ? <p data-testid={`backlog-watch-progress-${w.id}`}>{t(progress)}</p> : null}
                  <div className={ROW}>
                    {w.state === "active" ? (
                      <Button
                        type="button"
                        data-testid={`backlog-watch-run-${w.id}`}
                        disabled={running === w.id}
                        onClick={() => void run(w.id)}
                      >
                        {running === w.id ? messages.running : messages.run}
                      </Button>
                    ) : null}
                    {w.state === "active" ? (
                      <Button
                        type="button"
                        data-testid={`backlog-watch-pause-${w.id}`}
                        onClick={() => void act("git.watches.pause", w.id)}
                      >
                        {messages.pause}
                      </Button>
                    ) : null}
                    {w.state === "paused" ? (
                      <Button
                        type="button"
                        data-testid={`backlog-watch-resume-${w.id}`}
                        onClick={() => void act("git.watches.resume", w.id)}
                      >
                        {messages.resume}
                      </Button>
                    ) : null}
                    <Button
                      type="button"
                      data-testid={`backlog-watch-edit-${w.id}`}
                      onClick={() => setEditing(w)}
                    >
                      {messages.edit}
                    </Button>
                    <Button
                      type="button"
                      data-testid={`backlog-watch-delete-${w.id}`}
                      onClick={() => setDeleting(w.id)}
                    >
                      {messages.delete}
                    </Button>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
        {deleting ? (
          <ConfirmDialog
            testId="backlog-watch-delete-dialog"
            title={messages.deleteWatchTitle}
            body={messages.deleteWatchBody}
            confirmLabel={messages.delete}
            onConfirm={() => remove(deleting)}
            onClose={() => setDeleting(undefined)}
          />
        ) : null}
        <div role="status" aria-live="polite" data-testid="backlog-watches-notice">
          {notice ? t(notice) : ""}
        </div>
      </div>
    );
  };
}
