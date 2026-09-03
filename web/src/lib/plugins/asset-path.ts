import { isSafeApiPath } from "@/lib/api/protocol";

export const pluginIdPattern = /^[a-z][a-z0-9]*(\.[a-z0-9]+)+$/;
export const versionPattern = /^[A-Za-z0-9._+-]+$/;

export function isSafeAssetRel(rel: string): boolean {
  return (
    rel.length > 0 &&
    rel.length < 1024 &&
    !rel.startsWith("/") &&
    !rel.includes("\\") &&
    !rel.includes("..") &&
    !rel.includes("\0")
  );
}

export function normalizeAssetRel(entrypoint: string): string {
  return entrypoint.replace(/^\/+/, "");
}

export function pluginFrontendAssetPath(
  pluginId: string,
  version: string,
  entrypoint: string
): string | undefined {
  const rel = normalizeAssetRel(entrypoint);
  if (
    !pluginIdPattern.test(pluginId) ||
    !versionPattern.test(version) ||
    !isSafeAssetRel(rel)
  ) {
    return undefined;
  }
  const path = `/v1/plugins/${pluginId}/assets/${version}/${rel}`;
  return isSafeApiPath(path) ? path : undefined;
}

export function isPluginJavaScriptSource(
  source: string,
  contentType: string
): boolean {
  const type = contentType.toLowerCase();
  if (type.includes("javascript") || type.includes("ecmascript")) {
    return true;
  }
  const start = source.trimStart();
  return start.startsWith("import ") || start.startsWith("export ");
}
