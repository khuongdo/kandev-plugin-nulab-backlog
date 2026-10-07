import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { withImpact } from "../git/git-state";
import { hostUi } from "../host-ui";
import { loadImpactText } from "../issues/issues-state";
import { BUTTON, ROW, STACK } from "../layout";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { failureNotice, type ConnectionView, type Notice } from "./state";

export interface ConnectedPanelProps {
  workspaceId: string;
  /** Called with the view after a re-check or a disconnect. */
  onView: (view: ConnectionView) => void;
  /** Sends a message to the screen's live region. */
  announce: (notice: Notice) => void;
  /** U4: the Git check of the last Test connection, for the Git access section (AC5.5.2). */
  onGitCheck?: (gitCheck: string | undefined) => void;
}

/** Test connection and Disconnect for a connected workspace (US1.5). */
export function createConnectedPanel(
  host: PluginHostApi,
  messages: Messages = en,
): Component<ConnectedPanelProps> {
  const h = host.jsx;
  const { useState } = host.React;
  const { Button } = hostUi(host);
  const ConfirmDialog = createConfirmDialog(host, messages);
  const t = (n: Notice) => format(messages[n.key], n.params);

  return function ConnectedPanel({ workspaceId, onView, announce, onGitCheck }: ConnectedPanelProps) {
    const [testing, setTesting] = useState(false);
    const [result, setResult] = useState<Notice | undefined>(undefined);
    const [confirming, setConfirming] = useState(false);
    const [impact, setImpact] = useState("");

    const test = async () => {
      if (testing) return;
      setTesting(true);
      setResult(undefined);
      let notice: Notice;
      try {
        const view = await host.api.invokeAction<ConnectionView>("connection.test", { workspaceId });
        notice = { key: "testSucceeded", params: { name: view.connectedUserName ?? "" } };
        onGitCheck?.(view.gitCheck);
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
            variant="outline"
            className={BUTTON}
            data-testid="backlog-test-connection"
            disabled={testing}
            onClick={() => void test()}
          >
            {testing ? messages.testing : messages.testConnection}
          </Button>
          <Button
            type="button"
            variant="destructive"
            className={BUTTON}
            data-testid="backlog-disconnect"
            onClick={() => {
              setImpact("");
              setConfirming(true);
              void loadImpactText(host, workspaceId, undefined, messages).then(setImpact);
            }}
          >
            {messages.disconnect}
          </Button>
        </div>
        {result ? <p data-testid="backlog-test-result">{t(result)}</p> : null}
        {confirming ? (
          <ConfirmDialog
            testId="backlog-disconnect-dialog"
            title={messages.disconnectTitle}
            body={withImpact(messages.disconnectBody, impact)}
            confirmLabel={messages.disconnect}
            destructive
            onConfirm={disconnect}
            onClose={() => setConfirming(false)}
          />
        ) : null}
      </div>
    );
  };
}
