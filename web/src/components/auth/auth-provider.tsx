import { useNavigate } from "@tanstack/react-router";
import { useEffect, useReducer } from "react";
import type { ReactNode } from "react";

import { client, v1UsersGet } from "@/lib/api";
import { authClient } from "@/lib/auth/auth-client";
import { AuthContext } from "@/lib/auth/auth-context";
import {
  clearAllAuthData,
  getAccessToken,
  getRefreshToken,
  getUser,
  storeTokens,
  storeUser,
} from "@/lib/auth/session";
import { tokenRefreshService } from "@/lib/auth/token-refresh-service";
import type {
  AuthContextType,
  AuthState,
  LoginCredentials,
  User,
} from "@/lib/auth/types";

type AuthAction =
  | { type: "SET_LOADING"; payload: boolean }
  | { type: "SET_USER"; payload: User }
  | { type: "SET_ERROR"; payload: string | null }
  | { type: "CLEAR_AUTH" }
  | { type: "SET_AUTHENTICATED"; payload: { user: User; tokens: any } };

const initialState: AuthState = {
  user: null,
  tokens: null,
  isAuthenticated: false,
  isLoading: true,
  error: null,
};

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case "SET_LOADING":
      return { ...state, isLoading: action.payload };
    case "SET_USER":
      return {
        ...state,
        user: action.payload,
        isAuthenticated: true,
        isLoading: false,
      };
    case "SET_ERROR":
      return { ...state, error: action.payload, isLoading: false };
    case "CLEAR_AUTH":
      return { ...initialState, isLoading: false };
    case "SET_AUTHENTICATED":
      return {
        ...state,
        user: action.payload.user,
        tokens: action.payload.tokens,
        isAuthenticated: true,
        isLoading: false,
        error: null,
      };
    default:
      return state;
  }
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const navigate = useNavigate();
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Initialize auth state on mount
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        console.log("🔐 Starting auth initialization...");
        dispatch({ type: "SET_LOADING", payload: true });

        const refreshToken = getRefreshToken();
        const storedUser = await getUser();
        const accessToken = await getAccessToken();

        console.log("🔍 Auth state check:", {
          hasRefreshToken: !!refreshToken,
          hasStoredUser: !!storedUser,
          hasAccessToken: !!accessToken,
        });

        if (!refreshToken || !accessToken || !storedUser) {
          throw new Error(
            "No refresh token, access token, or stored user found"
          );
        }

        // If no refresh token, we have no valid session
        if (!refreshToken) {
          console.log("❌ No refresh token found, clearing session");
          clearAllAuthData();
          dispatch({ type: "CLEAR_AUTH" });
          return;
        }

        // If we have both access token and user data, validate token
        if (accessToken && storedUser) {
          console.log("🔑 Validating existing access token...");
          const isValid = await authClient.validateToken(accessToken);
          if (isValid) {
            console.log("✅ Access token valid, restoring session");
            dispatch({ type: "SET_USER", payload: storedUser });
            return;
          } else {
            console.log("❌ Access token invalid, will refresh");
          }
        }

        // If access token is invalid/missing/expired, try to refresh using refresh token
        console.log("🔄 Refreshing tokens...");
        try {
          const newTokens = await authClient.refreshToken(refreshToken);
          await storeTokens(newTokens);
          console.log("✅ Tokens refreshed successfully");

          // Always fetch fresh user data after token refresh to ensure it's current
          console.log("👤 Fetching user data...");
          const usersResponse = await v1UsersGet({
            client,
            auth: () => newTokens.access_token,
          });

          if (
            usersResponse.data &&
            Array.isArray(usersResponse.data) &&
            usersResponse.data.length > 0
          ) {
            const user: User = usersResponse.data[0];
            await storeUser(user);
            console.log("✅ User data updated:", {
              id: user.id,
              email: user.email,
            });
            dispatch({ type: "SET_USER", payload: user });
            return;
          } else {
            throw new Error("Failed to fetch user data after token refresh");
          }
        } catch (refreshError) {
          console.error("❌ Token refresh failed:", refreshError);
          // Clear invalid session if refresh fails
          clearAllAuthData();
          dispatch({ type: "CLEAR_AUTH" });
        }
      } catch (error) {
        console.error("❌ Auth initialization error:", error);
        clearAllAuthData();
        dispatch({ type: "CLEAR_AUTH" });

        navigate({
          to: "/login",
          search: {
            redirect: window.location.href,
          },
        });
      }
    };

    // Add a small delay to ensure localStorage is ready
    setTimeout(initializeAuth, 50);
  }, []);

  // Set up automatic token refresh service
  useEffect(() => {
    if (state.isAuthenticated) {
      tokenRefreshService.startAutoRefresh();

      // Listen for token refresh events
      const handleTokenRefreshed = () => {
        // Token was refreshed successfully in background
        console.debug("Token refreshed in background");
      };

      const handleTokenRefreshFailed = () => {
        // Token refresh failed, logout user
        dispatch({ type: "CLEAR_AUTH" });
      };

      window.addEventListener("tokenRefreshed", handleTokenRefreshed);
      window.addEventListener("tokenRefreshFailed", handleTokenRefreshFailed);

      return () => {
        window.removeEventListener("tokenRefreshed", handleTokenRefreshed);
        window.removeEventListener(
          "tokenRefreshFailed",
          handleTokenRefreshFailed
        );
      };
    } else {
      tokenRefreshService.stopAutoRefresh();
    }
  }, [state.isAuthenticated]);

  const login = async (credentials: LoginCredentials): Promise<void> => {
    try {
      dispatch({ type: "SET_LOADING", payload: true });
      dispatch({ type: "SET_ERROR", payload: null });

      const tokens = await authClient.login(credentials);
      await storeTokens(tokens);

      // Fetch user data using generated API client
      const usersResponse = await v1UsersGet({
        client,
        auth: () => tokens.access_token,
      });

      if (
        !usersResponse.data ||
        !Array.isArray(usersResponse.data) ||
        usersResponse.data.length === 0
      ) {
        throw new Error("Failed to fetch user data");
      }

      const user: User = usersResponse.data[0];

      await storeUser(user);
      dispatch({ type: "SET_AUTHENTICATED", payload: { user, tokens } });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Login failed";
      dispatch({ type: "SET_ERROR", payload: errorMessage });
      throw error;
    }
  };

  // eslint-disable-next-line @typescript-eslint/require-await
  const logout = async (): Promise<void> => {
    tokenRefreshService.stopAutoRefresh();
    clearAllAuthData();
    dispatch({ type: "CLEAR_AUTH" });
  };

  const refreshToken = async (): Promise<void> => {
    try {
      await tokenRefreshService.forceRefresh();
    } catch (error) {
      dispatch({ type: "CLEAR_AUTH" });
      throw error;
    }
  };

  const clearError = (): void => {
    dispatch({ type: "SET_ERROR", payload: null });
  };

  const contextValue: AuthContextType = {
    ...state,
    login,
    logout,
    refreshToken,
    clearError,
  };

  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
}
