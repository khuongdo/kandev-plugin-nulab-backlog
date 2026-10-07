# Plugin icon

The plugin icon (settings card, home Integrations entry and `/backlog` title bar) is an
original outline drawing, defined in `ui/src/brand/backlog-logo.tsx` behind the single
`PLUGIN_ICON` selection point.

## The icon

- A 24×24 outline of a ticket card with two text lines and a check mark: one `rect` and three
  `path` elements, `fill="none"`, `stroke="currentColor"`, `stroke-width="2"`, round caps and
  joins, so it takes the surrounding text colour like Kandev's own outline (Tabler) icons.
- No fixed `width`/`height`: the host sizes it with a CSS class (`h-4 w-4` in the nav,
  `h-5 w-5` on the settings card). It is decorative (`aria-hidden="true"`, `focusable="false"`).
- It is compiled into `ui/bundle.js`; the plugin never fetches an icon at runtime.
  `make verify-package` still fails if the bundle references any Nulab or Backlog asset URL
  (NFR3.10).

## The Nulab mark is no longer shipped

Earlier versions drew the official Backlog product icon (the filled `#42CE9F` tile with the white
mark) from Nulab's media kit. Intent 261007 replaced it (FR6): Kandev's other integration icons are
outline only, and redrawing the Nulab mark as an outline would be a modified logo, which Nulab's
logo guidelines (<https://nulab.com/logo-guidelines/>) do not allow. The new icon shares no path or
shape with the Nulab mark and uses no Nulab colour.

Backlog is a trademark of Nulab, Inc. This plugin is not made or endorsed by Nulab. The name
"Backlog" is used only to say which service the plugin connects to.
