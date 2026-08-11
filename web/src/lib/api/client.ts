import { toApiError } from "./errors";
import { isPublicApiRequest, transportRequestSchema } from "./protocol";

import { client } from "@/lib/client/client.gen";

const BFF_BASE_URL = "https://elemo-bff.invalid/api";
let configured = false;

async function bffFetch(input: RequestInfo | URL, init?: RequestInit) {
  const request = new Request(input, init);
  const url = new URL(request.url);
  const path = `${url.pathname.replace(/^\/api/, "")}${url.search}`;
  const headers: Record<string, string> = {};

  for (const name of ["accept", "content-type"]) {
    const value = request.headers.get(name);
    if (value) {
      headers[name] = value;
    }
  }

  const transportRequest = transportRequestSchema.parse({
    method: request.method,
    path,
    headers,
    body:
      request.method === "GET" || request.method === "HEAD"
        ? undefined
        : await request.text(),
  });

  try {
    const { protectedApiTransport, publicApiTransport } =
      await import("./transport");
    const transport = isPublicApiRequest(transportRequest)
      ? publicApiTransport
      : protectedApiTransport;
    const response = await transport({
      data: transportRequest,
      signal: request.signal,
    });

    return new Response(response.body || null, {
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
    });
  } catch (error) {
    const status =
      error && typeof error === "object" && "status" in error
        ? Number(error.status)
        : 500;
    return Response.json(
      {
        status: Number.isInteger(status) ? status : 500,
        message: status === 401 ? "Authentication required" : "Request failed",
      },
      {
        status: Number.isInteger(status) ? status : 500,
        headers: {
          "cache-control": "private, no-store",
        },
      }
    );
  }
}

export function ensureApiClientConfigured(): void {
  if (configured) {
    return;
  }

  client.setConfig({
    baseUrl: BFF_BASE_URL,
    cache: "no-store",
    fetch: bffFetch,
  });
  client.interceptors.error.use((error, response) =>
    toApiError(error, response)
  );
  configured = true;
}

ensureApiClientConfigured();

export { client };
