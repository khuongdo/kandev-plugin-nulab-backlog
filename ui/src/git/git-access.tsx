import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { en, type MessageKey, type Messages } from "../messages/en";
import { readFailure, type ConnectionView, type Notice } from "../settings/state";
import { gitNotice, noticeText } from "./git-state";

type AnyProps = Record<string, unknown>;

export interface GitAccessProps {
  workspaceId: string;
  hasGitCredential: boolean;
  /** The gitCheck of the last Test connection (AC5.5.2). */
  gitCheck?: string;
  announce: (notice: Notice) => void;
}

type GitField = "gitUsername" | "gitPassword";

const IDS: Record<GitField, string> = {
  gitUsername: "backlog-git-username",
  gitPassword: "backlog-git-password",
};
const ERRORS: Record<GitField, MessageKey> = {
  gitUsername: "errorGitUsername",
  gitPassword: "errorGitPassword",
};

const STACK = "flex flex-col gap-4";
const FIELD = "flex flex-col gap-2";

/** The optional Git access block of the connected settings panel (M1, US5.5). */
export function createGitAccess(host: PluginHostApi, messages: Messages = en): Component<GitAccessProps> {
  const h = host.jsx;
  const { useEffect, useState } = host.React;
  const Button = host.ui.Button as Component<AnyProps>;
  const Input = host.ui.Input as Component<AnyProps>;
  const Label = host.ui.Label as Component<AnyProps>;

  return function GitAccess({ workspaceId, hasGitCredential, gitCheck, announce }: GitAccessProps) {
    const [username, setUsername] = useState("");
    const [secret, setSecret] = useState("");
    const [stored, setStored] = useState(hasGitCredential);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<Notice | undefined>(undefined);
    const [fieldError, setFieldError] = useState<GitField | undefined>(undefined);

    useEffect(() => setStored(hasGitCredential), [hasGitCredential]);
    useEffect(() => {
      if (gitCheck === "invalid") announce({ key: "gitInvalid" });
    }, [gitCheck]);

    const save = async () => {
      if (saving) return;
      setSaving(true);
      setMessage(undefined);
      setFieldError(undefined);
      let notice: Notice;
      try {
        const view = await host.api.invokeAction<ConnectionView>("connection.set_git_credential", {
          workspaceId,
          body: { gitUsername: username, gitPassword: secret },
        });
        setStored(view?.hasGitCredential !== false);
        notice = { key: "gitSaved" };
      } catch (error) {
        const f = readFailure(error);
        const field =
          f.code === "validation" && (f.field === "gitUsername" || f.field === "gitPassword")
            ? f.field
            : undefined;
        setFieldError(field);
        notice = field ? { key: ERRORS[field] } : gitNotice(error);
      }
      setSecret(""); // the password never stays in the page
      setSaving(false);
      setMessage(notice);
      announce(notice);
    };

    const field = (
      name: GitField,
      label: string,
      value: string,
      set: (v: string) => void,
      extra: AnyProps,
    ) => {
      const id = IDS[name];
      const invalid = fieldError === name;
      return (
        <div className={FIELD}>
          <Label htmlFor={id}>{label}</Label>
          <Input
            id={id}
            data-testid={id}
            value={value}
            onChange={(e: { target: { value: string } }) => set(e.target.value)}
            aria-invalid={invalid ? "true" : undefined}
            aria-describedby={invalid ? `${id}-error` : undefined}
            {...extra}
          />
          {invalid ? (
            <p id={`${id}-error`} data-testid={`${id}-error`}>
              {messages[ERRORS[name]]}
            </p>
          ) : null}
        </div>
      );
    };

    return (
      <details data-testid="backlog-git-access" className={STACK}>
        <summary data-testid="backlog-git-toggle">{messages.gitAccessHeading}</summary>
        <div className={STACK}>
          <p>{messages.gitAccessHelp}</p>
          <p data-testid="backlog-git-status">{stored ? messages.gitStored : messages.gitNotStored}</p>
          {gitCheck === "invalid" ? (
            <p role="alert" data-testid="backlog-git-invalid">
              {messages.gitInvalid}
            </p>
          ) : null}
          {field("gitUsername", messages.gitUsernameLabel, username, setUsername, {
            autoComplete: "username",
          })}
          {field("gitPassword", messages.gitPasswordLabel, secret, setSecret, {
            type: "password",
            autoComplete: "new-password",
          })}
          <Button type="button" data-testid="backlog-git-save" disabled={saving} onClick={() => void save()}>
            {saving ? messages.saving : messages.gitSave}
          </Button>
          {message && !fieldError ? (
            <p data-testid="backlog-git-message">{noticeText(message, messages)}</p>
          ) : null}
        </div>
      </details>
    );
  };
}
