// JSX typing for the host-provided element factory (host.jsx). The plugin
// bundles no React, so intrinsic elements are typed loosely here.
declare namespace JSX {
  type Element = unknown;
  interface IntrinsicElements {
    [tag: string]: Record<string, unknown>;
  }
  interface ElementChildrenAttribute {
    children: unknown;
  }
}
