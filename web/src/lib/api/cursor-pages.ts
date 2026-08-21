import type { PageInfo } from "@/lib/client";

/** Default list page size used by generated list endpoints. */
export const DEFAULT_LIST_PAGE_SIZE = 100;

/**
 * Safety cap while following `next_page_token`. Ten default pages equals the
 * backend MaxPageSize of 1000.
 */
export const MAX_CURSOR_PAGES = 10;

export interface CursorPage<T> {
  readonly items?: readonly T[] | null;
  readonly page_info?: Pick<PageInfo, "has_more" | "next_page_token">;
}

/** Follow cursor pages until `!has_more`, or `maxPages` is reached. */
export async function collectCursorPages<T>(
  fetchPage: (pageToken?: string) => Promise<CursorPage<T> | null | undefined>,
  maxPages = MAX_CURSOR_PAGES
): Promise<T[]> {
  const items: T[] = [];
  let pageToken: string | undefined;

  for (let page = 0; page < maxPages; page += 1) {
    const result = await fetchPage(pageToken);
    if (result?.items) {
      items.push(...result.items);
    }

    const nextToken = result?.page_info?.next_page_token;
    if (!result?.page_info?.has_more || !nextToken) {
      break;
    }
    pageToken = nextToken;
  }

  return items;
}

export function collectedQueryKey(
  queryKey: readonly unknown[]
): readonly unknown[] {
  return [...queryKey, "collected"];
}

export function cursorPageQuery(pageToken?: string) {
  return {
    page_size: DEFAULT_LIST_PAGE_SIZE,
    ...(pageToken ? { page_token: pageToken } : {}),
  };
}

export function cursorPageQueryWith<TQuery extends Record<string, unknown>>(
  query: TQuery,
  pageToken?: string
): TQuery & { page_size: number; page_token?: string } {
  return {
    ...query,
    page_size: DEFAULT_LIST_PAGE_SIZE,
    ...(pageToken ? { page_token: pageToken } : {}),
  };
}

export function nextCursorPageToken<T>(
  page: CursorPage<T> | null | undefined
): string | undefined {
  if (!page?.page_info?.has_more) {
    return undefined;
  }
  return page.page_info.next_page_token ?? undefined;
}

export function flattenCursorPages<T>(
  pages: readonly (CursorPage<T> | null | undefined)[]
): T[] {
  const out: T[] = [];
  for (const page of pages) {
    if (page?.items?.length) {
      out.push(...page.items);
    }
  }
  return out;
}

export async function collectListedPage<T>(
  fetchPage: (pageToken?: string) => Promise<CursorPage<T> | null | undefined>,
  maxPages = MAX_CURSOR_PAGES
): Promise<{ items: T[]; page_info: { has_more: false } }> {
  const items = await collectCursorPages(fetchPage, maxPages);
  return {
    items,
    page_info: { has_more: false },
  };
}

export function collectedListQuery<TItem>(
  listOptions: {
    queryKey: readonly unknown[];
    staleTime?: unknown;
    gcTime?: unknown;
  },
  fetchPage: (
    pageToken: string | undefined,
    signal: AbortSignal
  ) => Promise<CursorPage<TItem> | null | undefined>
) {
  return {
    staleTime: listOptions.staleTime as number | undefined,
    gcTime: listOptions.gcTime as number | undefined,
    queryKey: collectedQueryKey(listOptions.queryKey),
    queryFn: async ({
      signal,
    }: {
      signal: AbortSignal;
    }): Promise<{ items: TItem[]; page_info: { has_more: false } }> =>
      collectListedPage((pageToken) => fetchPage(pageToken, signal)),
  };
}
