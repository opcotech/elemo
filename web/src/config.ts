import { assertValue } from "./lib/utils";

// Safely access import.meta.env with fallback for Node.js environments
const getEnv = (key: string): string | undefined => {
  if (typeof import.meta !== "undefined" && import.meta.env) {
    return import.meta.env[key];
  }
  // Fallback to process.env for Node.js environments (e.g., tests)
  if (typeof process !== "undefined" && process.env) {
    return process.env[key];
  }
  return undefined;
};

export const config = {
  auth: () => {
    return {
      debugBanner: getEnv("VITE_AUTH_DEBUG_BANNER") === "true",
      apiBaseUrl: assertValue(
        getEnv("VITE_API_BASE_URL"),
        "VITE_API_BASE_URL is required"
      ),
      clientId: assertValue(
        getEnv("VITE_AUTH_CLIENT_ID"),
        "VITE_AUTH_CLIENT_ID is required"
      ),
      clientSecret: assertValue(
        getEnv("VITE_AUTH_CLIENT_SECRET"),
        "VITE_AUTH_CLIENT_SECRET is required"
      ),
    };
  },
  env: () => {
    const nodeEnv = getEnv("NODE_ENV") || "development";
    return {
      nodeEnv,
      isDevelopment: nodeEnv !== "production",
      isProduction: nodeEnv === "production",
    };
  },
};
