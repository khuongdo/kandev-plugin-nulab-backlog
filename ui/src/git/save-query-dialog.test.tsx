import { act } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { en } from "../messages/en";
import { actionError, byTestId, expectErrorAlert, fakeHost, mount, unmount } from "../testing/harness";
import { createSaveQueryDialog, type Query } from "./save-query-dialog";

afterEach(unmount);

const QUERY: Query = {
  name: "Mine",
  projectKey: "PROJ",
  repoName: "web-app",
  statuses: ["open"],
  assignee: "me",
  creator: "anyone",
};

describe("Save query dialog errors (intent 261009, FR2.3)", () => {
  it("shows a failed save as the shared alert above the dialog actions", async () => {
    const host = fakeHost(async () => {
      throw actionError(503, { code: "unreachable" });
    });
    const onSaved = vi.fn();
    const c = await mount(createSaveQueryDialog(host, en), {
      workspaceId: "ws-1",
      query: QUERY,
      onSaved,
      onClose: vi.fn(),
    });
    await act(async () => byTestId(c, "backlog-save-query-save")!.click());
    const alert = expectErrorAlert(byTestId(c, "backlog-save-query-notice"));
    expect(alert.textContent).toBe(en.unreachable);
    const cancel = byTestId(c, "backlog-save-query-cancel")!;
    expect(alert.compareDocumentPosition(cancel) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(onSaved).not.toHaveBeenCalled();
  });
});
