import { describe, expect, it } from "vitest";
import type { PluginHostApi, PluginRegistry } from "@kandev/plugin-sdk";

import { en } from "../messages/en";
import { fakeHost, fakeI18n, fakeRegistry } from "../testing/harness";
import { messagesFor, registerMessages } from "./i18n";

const hostWith = (locale: string, catalogue: Record<string, string> = {}) =>
  fakeHost(async () => ({}), { i18n: fakeI18n(locale, catalogue) });

describe("Text follows the Kandev language (US8.5)", () => {
  it("registers the English catalogue", () => {
    const registry = fakeRegistry();
    registerMessages(registry as unknown as PluginRegistry);
    expect(registry.registerTranslations).toHaveBeenCalledTimes(1);
    const catalogs = registry.registerTranslations.mock.calls[0]![0] as Record<
      string,
      Record<string, string>
    >;
    expect(catalogs.en).toEqual(en);
    expect(Object.keys(catalogs.en!).length).toBeLessThanOrEqual(1000);
    for (const v of Object.values(catalogs.en!)) expect(v.length).toBeLessThanOrEqual(4096);
  });

  it("looks every key up with the English text as the default", () => {
    const seen: [string, unknown][] = [];
    const host = fakeHost(async () => ({}), {
      i18n: {
        locale: "en",
        t: (key: string, o: { defaultValue: string }) => {
          seen.push([key, o.defaultValue]);
          return o.defaultValue;
        },
      },
    });
    const m = messagesFor(host);
    expect(m).toEqual(en);
    expect(seen).toHaveLength(Object.keys(en).length);
    expect(seen).toContainEqual(["issuesEmpty", en.issuesEmpty]);
  });

  it("shows another locale's text (AC8.5.1)", () => {
    const pseudo = Object.fromEntries(Object.entries(en).map(([k, v]) => [k, `[ps] ${v}`]));
    const m = messagesFor(hostWith("x-pseudo", pseudo));
    expect(m.issuesEmpty).toBe(`[ps] ${en.issuesEmpty}`);
    expect(m.linkToTask).toBe(`[ps] ${en.linkToTask}`);
  });

  it("falls back to English, never a raw key (AC8.5.2)", () => {
    const m = messagesFor(hostWith("fr", { linkToTask: "Lier à une tâche" }));
    expect(m.linkToTask).toBe("Lier à une tâche");
    expect(m.issuesEmpty).toBe(en.issuesEmpty);
    for (const [k, v] of Object.entries(m)) expect(v, k).not.toBe(k);
  });

  it("uses English on a host without i18n", () => {
    const host = { ...fakeHost(async () => ({})), i18n: undefined } as unknown as PluginHostApi;
    expect(messagesFor(host)).toEqual(en);
  });
});
