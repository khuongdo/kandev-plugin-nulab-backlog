import { describe, expect, it } from "vitest";

// Vite's raw import of every plugin source file (typed here: the UI config has no vite/client types).
declare global {
  interface ImportMeta {
    glob<T>(patterns: string[], options: { query: string; import: string; eager: true }): Record<string, T>;
  }
}

const sources = import.meta.glob<string>(["./**/*.tsx", "!./**/*.test.tsx", "!./testing/**"], {
  query: "?raw",
  import: "default",
  eager: true,
});

const RAW_CONTROLS = [/<select\b/, /<table\b/, /<details\b/, /<button\b/, /type="checkbox"/, /type="radio"/];

describe("Host controls only (FR5.1, BR5.1)", () => {
  it("finds the plugin components", () => {
    expect(Object.keys(sources).length).toBeGreaterThan(10);
  });

  it.each(Object.entries(sources))(
    "%s draws no raw select, table, details, button, checkbox or radio",
    (_, src) => {
      for (const pattern of RAW_CONTROLS) expect(src).not.toMatch(pattern);
    },
  );

  it.each(Object.entries(sources))("%s gives every host Button the GitHub cursor (BR5.2)", (_, src) => {
    // className={BUTTON} is layout.ts's "cursor-pointer".
    for (const button of src.match(/<Button\b[^>]*>/gs) ?? [])
      expect(button).toMatch(/cursor-pointer|\{BUTTON\}|\$\{BUTTON\}/);
  });
});
