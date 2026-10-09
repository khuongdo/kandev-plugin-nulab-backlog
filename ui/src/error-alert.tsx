import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "./host-ui";
import { ROW } from "./layout";

export interface ErrorAlertProps {
  /** The catalogue text; nothing renders when it is empty. */
  message: string;
  /** An optional bold first line, such as "Backlog is not available". */
  title?: string;
  /** The alert's test id. */
  testId: string;
  /** The message's test id; `<testId>-message` by default. */
  messageTestId?: string;
  /** An optional retry or reconnect control, kept inside the alert. */
  action?: unknown;
}

/** Full width, wraps long text, never forces horizontal scroll (NFR1). */
export const ALERT = "w-full min-w-0 whitespace-normal break-words text-destructive";

/**
 * The one inline error of the Backlog page (intent 261009, FR2.2-FR2.4): the
 * host's destructive Alert for page-state errors at the top of a section and
 * for dialog errors above the dialog actions. Click errors use host.toast.
 */
export function createErrorAlert(host: PluginHostApi): Component<ErrorAlertProps> {
  const h = host.jsx;
  const { Alert, AlertDescription, AlertTitle } = hostUi(host);

  return function ErrorAlert({ message, title, testId, messageTestId, action }: ErrorAlertProps) {
    if (!message) return null;
    return (
      <Alert variant="destructive" role="alert" className={ALERT} data-testid={testId}>
        {title ? <AlertTitle>{title}</AlertTitle> : null}
        <AlertDescription data-testid={messageTestId ?? `${testId}-message`}>{message}</AlertDescription>
        {action ? <div className={ROW}>{action}</div> : null}
      </Alert>
    );
  };
}
