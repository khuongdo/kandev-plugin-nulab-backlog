import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { noticeText } from "../git/git-state";
import { en, type Messages } from "../messages/en";
import { readFailure, type Notice } from "../settings/state";
import { issueNotice } from "./issues-state";

type AnyProps = Record<string, unknown>;

const FIELD = "flex flex-col gap-2";

/** A whole number of minutes from 1 to 1440, or undefined (AC4.2.2). */
function parseMinutes(raw: string): number | undefined {
  if (!/^\d+$/.test(raw.trim())) return undefined;
  const n = Number(raw);
  return n >= 1 && n <= 1440 ? n : undefined;
}

/** The sync interval field of the connected settings (M1, US4.2). */
export function createPollInterval(
  host: PluginHostApi,
  messages: Messages = en,
): Component<{ workspaceId: string }> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const Input = host.ui.Input as Component<AnyProps>;
  const Label = host.ui.Label as Component<AnyProps>;

  return function PollInterval({ workspaceId }: { workspaceId: string }) {
    const [value, setValue] = useState("5");
    const [invalid, setInvalid] = useState(false);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<Notice | undefined>(undefined);

    useEffect(() => {
      host.api
        .invokeAction<{ pollMinutes?: number }>("issues.settings.get", { workspaceId })
        .then((s) => s?.pollMinutes && setValue(String(s.pollMinutes)))
        .catch(() => undefined); // the default stays; saving still works
    }, [workspaceId]);

    const save = async () => {
      setMessage(undefined);
      const minutes = parseMinutes(value);
      setInvalid(minutes === undefined);
      if (minutes === undefined) return;
      setSaving(true);
      try {
        await host.api.invokeAction("issues.set_poll_interval", { workspaceId, body: { minutes } });
        setMessage({ key: "pollSaved" });
      } catch (error) {
        const f = readFailure(error);
        if (f.code === "validation" && f.field === "minutes") setInvalid(true);
        else setMessage(issueNotice(error));
      }
      setSaving(false);
    };

    return (
      <div data-testid="backlog-poll-interval" className={FIELD}>
        <Label htmlFor="backlog-poll-minutes">{messages.pollLabel}</Label>
        <Input
          id="backlog-poll-minutes"
          data-testid="backlog-poll-minutes"
          type="number"
          min={1}
          max={1440}
          step={1}
          inputMode="numeric"
          value={value}
          aria-invalid={invalid}
          aria-describedby={invalid ? "backlog-poll-help backlog-poll-error" : "backlog-poll-help"}
          onChange={(e: { target: { value: string } }) => setValue(e.target.value)}
        />
        <p id="backlog-poll-help">{messages.pollHelp}</p>
        {invalid ? (
          <p id="backlog-poll-error" role="alert" data-testid="backlog-poll-error">
            {messages.pollMinError}
          </p>
        ) : null}
        <div>
          <Button type="button" data-testid="backlog-poll-save" disabled={saving} onClick={() => void save()}>
            {saving ? messages.saving : messages.pollSave}
          </Button>
        </div>
        <p role="status" data-testid="backlog-poll-message">
          {message ? noticeText(message, messages) : ""}
        </p>
      </div>
    );
  };
}
