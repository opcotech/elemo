export type RouteBreadcrumb =
  | string
  | {
      label: string;
      /** Omit to use the match pathname; `null` makes the crumb non-navigatable. */
      href?: string | null;
    };

export function resolveRouteBreadcrumb(
  breadcrumb: RouteBreadcrumb | ((loaderData: unknown) => RouteBreadcrumb),
  loaderData: unknown
): RouteBreadcrumb {
  return typeof breadcrumb === "function" ? breadcrumb(loaderData) : breadcrumb;
}

export function entityBreadcrumb(
  data: unknown,
  key: string,
  fallback: string
): string {
  if (!data || typeof data !== "object" || !(key in data)) {
    return fallback;
  }

  const entity = (data as Record<string, unknown>)[key];
  if (!entity || typeof entity !== "object" || !("name" in entity)) {
    return fallback;
  }

  const name = (entity as { name: unknown }).name;
  return typeof name === "string" && name.length > 0 ? name : fallback;
}

/** Breadcrumb label when loader data is the named entity itself (e.g. organization). */
export function namedEntityBreadcrumb(data: unknown, fallback: string): string {
  if (!data || typeof data !== "object" || !("name" in data)) {
    return fallback;
  }

  const name = (data as { name: unknown }).name;
  return typeof name === "string" && name.length > 0 ? name : fallback;
}
