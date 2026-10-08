import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import {
  createSaveQueryDialog,
  queryFilters,
  scmQueryFilters,
  type Query,
  type ScmQuery,
} from "../git/save-query-dialog";
import { hostUi } from "../host-ui";
import { issueQueryFilters, type IssueQuery } from "../issues/issues-state";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { createSectionParts } from "./section-parts";
import { useActionList } from "./use-list";

type Saved = (Query | IssueQuery | ScmQuery) & { id?: string; isDefault?: boolean };

interface Kind {
  /** issues.queries or git.queries. */
  action: string;
  testId: string;
  title: string;
  empty: string;
  filters: (q: Saved) => string;
}

/**
 * Saved queries in Settings (WF6, Q3; FR3.3, FR4.3, FR4.4): one table for
 * issue queries and one for PR queries, each row with Rename, Make default
 * (at most one per kind) and Delete. New queries are saved from the lists.
 */
export function createSavedQueriesSection(
  host: PluginHostApi,
  messages: Messages = en,
): Component<{ workspaceId: string }> {
  const h = host.jsx;
  const { useState } = host.React;
  const ui = hostUi(host);
  const {
    Badge,
    Card,
    CardContent,
    SettingsSection,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } = ui;
  const { ListEmpty, ListError, RowMenu } = createSectionParts(host, messages);
  const ConfirmDialog = createConfirmDialog(host, messages);
  const SaveQueryDialog = createSaveQueryDialog(host, messages);
  const issueKind: Kind = {
    action: "issues.queries",
    testId: "backlog-saved-issue-quer",
    title: messages.scopeIssues,
    empty: messages.savedIssueQueriesEmpty,
    filters: (q) => issueQueryFilters(q as IssueQuery, messages),
  };
  const prKind: Kind = {
    action: "git.queries",
    testId: "backlog-saved-quer",
    title: messages.scopePRs,
    empty: messages.savedQueriesEmpty,
    filters: (q) => queryFilters(q as Query, messages),
  };

  // Intent 261007-source-control-agnostic (FR4.2, FR3.4): one default per provider.
  const scmKind: Kind = {
    action: "scm.queries",
    testId: "backlog-saved-scm-quer",
    title: messages.scmSavedQueriesTitle,
    empty: messages.scmSavedQueriesEmpty,
    filters: (q) => {
      const text = scmQueryFilters(q as ScmQuery, messages);
      return (q as ScmQuery).unmapped ? `${text} · ${messages.scmUnmapped}` : text;
    },
  };

  function QueryTable({ workspaceId, kind }: { workspaceId: string; kind: Kind }) {
    const list = useActionList<Saved>(host, workspaceId, `${kind.action}.list`, "queries");
    const [editing, setEditing] = useState<Saved | undefined>(undefined);
    const [deleting, setDeleting] = useState<Saved | undefined>(undefined);
    const many = `${kind.testId}ies`;
    const one = `${kind.testId}y`;

    const star = async (q: Saved) => {
      const r = await host.api.invokeAction<{ queries?: Saved[] }>(`${kind.action}.set_default`, {
        workspaceId,
        body: { id: q.id, isDefault: !q.isDefault },
      });
      list.setItems((items) => r?.queries ?? items);
    };

    const body = list.error ? (
      <ListError testId={many} notice={list.error} onRetry={list.reload} />
    ) : !list.loading && list.items.length === 0 ? (
      <ListEmpty testId={many} title={kind.empty} />
    ) : (
      <Card>
        <CardContent className="p-0">
          <Table aria-label={kind.title}>
            <TableHeader>
              <TableRow>
                <TableHead>{messages.colName}</TableHead>
                <TableHead>{messages.colFilters}</TableHead>
                <TableHead>{messages.colActions}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {list.items.map((q) => (
                <TableRow key={q.id} data-testid={`${one}-row-${q.id}`}>
                  <TableCell>
                    {q.name} {q.isDefault ? <Badge variant="secondary">{messages.defaultQuery}</Badge> : null}
                  </TableCell>
                  <TableCell>{kind.filters(q)}</TableCell>
                  <TableCell>
                    <RowMenu
                      testId={`${one}-menu-${q.id}`}
                      label={format(messages.rowActions, { name: q.name })}
                      items={[
                        {
                          testId: `${one}-edit-${q.id}`,
                          label: messages.edit,
                          onSelect: () => setEditing(q),
                        },
                        {
                          testId: `${one}-star-${q.id}`,
                          label: q.isDefault ? messages.removeDefault : messages.makeDefault,
                          onSelect: () => void star(q).catch(list.reload),
                        },
                        {
                          testId: `${one}-delete-${q.id}`,
                          label: messages.delete,
                          onSelect: () => setDeleting(q),
                        },
                      ]}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );

    return (
      <div className="flex flex-col gap-2">
        <h4 className="text-sm font-medium">{kind.title}</h4>
        {body}
        {editing ? (
          <SaveQueryDialog
            workspaceId={workspaceId}
            query={editing}
            action={`${kind.action}.save`}
            description={kind.filters(editing)}
            onSaved={(q: Saved) => {
              list.setItems((items) => items.map((x) => (x.id === q.id ? q : x)));
              setEditing(undefined);
            }}
            onClose={() => setEditing(undefined)}
          />
        ) : null}
        {deleting ? (
          <ConfirmDialog
            testId="backlog-saved-query-delete-dialog"
            title={messages.deleteQueryTitle}
            body={messages.deleteQueryBody}
            confirmLabel={messages.delete}
            destructive
            onConfirm={async () => {
              await host.api.invokeAction(`${kind.action}.delete`, {
                workspaceId,
                body: { id: deleting.id },
              });
              list.setItems((items) => items.filter((x) => x.id !== deleting.id));
            }}
            onClose={() => setDeleting(undefined)}
          />
        ) : null}
      </div>
    );
  }

  return function SavedQueriesSection({ workspaceId }: { workspaceId: string }) {
    return (
      <div data-testid="backlog-section-saved-queries">
        <SettingsSection title={messages.sectionSavedQueries} description={messages.savedQueriesDescription}>
          <div className="flex flex-col gap-4">
            <QueryTable workspaceId={workspaceId} kind={issueKind} />
            <QueryTable workspaceId={workspaceId} kind={prKind} />
            <QueryTable workspaceId={workspaceId} kind={scmKind} />
          </div>
        </SettingsSection>
      </div>
    );
  };
}
