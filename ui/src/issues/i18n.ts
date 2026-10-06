import type { PluginHostApi, PluginRegistry } from "@kandev/plugin-sdk";

import { en, type MessageKey, type Messages } from "../messages/en";

/**
 * Registers the English catalogue with Kandev (US8.5). Placeholders stay in
 * the plugin's {name} form: the plugin fills them itself with format(), so
 * Kandev's {{name}} interpolation is not needed.
 */
export function registerMessages(registry: Pick<PluginRegistry, "registerTranslations">): void {
  registry.registerTranslations({ en });
}

/**
 * The catalogue in Kandev's language: each key through host.i18n.t with the
 * English text as the default, so a missing message is English, never a key
 * (AC8.5.2). Read when a screen is created; a language change applies after
 * a reload.
 */
export function messagesFor(host: PluginHostApi): Messages {
  const i18n = (host as Partial<PluginHostApi>).i18n;
  if (!i18n) return en;
  const out = {} as Record<MessageKey, string>;
  for (const key of Object.keys(en) as MessageKey[]) {
    out[key] = i18n.t(key, { defaultValue: en[key] }) || en[key];
  }
  return out;
}
