import * as React from "react";
import { afterEach, describe, expect, it } from "vitest";
import type { PluginHostApi } from "@kandev/plugin-sdk";

import { createBacklogIcon, PLUGIN_ICON } from "./backlog-logo";
import { mount, unmount } from "../testing/harness";

const host = { React, jsx: React.createElement } as unknown as PluginHostApi;
const ALLOWED_ATTRIBUTES = new Set([
  "viewBox",
  "fill",
  "stroke",
  "stroke-width",
  "stroke-linecap",
  "stroke-linejoin",
  "d",
  "x",
  "y",
  "width",
  "height",
  "rx",
  "aria-hidden",
  "focusable",
  "class",
  "xmlns",
]);

afterEach(unmount);

describe("Plugin icon (FR6, BR6.1)", () => {
  it("is an outline drawing in the current text colour", async () => {
    const c = await mount(createBacklogIcon(host), {});
    const svg = c.querySelector("svg")!;
    expect(svg.getAttribute("viewBox")).toBe("0 0 24 24");
    expect(svg.getAttribute("fill")).toBe("none");
    expect(svg.getAttribute("stroke")).toBe("currentColor");
    expect(svg.getAttribute("stroke-width")).toBe("2");
    expect(svg.getAttribute("stroke-linecap")).toBe("round");
    expect(svg.getAttribute("stroke-linejoin")).toBe("round");
    for (const el of c.querySelectorAll("svg *")) {
      expect(el.getAttribute("fill"), el.outerHTML).toBeNull();
    }
  });

  it("is not the Nulab mark: no brand colour and none of its paths", async () => {
    const c = await mount(createBacklogIcon(host), {});
    expect(c.innerHTML).not.toMatch(/#42CE9F|white/i);
    expect(c.innerHTML).not.toContain("M38.2091 48H9.79093");
    expect(c.innerHTML).not.toContain("M21.1588 20.4481");
  });

  it("renders only svg shapes, decorative, with fixed attributes", async () => {
    const c = await mount(createBacklogIcon(host), { className: "h-4 w-4" });
    const tags = new Set([...c.querySelectorAll("*")].map((el) => el.tagName.toLowerCase()));
    expect([...tags].every((t) => ["svg", "path", "rect"].includes(t))).toBe(true);
    const svg = c.querySelector("svg")!;
    expect(svg.getAttribute("aria-hidden")).toBe("true");
    expect(svg.getAttribute("focusable")).toBe("false");
    for (const el of c.querySelectorAll("*")) {
      for (const attr of el.getAttributeNames()) {
        expect(ALLOWED_ATTRIBUTES.has(attr), `unexpected attribute ${attr}`).toBe(true);
      }
    }
    expect(c.innerHTML).not.toMatch(/url\(|https?:\/\/(?!www\.w3\.org)|javascript:/i);
  });

  it("takes its size from className only", async () => {
    const c = await mount(createBacklogIcon(host), { className: "h-5 w-5", onClick: () => undefined });
    const svg = c.querySelector("svg")!;
    expect(svg.getAttribute("class")).toBe("h-5 w-5");
    expect(svg.getAttribute("width")).toBeNull();
    expect(svg.getAttribute("height")).toBeNull();
  });

  it("is the plugin icon through the one PLUGIN_ICON selection point", () => {
    expect(PLUGIN_ICON).toBe(createBacklogIcon);
  });
});
