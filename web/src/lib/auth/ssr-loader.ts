/**
 * Shared SSR loader helper to parse cookies, pre-refresh tokens if needed,
 * and retry once on unauthorized using the refresh token. The page must
 * provide its own request function.
 */
export async function runSsrLoader<T>({
  context,
  request,
  isJwtExpired: customIsJwtExpired,
}: {
  context: any;
  request: (token?: string) => Promise<T>;
  isJwtExpired?: (token?: string) => boolean;
}): Promise<T> {
  const isBrowser = typeof window !== "undefined";
  if (isBrowser) {
    try {
      return await request();
    } catch {
      return undefined as unknown as T;
    }
  }

  let accessToken: string | undefined = context?.accessToken;
  const refreshToken: string | undefined = context?.refreshToken;

  const isJwtExpired =
    customIsJwtExpired ??
    ((token?: string) => {
      if (!token) return true;
      try {
        const payloadBase64 = token.split(".")[1];
        const decoded =
          typeof Buffer !== "undefined"
            ? Buffer.from(payloadBase64, "base64").toString("utf8")
            : (globalThis as any).atob?.(payloadBase64);
        const { exp } = JSON.parse(decoded ?? "{}");
        return exp ? exp * 1000 < Date.now() + 30000 : false;
      } catch {
        return true;
      }
    });

  const refreshAccessToken = async () => {
    const { authClient } = await import("@/lib/auth/auth-client");
    return authClient.refreshToken(refreshToken!);
  };

  const tryRequest = async () => request(accessToken);

  if ((isJwtExpired(accessToken) || !accessToken) && refreshToken) {
    try {
      accessToken = (await refreshAccessToken()).access_token;
    } catch (err) {
      console.warn("SSR token pre-refresh failed", err);
    }
  }

  try {
    return await tryRequest();
  } catch (err) {
    if (!refreshToken) throw err;
    try {
      accessToken = (await refreshAccessToken()).access_token;
      return await tryRequest();
    } catch (err2) {
      console.warn("SSR retry after refresh also failed", err2);
      throw err2;
    }
  }
}
