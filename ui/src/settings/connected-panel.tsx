import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { createGitAccess } from "../git/git-access";
import { loadImpact, withImpact } from "../git/git-state";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { failureNotice, type ConnectionView, type Notice } from "./state";

type AnyProps = Record<string, unknown>;

export interface ConnectedPanelProps {
  workspaceId: string;
  /** Called with the view after a re-check or a disconnect. */
  onView: (view: ConnectionView) => void;
  /** Sends a message to the screen's live region. */
  announce: (notice: Notice) => void;
  /** U4: the current view, for the Git access block (US5.5). */
  view?: ConnectionView;
}

const STACK = "flex flex-col gap-4";
const ROW = "flex gap-2";

/** Test connection and Disconnect for a connected workspace (US1.5). */
export function createConnectedPanel(
  host: PluginHostApi,
  messages: Messages = en,
): Component<ConnectedPanelProps> {
  const h = host.jsx;
  const { useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const ConfirmDialog = createConfirmDialog(host, messages);
  const GitAccess = createGitAccess(host, messages);
  const t = (n: Notice) => format(messages[n.key], n.params);

  return function ConnectedPanel({ workspaceId, onView, announce, view }: ConnectedPanelProps) {
    const [testing, setTesting] = useState(false);
    const [result, setResult] = useState<Notice | undefined>(undefined);
    const [confirming, setConfirming] = useState(false);
    const [gitCheck, setGitCheck] = useState<string | undefined>(undefined);
    const [impact, setImpact] = useState("");

    const test = async () => {
      if (testing) return;
      setTesting(true);
      setResult(undefined);
      let notice: Notice;
      try {
        const view = await host.api.invokeAction<ConnectionView>("connection.test", { workspaceId });
        notice = { key: "testSucceeded", params: { name: view.connectedUserName ?? "" } };
        setGitCheck(view.gitCheck);
        onView(view);
      } catch (error) {
        notice = failureNotice(error).notice ?? { key: "actionFailed" };
      }
      setTesting(false);
      setResult(notice);
      announce(notice);
    };

    const disconnect = async () => {
      const view = await host.api.invokeAction<ConnectionView>("connection.disconnect", { workspaceId });
      onView(view);
    };

    return (
      <div data-testid="backlog-connected-panel" className={STACK}>
        <div className={ROW}>
          <Button
            type="button"
            data-testid="backlog-test-connection"
            disabled={testing}
            onClick={() => void test()}
          >
            {testing ? messages.testing : messages.testConnection}
          </Button>
          <Button
            type="button"
            data-testid="backlog-disconnect"
            onClick={() => {
              setImpact("");
              setConfirming(true);
              void loadImpact(host, workspaceId, undefined, messages).then(setImpact);
            }}
          >
            {messages.disconnect}
          </Button>
        </div>
        {result ? <p data-testid="backlog-test-result">{t(result)}</p> : null}
        <GitAccess
          workspaceId={workspaceId}
          hasGitCredential={Boolean(view?.hasGitCredential)}
          gitCheck={gitCheck}
          announce={announce}
        />
        {confirming ? (
          <ConfirmDialog
            testId="backlog-disconnect-dialog"
            title={messages.disconnectTitle}
            body={withImpact(messages.disconnectBody, impact)}
            confirmLabel={messages.disconnect}
            onConfirm={disconnect}
            onClose={() => setConfirming(false)}
          />
        ) : null}
      </div>
    );
  };
}
