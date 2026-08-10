import { useRouter } from "@tanstack/react-router";
import { useContext, useState } from "react";

import { AuthContext } from "@/lib/auth/auth-context";
import { sanitizeRedirectTarget } from "@/lib/auth/redirect";
import type { LoginCredentials } from "@/lib/auth/types";

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

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
      await router.invalidate();

      const targetPath = sanitizeRedirectTarget(
        redirectTo,
        window.location.origin
      );

      await router.navigate({ href: targetPath });
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

export const useLogout = () => {
  const [isLoading, setIsLoading] = useState(false);
  const { logout: authLogout } = useAuth();
  const router = useRouter();

  const logout = async () => {
    try {
      setIsLoading(true);

      await authLogout();
      await router.invalidate();

      await router.navigate({
        to: "/login",
        search: {
          redirect: undefined,
        },
      });
    } catch {
      await router.invalidate();
      await router.navigate({
        to: "/login",
        search: {
          redirect: undefined,
        },
      });
    } finally {
      setIsLoading(false);
    }
  };

  return {
    logout,
    isLoading,
  };
};
