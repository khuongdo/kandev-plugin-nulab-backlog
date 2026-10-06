import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { en, type Messages } from "../messages/en";

type AnyProps = Record<string, unknown>;

export interface ConfirmProps {
  /** Base data-testid; the buttons get `-cancel` and `-confirm`. */
  testId: string;
  title: string;
  body: string;
  confirmLabel: string;
  /** Runs on Confirm; a rejection keeps the dialog open with an inline error. */
  onConfirm: () => Promise<void>;
  onClose: () => void;
  /** U3: extra content under the body, such as a link (the three-choice dialog). */
  children?: unknown;
}

const STACK = "flex flex-col gap-4";
const ROW = "flex gap-2";

/**
 * The shared confirmation dialog (interaction-spec ConfirmDialog). Kandev's
 * UI kit has no AlertDialog, so the plugin sets role="alertdialog", moves the
 * focus to Cancel, traps Tab, closes on Esc and gives the focus back to the
 * element that opened it.
 */
export function createConfirmDialog(host: PluginHostApi, messages: Messages = en): Component<ConfirmProps> {
  const h = host.jsx;
  const { useEffect, useRef, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;

  return function ConfirmDialog({
    testId,
    title,
    body,
    confirmLabel,
    onConfirm,
    onClose,
    children,
  }: ConfirmProps) {
    const [working, setWorking] = useState(false);
    const [failed, setFailed] = useState(false);
    const root = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
      const opener = document.activeElement as HTMLElement | null;
      root.current?.querySelector<HTMLElement>(`[data-testid="${testId}-cancel"]`)?.focus();
      return () => opener?.focus();
    }, []);

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        if (!working) onClose();
        return;
      }
      if (e.key !== "Tab" || !root.current) return;
      const items = [...root.current.querySelectorAll<HTMLElement>("a[href], button:not([disabled])")];
      if (items.length === 0) return;
      const first = items[0];
      const last = items[items.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };

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
      <div
        ref={root}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={`${testId}-title`}
        aria-describedby={`${testId}-body`}
        data-testid={testId}
        className={STACK}
        onKeyDown={onKeyDown}
      >
        <h3 id={`${testId}-title`}>{title}</h3>
        <p id={`${testId}-body`}>{body}</p>
        {children ?? null}
        {failed ? (
          <p role="alert" data-testid={`${testId}-error`}>
            {messages.actionFailed}
          </p>
        ) : null}
        <div className={ROW}>
          <Button type="button" data-testid={`${testId}-cancel`} disabled={working} onClick={onClose}>
            {messages.cancel}
          </Button>
          <Button
            type="button"
            data-testid={`${testId}-confirm`}
            disabled={working}
            onClick={() => void confirm()}
          >
            {working ? messages.working : confirmLabel}
          </Button>
        </div>
      </div>
    );
  };
}
