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
  fallback: string,
  property = "name"
): string {
  if (!data || typeof data !== "object" || !(key in data)) {
    return fallback;
  }

  const entity = (data as Record<string, unknown>)[key];
  if (!entity || typeof entity !== "object" || !(property in entity)) {
    return fallback;
  }

  const value = (entity as Record<string, unknown>)[property];
  return typeof value === "string" && value.length > 0 ? value : fallback;
}

/** Breadcrumb label when loader data is the named entity itself (e.g. organization). */
export function namedEntityBreadcrumb(data: unknown, fallback: string): string {
  if (!data || typeof data !== "object" || !("name" in data)) {
    return fallback;
  }

  const name = (data as { name: unknown }).name;
  return typeof name === "string" && name.length > 0 ? name : fallback;
}
