import { useRouter } from "@tanstack/react-router";
import { useState } from "react";

import { useAuth } from "./use-auth";

import type { LoginCredentials } from "@/lib/auth/types";

export const useLogin = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { login: authLogin } = useAuth();
  const router = useRouter();

  const login = async (credentials: LoginCredentials, redirectTo?: string) => {
    try {
      setIsLoading(true);
      setError(null);

      await authLogin(credentials);

      // Redirect to intended destination or dashboard
      let targetPath = "/dashboard";

      if (redirectTo) {
        try {
          // If redirectTo is a full URL, extract the pathname
          const url = new URL(redirectTo);
          targetPath = url.pathname;
        } catch {
          // If it's not a valid URL, treat it as a path
          targetPath = redirectTo.startsWith("/")
            ? redirectTo
            : `/${redirectTo}`;
        }
      }

      await router.navigate({
        to: targetPath,
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Login failed";
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  const clearError = () => setError(null);

  return {
    login,
    isLoading,
    error,
    clearError,
  };
};
