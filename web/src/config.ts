import { assertValue } from "./lib/utils";

export const config = {
  auth: () => {
    return {
      debugBanner: import.meta.env.VITE_AUTH_DEBUG_BANNER === "true",
      apiBaseUrl: assertValue(
        import.meta.env.VITE_API_BASE_URL,
        "VITE_API_BASE_URL is required"
      ),
      clientId: assertValue(
        import.meta.env.VITE_AUTH_CLIENT_ID,
        "VITE_AUTH_CLIENT_ID is required"
      ),
      clientSecret: assertValue(
        import.meta.env.VITE_AUTH_CLIENT_SECRET,
        "VITE_AUTH_CLIENT_SECRET is required"
      ),
    };
  },
  env: () => {
    return {
      nodeEnv: import.meta.env.NODE_ENV || "development",
      isDevelopment: import.meta.env.NODE_ENV !== "production",
      isProduction: import.meta.env.NODE_ENV === "production",
    };
  },
};
