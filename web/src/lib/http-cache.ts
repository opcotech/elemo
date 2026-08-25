const ASSET_PATH_PREFIX = "/assets/";

/**
 * Cache-Control for responses that TanStack Start handles (SSR HTML and
 * server functions). Hashed Vite assets keep their own immutable headers.
 */
export function cacheControlForRequestPath(pathname: string): string | null {
  if (pathname === "/assets" || pathname.startsWith(ASSET_PATH_PREFIX)) {
    return null;
  }
  return "private, no-store";
}
