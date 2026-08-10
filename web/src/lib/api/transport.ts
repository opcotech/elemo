import { createServerFn } from "@tanstack/react-start";

import {
  buildUpstreamUrl,
  isPublicApiRequest,
  transportRequestSchema,
} from "./protocol";
import type { ApiTransportRequest, ApiTransportResponse } from "./protocol";

import { authMiddleware } from "@/lib/auth/middleware";
import { getAuthServerEnv } from "@/lib/auth/server-env";
import { refreshSession } from "@/lib/auth/session.server";

const RESPONSE_HEADERS = ["content-type", "retry-after"] as const;

function normalizeFailureBody(response: Response, body: string): string {
  if (response.status < 500) {
    return body;
  }

  return JSON.stringify({
    status: response.status,
    message: "The upstream service could not complete the request",
  });
}

async function serializeResponse(
  response: Response
): Promise<ApiTransportResponse> {
  let body = await response.text();
  if (!response.ok) {
    body = normalizeFailureBody(response, body);
  }

  const headers: Record<string, string> = {
    "cache-control": "private, no-store",
  };
  for (const name of RESPONSE_HEADERS) {
    const value = response.headers.get(name);
    if (value) {
      headers[name] = value;
    }
  }
  if (!headers["content-type"] && body) {
    headers["content-type"] = "application/json";
  }

  return {
    status: response.status,
    statusText: response.statusText,
    headers,
    body,
  };
}

async function sendUpstream(
  request: ApiTransportRequest,
  accessToken?: string
): Promise<Response> {
  const { apiBaseUrl } = getAuthServerEnv();
  const headers = new Headers(request.headers);
  if (accessToken) {
    headers.set("authorization", `Bearer ${accessToken}`);
  }

  // request.path is already a relative Go API path (e.g. "/v1/notifications").
  return fetch(buildUpstreamUrl(apiBaseUrl, request.path), {
    method: request.method,
    headers,
    body: request.body,
    cache: "no-store",
    redirect: "manual",
  });
}

async function proxyProtectedRequest(
  request: ApiTransportRequest,
  accessToken: string
): Promise<ApiTransportResponse> {
  let response = await sendUpstream(request, accessToken);

  if (response.status === 401) {
    const refreshed = await refreshSession();
    if (refreshed?.accessToken) {
      response = await sendUpstream(request, refreshed.accessToken);
    }
  }

  return serializeResponse(response);
}

export const protectedApiTransport = createServerFn({ method: "POST" })
  .middleware([authMiddleware])
  .validator(transportRequestSchema)
  .handler(({ data, context }) =>
    proxyProtectedRequest(data, context.accessToken)
  );

export const publicApiTransport = createServerFn({ method: "POST" })
  .validator(transportRequestSchema)
  .handler(async ({ data }) => {
    if (!isPublicApiRequest(data)) {
      return {
        status: 403,
        statusText: "Forbidden",
        headers: {
          "cache-control": "private, no-store",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          status: 403,
          message: "Endpoint is not available through the public transport",
        }),
      } satisfies ApiTransportResponse;
    }

    return serializeResponse(await sendUpstream(data));
  });
