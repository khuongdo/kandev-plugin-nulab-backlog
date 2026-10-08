// The plugin ships no CSS: layouts use the host's utility classes only (BR6.5).

/** One vertical stack with one gap. */
export const STACK = "flex flex-col gap-4";
/** A label above its control, with a smaller gap. */
export const FIELD = "flex flex-col gap-2";
/** Controls side by side, wrapping on narrow screens. */
export const ROW = "flex flex-wrap items-center gap-2";
/** A framed card that sets one block apart, such as the active source control service (FR4.1). */
export const CARD = "flex flex-col gap-4 rounded-lg border p-4";
/** Every host Button gets the pointer cursor, like the GitHub integration (BR5.2). */
export const BUTTON = "cursor-pointer";
/** The list toolbar under the scope bar, like the host's IntegrationListToolbar (FR5.2). */
export const TOOLBAR =
  "flex shrink-0 flex-col gap-2 border-b px-4 py-2.5 sm:px-6 md:flex-row md:flex-wrap md:items-center md:gap-3";
/** The results area under the toolbar, like the GitHub page (FR5.3). */
export const RESULTS = "flex flex-col gap-4 px-3 py-4 md:px-6";
/** A toolbar's filter slot: stacked at full width on phones, one row from md (BR4.8, R-02). */
export const FILTERS = "flex w-full flex-col gap-2 md:w-auto md:flex-row md:items-center";
/** IntegrationRepositoryFilter's trigger, as on the GitHub list: full width on phones, 220 px from md. */
export const FILTER_TRIGGER =
  "w-full border border-input bg-background px-2 text-xs/relaxed hover:bg-secondary/50 md:w-[220px]";
/** IntegrationRepositoryFilter's list, as on the GitHub list. */
export const FILTER_POPOVER = "md:min-w-[360px]";
