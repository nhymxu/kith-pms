// Captured once at module evaluation time (app boot, before the router mounts
// and normalizes the URL bar with schema-defaulted search params). Routes
// that need to distinguish "no search params in the URL the user actually
// navigated to" from "params defaulted in by validateSearch" must read this
// instead of `window.location.search`, which reflects the post-normalization
// state by the time any route component's effects run.
export const initialSearch = window.location.search;
