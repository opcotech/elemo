import type { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import type { Document } from "@/lib/api/types";
import { loadDocumentPage } from "@/lib/documents/load-document";

const liveDocument: Document = {
  id: "document-live",
  title: "Project Plan",
  excerpt: "Overview of the project plan",
  content: "# Project Plan\n\nGoals and timeline.",
  created_by: {
    id: "user-1",
    first_name: "Test",
    last_name: "User",
    picture: null,
  },
  labels: [],
  library: {
    id: "namespace-1",
    type: "Namespace",
    name: "Product",
  },
  folder: null,
  relations: [],
  comment_count: 0,
  attachment_count: 0,
  created_at: "2026-08-01T00:00:00Z",
  updated_at: "2026-08-10T00:00:00Z",
};

function createQueryClient(result: unknown) {
  return {
    fetchQuery: vi.fn(() =>
      result instanceof Error ? Promise.reject(result) : Promise.resolve(result)
    ),
  } as unknown as QueryClient;
}

describe("loadDocumentPage", () => {
  it("returns a live document from the API", async () => {
    const queryClient = createQueryClient(liveDocument);

    await expect(
      loadDocumentPage(queryClient, "document-live")
    ).resolves.toEqual({
      document: liveDocument,
    });
  });

  it("throws not-found when the API returns 404", async () => {
    const queryClient = createQueryClient(new ApiError(404, "missing"));

    await expect(
      loadDocumentPage(queryClient, "document-missing")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("throws not-found for permission errors", async () => {
    const queryClient = createQueryClient(new ApiError(403, "denied"));

    await expect(
      loadDocumentPage(queryClient, "document-live")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("propagates unexpected API failures", async () => {
    const error = new ApiError(500, "boom");
    const queryClient = createQueryClient(error);

    await expect(loadDocumentPage(queryClient, "document-live")).rejects.toBe(
      error
    );
  });
});
