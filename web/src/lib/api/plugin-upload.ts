import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";

import { buildUpstreamUrl } from "./protocol";

import { authMiddleware } from "@/lib/auth/middleware";
import { getAuthServerEnv } from "@/lib/auth/server-env";
import { refreshSession } from "@/lib/auth/session.server";

const MAX_PLUGIN_PACKAGE_BYTES = 32 * 1024 * 1024;
const pluginIdPattern = /^[a-z][a-z0-9]*(\.[a-z0-9]+)+$/;

const uploadSchema = z.object({
  filename: z.string().min(1).max(255),
  bytes: z.string().min(1),
  pluginId: z.string().regex(pluginIdPattern).optional(),
});

function decodePackage(bytes: string): Uint8Array {
  return new Uint8Array(Buffer.from(bytes, "base64"));
}

async function sendPackage(
  path: string,
  filename: string,
  zip: Uint8Array,
  accessToken: string
): Promise<{ status: number; body: string }> {
  const { apiBaseUrl } = getAuthServerEnv();
  const form = new FormData();
  const copy = new ArrayBuffer(zip.byteLength);
  new Uint8Array(copy).set(zip);
  form.append(
    "package",
    new File([copy], filename, { type: "application/zip" })
  );

  const headers = new Headers();
  headers.set("authorization", `Bearer ${accessToken}`);

  let response = await fetch(buildUpstreamUrl(apiBaseUrl, path), {
    method: "POST",
    headers,
    body: form,
    cache: "no-store",
    redirect: "manual",
  });

  if (response.status === 401) {
    const refreshed = await refreshSession();
    if (refreshed?.accessToken) {
      headers.set("authorization", `Bearer ${refreshed.accessToken}`);
      response = await fetch(buildUpstreamUrl(apiBaseUrl, path), {
        method: "POST",
        headers,
        body: form,
        cache: "no-store",
        redirect: "manual",
      });
    }
  }

  return { status: response.status, body: await response.text() };
}

export const uploadPluginPackageFn = createServerFn({ method: "POST" })
  .middleware([authMiddleware])
  .validator(uploadSchema)
  .handler(async ({ data, context }) => {
    const zip = decodePackage(data.bytes);
    if (zip.byteLength > MAX_PLUGIN_PACKAGE_BYTES) {
      return {
        status: 400,
        body: JSON.stringify({ message: "Plugin package exceeds 32MB" }),
      };
    }
    const path = data.pluginId
      ? `/v1/plugins/${encodeURIComponent(data.pluginId)}/upgrade`
      : "/v1/plugins";
    return sendPackage(path, data.filename, zip, context.accessToken);
  });
