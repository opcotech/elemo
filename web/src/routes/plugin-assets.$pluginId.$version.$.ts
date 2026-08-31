import { createFileRoute } from "@tanstack/react-router";

import { buildUpstreamUrl } from "@/lib/api/protocol";
import { getAuthServerEnv } from "@/lib/auth/server-env";
import {
  AuthenticationError,
  refreshSession,
  requireSessionTokens,
} from "@/lib/auth/session.server";
import { pluginFrontendAssetPath } from "@/lib/plugins/asset-path";

export const Route = createFileRoute("/plugin-assets/$pluginId/$version/$")({
  ssr: false,
  component: () => null,
  server: {
    handlers: {
      GET: async ({ params }) => {
        const pluginId = params.pluginId;
        const version = params.version;
        const rel = (params._splat ?? "").replace(/^\/+/, "");
        const path = pluginFrontendAssetPath(pluginId, version, rel);
        if (!path) {
          return new Response("Not found", {
            status: 404,
            headers: { "cache-control": "private, no-store" },
          });
        }
        try {
          const session = await requireSessionTokens();
          const { apiBaseUrl } = getAuthServerEnv();
          const headers = new Headers();
          headers.set("authorization", `Bearer ${session.accessToken}`);
          headers.set(
            "accept",
            "text/javascript, application/javascript;q=0.9, */*;q=0.1"
          );
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
          const out = new Headers();
          const contentType = response.headers.get("content-type");
          if (contentType) {
            out.set("content-type", contentType);
          }
          out.set(
            "cache-control",
            response.ok ? "private, max-age=60" : "private, no-store"
          );
          out.set("x-content-type-options", "nosniff");
          return new Response(response.body, {
            status: response.status,
            headers: out,
          });
        } catch (error) {
          if (error instanceof AuthenticationError) {
            return new Response("Unauthorized", {
              status: 401,
              headers: { "cache-control": "private, no-store" },
            });
          }
          return new Response("Not found", {
            status: 404,
            headers: { "cache-control": "private, no-store" },
          });
        }
      },
    },
  },
});
