import { useRouter } from "@tanstack/react-router";
import { useState } from "react";

import { useAuth } from "./use-auth";

export const useLogout = () => {
  const [isLoading, setIsLoading] = useState(false);
  const { logout: authLogout } = useAuth();
  const router = useRouter();

  const logout = async () => {
    try {
      setIsLoading(true);

      await authLogout();

      // Redirect to login page
      await router.navigate({
        to: "/login",
        search: {
          redirect: undefined,
        },
      });
    } catch {
      // Even if logout fails, clear local state and redirect
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
