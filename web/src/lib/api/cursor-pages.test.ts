import { describe, expect, it, vi } from "vitest";

import {
  MAX_CURSOR_PAGES,
  collectCursorPages,
  collectListedPage,
  collectedQueryKey,
  cursorPageQuery,
} from "./cursor-pages";
import type { CursorPage } from "./cursor-pages";

function page(
  items: string[],
  pageInfo: CursorPage<string>["page_info"]
): CursorPage<string> {
  return { items, page_info: pageInfo };
}

describe("collectCursorPages", () => {
  it("returns a single page when has_more is false", async () => {
    const fetchPage = vi.fn(() =>
      Promise.resolve(page(["one", "two"], { has_more: false }))
    );

    await expect(collectCursorPages(fetchPage)).resolves.toEqual([
      "one",
      "two",
    ]);
    expect(fetchPage).toHaveBeenCalledTimes(1);
    expect(fetchPage).toHaveBeenCalledWith(undefined);
  });

  it("follows next_page_token until has_more is false", async () => {
    const fetchPage = vi.fn((pageToken?: string) => {
      if (!pageToken) {
        return Promise.resolve(
          page(["one"], {
            has_more: true,
            next_page_token: "cursor-2",
          })
        );
      }
      if (pageToken === "cursor-2") {
        return Promise.resolve(
          page(["two"], {
            has_more: true,
            next_page_token: "cursor-3",
          })
        );
      }
      return Promise.resolve(page(["three"], { has_more: false }));
    });

    await expect(collectCursorPages(fetchPage)).resolves.toEqual([
      "one",
      "two",
      "three",
    ]);
    expect(fetchPage).toHaveBeenCalledTimes(3);
    expect(fetchPage).toHaveBeenNthCalledWith(1, undefined);
    expect(fetchPage).toHaveBeenNthCalledWith(2, "cursor-2");
    expect(fetchPage).toHaveBeenNthCalledWith(3, "cursor-3");
  });

  it("stops when has_more is true but next_page_token is missing", async () => {
    const fetchPage = vi.fn(() =>
      Promise.resolve(page(["one"], { has_more: true, next_page_token: null }))
    );

    await expect(collectCursorPages(fetchPage)).resolves.toEqual(["one"]);
    expect(fetchPage).toHaveBeenCalledTimes(1);
  });

  it("caps collection at MAX_CURSOR_PAGES", async () => {
    let pageNumber = 0;
    const fetchPage = vi.fn(() => {
      pageNumber += 1;
      return Promise.resolve(
        page([`item-${pageNumber}`], {
          has_more: true,
          next_page_token: `cursor-${pageNumber + 1}`,
        })
      );
    });

    const items = await collectCursorPages(fetchPage);

    expect(items).toEqual(
      Array.from(
        { length: MAX_CURSOR_PAGES },
        (_, index) => `item-${index + 1}`
      )
    );
    expect(fetchPage).toHaveBeenCalledTimes(MAX_CURSOR_PAGES);
  });

  it("treats a missing page as an empty terminal page", async () => {
    const fetchPage = vi.fn(() => Promise.resolve(undefined));

    await expect(collectCursorPages(fetchPage)).resolves.toEqual([]);
    expect(fetchPage).toHaveBeenCalledTimes(1);
  });

  it("wraps collected items as a terminal page", async () => {
    const fetchPage = vi.fn(() =>
      Promise.resolve(page(["one"], { has_more: false }))
    );

    await expect(collectListedPage(fetchPage)).resolves.toEqual({
      items: ["one"],
      page_info: { has_more: false },
    });
  });

  it("appends a collected suffix to query keys", () => {
    expect(collectedQueryKey(["issues", { page_size: 100 }])).toEqual([
      "issues",
      { page_size: 100 },
      "collected",
    ]);
  });

  it("builds cursor page query params", () => {
    expect(cursorPageQuery()).toEqual({ page_size: 100 });
    expect(cursorPageQuery("cursor-2")).toEqual({
      page_size: 100,
      page_token: "cursor-2",
    });
  });
});
