# Backlog logo

The plugin icon (settings card, home Integrations entry and `/backlog` title bar) is the
official Backlog product icon, drawn as inline SVG in `ui/src/brand/backlog-logo.tsx`.

## Source

- Page: Nulab press media assets, <https://nulab.com/press/media-assets/>
  ("Logos and icons", download "All logos").
- File: `Nulab-all-logos.zip` (<https://nulab.com/npma/Nulab-all-logos.zip>),
  entry `all-logos/Icons/Backlog/Backlog-Product-Icon-Color.svg`
  (SHA-256 of the SVG: `10a62c797ddb5ca4b66fa845f57d9aa66cc228a696f287f2df8c8f6936dd202b`).
- Retrieved: 2026-10-06.

The two `path` elements (shape and fill colours `#42CE9F` and `white`) are copied unchanged.
Only the wrapper differs: the SVG keeps the original `viewBox="0 0 48 48"`, drops the fixed
`width`/`height` so the host can size it with a CSS class, and adds `aria-hidden="true"` and
`focusable="false"`. The proportions and colours are not changed.

The logo is compiled into `ui/bundle.js`; the plugin never fetches it at runtime.
`make verify-package` fails if the bundle references any Nulab or Backlog asset URL (NFR3.10).

## Terms

- Nulab logo guidelines, <https://nulab.com/logo-guidelines/>: company and product logos
  "should be used for development and marketing purposes ONLY". It is not allowed to create a
  modified version, integrate the logo into your own logo, animate it or add effects, or change
  its colours or dimensions or add text or images to it. Nulab may change these permissions at
  any time; questions go to info@nulab.com.
- The media assets page says its downloads may be used "for editorial purposes only with
  credit: 'Copyright: Nulab, Inc'", and that other use, including commercial use, is prohibited,
  and points to the logo guidelines above for logo use.

Backlog and the Backlog logo are trademarks of Nulab, Inc. Copyright: Nulab, Inc.
This plugin is not made or endorsed by Nulab.

## Before release

**The user must confirm that Nulab's brand guidelines allow this use of the logo before the
first release** (functional-design Q8; the two statements above are not fully consistent for a
third-party plugin). If they do not, switch to a host built-in icon by changing the one line
that defines `PLUGIN_ICON` in `ui/src/brand/backlog-logo.tsx`; nothing else changes.
