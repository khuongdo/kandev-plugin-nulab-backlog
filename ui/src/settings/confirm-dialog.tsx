import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { hostUi } from "../host-ui";
import { BUTTON } from "../layout";
import { en, type Messages } from "../messages/en";

export interface ConfirmProps {
  /** Base data-testid; the buttons get `-cancel` and `-confirm`. */
  testId: string;
  title: string;
  body: string;
  confirmLabel: string;
  /** Runs on Confirm; a rejection keeps the dialog open with an inline error. */
  onConfirm: () => Promise<void>;
  onClose: () => void;
  /** Destructive confirms (delete, disconnect) use the destructive button (BR5.2). */
  destructive?: boolean;
  /** U3: extra content under the body, such as a link (the three-choice dialog). */
  children?: unknown;
}

/**
 * The shared confirmation dialog on the host's Dialog (SDK v0.96.0 has no
 * AlertDialog). The host traps the focus and closes on Esc; the plugin puts
 * the focus on Cancel and gives it back to the element that opened it.
 */
export function createConfirmDialog(host: PluginHostApi, messages: Messages = en): Component<ConfirmProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const { Button, Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } =
    hostUi(host);

  return function ConfirmDialog({
    testId,
    title,
    body,
    confirmLabel,
    onConfirm,
    onClose,
    destructive,
    children,
  }: ConfirmProps) {
    const [working, setWorking] = useState(false);
    const [failed, setFailed] = useState(false);

    useEffect(() => {
      const opener = document.activeElement as HTMLElement | null;
      document.querySelector<HTMLElement>(`[data-testid="${testId}-cancel"]`)?.focus();
      return () => opener?.focus();
    }, []);

    const confirm = async () => {
      if (working) return;
      setWorking(true);
      setFailed(false);
      try {
        await onConfirm();
      } catch {
        setWorking(false);
        setFailed(true);
        return;
      }
      onClose();
    };

    return (
      <Dialog open onOpenChange={(open: boolean) => !open && !working && onClose()}>
        <DialogContent
          data-testid={testId}
          aria-labelledby={`${testId}-title`}
          aria-describedby={`${testId}-body`}
        >
          <DialogHeader>
            <DialogTitle id={`${testId}-title`}>{title}</DialogTitle>
            <DialogDescription id={`${testId}-body`}>{body}</DialogDescription>
          </DialogHeader>
          {children ?? null}
          {failed ? (
            <p role="alert" data-testid={`${testId}-error`}>
              {messages.actionFailed}
            </p>
          ) : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              className={BUTTON}
              data-testid={`${testId}-cancel`}
              disabled={working}
              onClick={onClose}
            >
              {messages.cancel}
            </Button>
            <Button
              type="button"
              variant={destructive ? "destructive" : "default"}
              className={BUTTON}
              data-testid={`${testId}-confirm`}
              disabled={working}
              onClick={() => void confirm()}
            >
              {working ? messages.working : confirmLabel}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  };
}
