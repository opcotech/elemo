/**
 * Explicit lazy boundary for fixture-backed command palette entries.
 *
 * Keep consumers on a dynamic import so the authenticated shell does not load
 * the full mock data set before the command palette opens.
 */
export { mockGlobalSearchEntries as mockCommandSearchEntries } from "./fixtures";
