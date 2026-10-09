import { afterEach, describe, expect, it } from "vitest";

import { createErrorAlert } from "./error-alert";
import { axeViolations, byTestId, fakeHost, mount, unmount } from "./testing/harness";

afterEach(unmount);

const LONG = "The Backlog issues could not be loaded. ".repeat(8).trim();

describe("Shared error alert (intent 261009, FR2.2-FR2.4)", () => {
  it("shows the message as a full-width, wrapping destructive alert", async () => {
    const host = fakeHost(async () => ({}));
    const c = await mount(createErrorAlert(host), { message: LONG, testId: "x-error" });
    const alert = byTestId(c, "x-error")!;
    expect(alert.getAttribute("data-host")).toBe("Alert");
    expect(alert.getAttribute("role")).toBe("alert");
    expect(alert.getAttribute("variant")).toBe("destructive");
    expect(alert.className.split(" ")).toEqual(
      expect.arrayContaining(["w-full", "min-w-0", "whitespace-normal", "break-words", "text-destructive"]),
    );
    expect(alert.textContent).toBe(LONG);
    expect(await axeViolations(c)).toEqual([]);
  });

  it("keeps a retry or reconnect control inside the alert", async () => {
    const host = fakeHost(async () => ({}));
    const action = host.jsx("button", { type: "button", "data-testid": "x-retry" }, "Retry");
    const c = await mount(createErrorAlert(host), { message: "Failed.", testId: "x-error", action });
    const alert = byTestId(c, "x-error")!;
    expect(alert.contains(byTestId(c, "x-retry"))).toBe(true);
    expect(byTestId(c, "x-error-message")!.textContent).toBe("Failed.");
  });

  it("shows an optional title above the message, under a chosen message test id", async () => {
    const host = fakeHost(async () => ({}));
    const c = await mount(createErrorAlert(host), {
      message: "Not connected.",
      title: "Backlog is not available",
      testId: "x-error",
      messageTestId: "x-status",
    });
    const title = byTestId(c, "x-error")!.querySelector('[data-host="AlertTitle"]')!;
    expect(title.textContent).toBe("Backlog is not available");
    expect(byTestId(c, "x-status")!.textContent).toBe("Not connected.");
  });

  it("renders nothing without a message", async () => {
    const host = fakeHost(async () => ({}));
    const c = await mount(createErrorAlert(host), { message: "", testId: "x-error" });
    expect(byTestId(c, "x-error")).toBeNull();
  });
});
