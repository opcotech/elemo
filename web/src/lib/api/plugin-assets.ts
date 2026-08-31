import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";

import { buildUpstreamUrl } from "./protocol";

import { authMiddleware } from "@/lib/auth/middleware";
import { getAuthServerEnv } from "@/lib/auth/server-env";
import { refreshSession } from "@/lib/auth/session.server";
import {
  pluginFrontendAssetPath,
  pluginIdPattern,
  versionPattern,
} from "@/lib/plugins/asset-path";

const MAX_FRONTEND_SOURCE_BYTES = 2 * 1024 * 1024;
const JS_ACCEPT = "text/javascript, application/javascript;q=0.9, */*;q=0.1";

const fetchSchema = z.object({
  pluginId: z.string().regex(pluginIdPattern),
  version: z.string().regex(versionPattern),
  entrypoint: z.string().min(1).max(1024),
});

export interface PluginFrontendSource {
  status: number;
  contentType: string;
  source: string;
}

async function fetchUpstreamSource(
  path: string,
  accessToken: string
): Promise<PluginFrontendSource> {
  const { apiBaseUrl } = getAuthServerEnv();
  const headers = new Headers();
  headers.set("authorization", `Bearer ${accessToken}`);
  headers.set("accept", JS_ACCEPT);

  let response = await fetch(buildUpstreamUrl(apiBaseUrl, path), {
    method: "GET",
    headers,
    cache: "no-store",
    redirect: "manual",
  });
  if (response.status === 401) {
    const refreshed = await refreshSession();
    if (refreshed?.accessToken) {
      headers.set("authorization", `Bearer ${refreshed.accessToken}`);
      response = await fetch(buildUpstreamUrl(apiBaseUrl, path), {
        method: "GET",
        headers,
        cache: "no-store",
        redirect: "manual",
      });
    }
  }

  const source = await response.text();
  if (source.length > MAX_FRONTEND_SOURCE_BYTES) {
    return { status: 413, contentType: "", source: "" };
  }
  return {
    status: response.status,
    contentType: response.headers.get("content-type") ?? "",
    source,
  };
}

export const fetchPluginFrontendSourceFn = createServerFn({ method: "POST" })
  .middleware([authMiddleware])
  .validator(fetchSchema)
  .handler(async ({ data, context }): Promise<PluginFrontendSource> => {
    const path = pluginFrontendAssetPath(
      data.pluginId,
      data.version,
      data.entrypoint
    );
    if (!path) {
      return { status: 400, contentType: "", source: "" };
    }
    return fetchUpstreamSource(path, context.accessToken);
  });
