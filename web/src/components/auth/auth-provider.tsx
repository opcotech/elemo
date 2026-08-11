import { useQueryClient } from "@tanstack/react-query";
import { useReducer } from "react";
import type { ReactNode } from "react";

import { AuthContext } from "@/lib/auth/auth-context";
import type { SessionView } from "@/lib/auth/functions";
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
  | { type: "CLEAR_AUTH" };

const initialState: AuthState = {
  user: null,
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
    default:
      return state;
  }
}

interface AuthProviderProps {
  children: ReactNode;
  initialSession?: SessionView | null;
}

export function AuthProvider({
  children,
  initialSession = null,
}: AuthProviderProps) {
  const queryClient = useQueryClient();
  const [state, dispatch] = useReducer(
    authReducer,
    initialSession,
    (session) =>
      session
        ? {
            user: session.user,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          }
        : { ...initialState, isLoading: false }
  );

  const login = async (credentials: LoginCredentials): Promise<void> => {
    try {
      queryClient.clear();
      dispatch({ type: "SET_LOADING", payload: true });
      dispatch({ type: "SET_ERROR", payload: null });

      const { loginFn } = await import("@/lib/auth/functions");
      const session = await loginFn({ data: credentials });
      dispatch({ type: "SET_USER", payload: session.user });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Login failed";
      dispatch({ type: "SET_ERROR", payload: errorMessage });
      throw error;
    }
  };

  const logout = async (): Promise<void> => {
    const { logoutFn } = await import("@/lib/auth/functions");
    await logoutFn();
    queryClient.clear();
    dispatch({ type: "CLEAR_AUTH" });
  };

  const refreshToken = async (): Promise<void> => {
    try {
      const { refreshSessionFn } = await import("@/lib/auth/functions");
      const session = await refreshSessionFn();
      if (!session) {
        throw new Error("Session refresh failed");
      }
      dispatch({ type: "SET_USER", payload: session.user });
    } catch (error) {
      queryClient.clear();
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
