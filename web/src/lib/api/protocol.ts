import { z } from "zod";

const ALLOWED_REQUEST_HEADERS = new Set(["accept", "content-type"]);
const FORBIDDEN_HEADER_VALUE = /[\r\n\0]/;
const MAX_BODY_BYTES = 2 * 1024 * 1024;

export const apiMethodSchema = z.enum([
  "GET",
  "POST",
  "PUT",
  "PATCH",
  "DELETE",
  "HEAD",
]);

export const transportRequestSchema = z
  .object({
    method: apiMethodSchema,
    path: z.string().min(1).max(4096),
    headers: z.record(z.string(), z.string()).default({}),
    body: z.string().max(MAX_BODY_BYTES).optional(),
  })
  .strict()
  .superRefine((request, context) => {
    if (!isSafeApiPath(request.path)) {
      context.addIssue({
        code: "custom",
        path: ["path"],
        message: "Only relative /v1 API paths are allowed",
      });
    }

    for (const [name, value] of Object.entries(request.headers)) {
      if (
        !ALLOWED_REQUEST_HEADERS.has(name.toLowerCase()) ||
        FORBIDDEN_HEADER_VALUE.test(name) ||
        FORBIDDEN_HEADER_VALUE.test(value)
      ) {
        context.addIssue({
          code: "custom",
          path: ["headers", name],
          message: "Header is not allowed",
        });
      }
    }

    if (
      ["GET", "HEAD"].includes(request.method) &&
      request.body !== undefined
    ) {
      context.addIssue({
        code: "custom",
        path: ["body"],
        message: `${request.method} requests cannot contain a body`,
      });
    }
  });

export type ApiTransportRequest = z.infer<typeof transportRequestSchema>;

export interface ApiTransportResponse {
  status: number;
  statusText: string;
  headers: Record<string, string>;
  body: string;
}

export function isSafeApiPath(path: string): boolean {
  if (
    !path.startsWith("/v1/") ||
    path.startsWith("//") ||
    path.includes("\\") ||
    path.includes("#") ||
    /[\r\n\0]/.test(path)
  ) {
    return false;
  }

  try {
    const parsed = new URL(path, "https://bff.invalid");
    return (
      parsed.origin === "https://bff.invalid" &&
      parsed.pathname.startsWith("/v1/") &&
      !parsed.pathname.includes("/../") &&
      !parsed.pathname.includes("/./")
    );
  } catch {
    return false;
  }
}

/** Join a validated relative API path with the upstream Go API base URL. */
export function buildUpstreamUrl(apiBaseUrl: string | URL, path: string): URL {
  if (!isSafeApiPath(path)) {
    throw new Error("Only relative /v1 API paths are allowed");
  }
  return new URL(path, apiBaseUrl);
}

export function isPublicApiRequest(
  request: Pick<ApiTransportRequest, "method" | "path">
): boolean {
  const pathname = new URL(request.path, "https://bff.invalid").pathname;

  if (
    pathname === "/v1/users/reset" &&
    (request.method === "GET" || request.method === "POST")
  ) {
    return true;
  }

  return (
    request.method === "POST" &&
    /^\/v1\/organizations\/[^/]+\/members\/accept$/.test(pathname)
  );
}
