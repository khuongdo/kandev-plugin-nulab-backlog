import * as React from "react";
import { afterEach, describe, expect, it } from "vitest";
import type { PluginHostApi } from "@kandev/plugin-sdk";

import { createBacklogLogo, PLUGIN_ICON } from "./backlog-logo";
import { mount, unmount } from "../testing/harness";

const host = { React, jsx: React.createElement } as unknown as PluginHostApi;
const ALLOWED_ATTRIBUTES = new Set([
  "viewBox",
  "fill",
  "fill-rule",
  "clip-rule",
  "d",
  "aria-hidden",
  "focusable",
  "class",
  "xmlns",
]);

afterEach(unmount);

describe("Backlog logo (NFR3.10, BR7.8)", () => {
  it("renders only svg and path elements", async () => {
    const c = await mount(createBacklogLogo(host), {});
    const tags = [...c.querySelectorAll("*")].map((el) => el.tagName.toLowerCase());
    expect(tags[0]).toBe("svg");
    expect(new Set(tags)).toEqual(new Set(["svg", "path"]));
  });

  it("is decorative: aria-hidden and not focusable", async () => {
    const c = await mount(createBacklogLogo(host), {});
    const svg = c.querySelector("svg")!;
    expect(svg.getAttribute("aria-hidden")).toBe("true");
    expect(svg.getAttribute("focusable")).toBe("false");
  });

  it("has no script, foreignObject, href or event handler, and only fixed attributes", async () => {
    const c = await mount(createBacklogLogo(host), { className: "size-4" });
    expect(c.querySelector("script, foreignObject, use, image, a, style")).toBeNull();
    for (const el of c.querySelectorAll("*")) {
      for (const attr of el.getAttributeNames()) {
        expect(attr.startsWith("on"), `${attr} on ${el.tagName}`).toBe(false);
        expect(attr.includes("href"), `${attr} on ${el.tagName}`).toBe(false);
        expect(ALLOWED_ATTRIBUTES.has(attr), `unexpected attribute ${attr}`).toBe(true);
      }
    }
    expect(c.innerHTML).not.toMatch(/url\(|https?:|javascript:/i);
  });

  it("keeps the official colours and passes only the host's sizing class through", async () => {
    const c = await mount(createBacklogLogo(host), { className: "size-4", onClick: () => undefined });
    const svg = c.querySelector("svg")!;
    expect(svg.getAttribute("class")).toBe("size-4");
    expect([...c.querySelectorAll("path")].map((p) => p.getAttribute("fill"))).toEqual(["#42CE9F", "white"]);
  });

  it("is the plugin icon through the one PLUGIN_ICON constant", () => {
    expect(PLUGIN_ICON).toBe(createBacklogLogo);
  });
});
