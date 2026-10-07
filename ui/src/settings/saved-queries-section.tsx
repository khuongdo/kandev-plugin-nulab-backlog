import type { Component, PluginHostApi } from "@kandev/plugin-sdk";

import { createSaveQueryDialog, queryFilters, type Query } from "../git/save-query-dialog";
import { hostUi } from "../host-ui";
import { en, format, type Messages } from "../messages/en";
import { createConfirmDialog } from "./confirm-dialog";
import { createSectionParts } from "./section-parts";
import { useActionList } from "./use-list";

/**
 * Saved PR queries in Settings (WF6, Q3): a table with Edit (rename) and
 * Delete. New queries are saved from the Pull requests list only (BR2.5).
 */
export function createSavedQueriesSection(
  host: PluginHostApi,
  messages: Messages = en,
): Component<{ workspaceId: string }> {
  const h = host.jsx;
  const { useState } = host.React;
  const ui = hostUi(host);
  const {
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

  return function SavedQueriesSection({ workspaceId }: { workspaceId: string }) {
    const list = useActionList<Query>(host, workspaceId, "git.queries.list", "queries");
    const [editing, setEditing] = useState<Query | undefined>(undefined);
    const [deleting, setDeleting] = useState<Query | undefined>(undefined);

    const body = list.error ? (
      <ListError testId="backlog-saved-queries" notice={list.error} onRetry={list.reload} />
    ) : !list.loading && list.items.length === 0 ? (
      <ListEmpty testId="backlog-saved-queries" title={messages.savedQueriesEmpty} />
    ) : (
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{messages.colName}</TableHead>
                <TableHead>{messages.colFilters}</TableHead>
                <TableHead>{messages.colActions}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {list.items.map((q) => (
                <TableRow key={q.id} data-testid={`backlog-saved-query-row-${q.id}`}>
                  <TableCell>{q.name}</TableCell>
                  <TableCell>{queryFilters(q, messages)}</TableCell>
                  <TableCell>
                    <RowMenu
                      testId={`backlog-saved-query-menu-${q.id}`}
                      label={format(messages.rowActions, { name: q.name })}
                      items={[
                        {
                          testId: `backlog-saved-query-edit-${q.id}`,
                          label: messages.edit,
                          onSelect: () => setEditing(q),
                        },
                        {
                          testId: `backlog-saved-query-delete-${q.id}`,
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
      <div data-testid="backlog-section-saved-queries">
        <SettingsSection title={messages.sectionSavedQueries} description={messages.savedQueriesDescription}>
          {body}
        </SettingsSection>
        {editing ? (
          <SaveQueryDialog
            workspaceId={workspaceId}
            query={editing}
            onSaved={(q: Query) => {
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
              await host.api.invokeAction("git.queries.delete", { workspaceId, body: { id: deleting.id } });
              list.setItems((items) => items.filter((x) => x.id !== deleting.id));
            }}
            onClose={() => setDeleting(undefined)}
          />
        ) : null}
      </div>
    );
  };
}
