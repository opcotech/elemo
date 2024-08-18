import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/use-auth";
import {
  getAccessToken,
  getRefreshToken,
  getTokenTimeRemaining,
  getUser,
  hasValidSession,
  isAccessTokenExpired,
  manualAuthCleanup,
} from "@/lib/auth/session";

interface DebugState {
  hasAccessToken: boolean;
  hasRefreshToken: boolean;
  hasUser: boolean;
  hasValidSession: boolean;
  isTokenExpired: boolean;
  timeRemaining: number;
  user: any;
  authState: any;
}

export function AuthDebug() {
  const auth = useAuth();
  const [debugState, setDebugState] = useState<DebugState | null>(null);
  const [loading, setLoading] = useState(true);

  // Log startup state for debugging
  useEffect(() => {
    const logStartupState = () => {
      console.log("🔍 AuthDebug component mounted");
      console.log("📊 Initial localStorage state:", {
        keys: Object.keys(localStorage).filter(
          (key) =>
            key.includes("elemo") ||
            key.includes("auth") ||
            key.includes("token")
        ),
        authState: {
          isAuthenticated: auth.isAuthenticated,
          isLoading: auth.isLoading,
          hasUser: !!auth.user,
        },
      });
    };

    logStartupState();
  }, []);

  const loadDebugState = async () => {
    try {
      setLoading(true);
      const [
        accessToken,
        refreshToken,
        user,
        validSession,
        expired,
        timeRemaining,
      ] = await Promise.all([
        getAccessToken(),
        getRefreshToken(),
        getUser(),
        hasValidSession(),
        isAccessTokenExpired(),
        getTokenTimeRemaining(),
      ]);

      setDebugState({
        hasAccessToken: !!accessToken,
        hasRefreshToken: !!refreshToken,
        hasUser: !!user,
        hasValidSession: validSession,
        isTokenExpired: expired,
        timeRemaining,
        user: user ? { id: user.id, email: user.email } : null,
        authState: {
          isAuthenticated: auth.isAuthenticated,
          isLoading: auth.isLoading,
          error: auth.error,
        },
      });
    } catch (error) {
      console.error("Failed to load debug state:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDebugState();
  }, [auth.isAuthenticated, auth.isLoading]);

  const formatTime = (ms: number): string => {
    if (ms <= 0) return "Expired";
    const minutes = Math.floor(ms / 60000);
    const seconds = Math.floor((ms % 60000) / 1000);
    return `${minutes}m ${seconds}s`;
  };

  const handleCleanup = () => {
    if (
      window.confirm(
        "This will clear ALL authentication data and refresh the page. Continue?"
      )
    ) {
      manualAuthCleanup();
      window.location.reload();
    }
  };

  const handleDebugConsole = async () => {
    console.group("🔐 Manual Auth Debug");
    try {
      const [
        accessToken,
        refreshToken,
        user,
        hasSession,
        isExpired,
        timeRemaining,
      ] = await Promise.all([
        getAccessToken(),
        getRefreshToken(),
        getUser(),
        hasValidSession(),
        isAccessTokenExpired(),
        getTokenTimeRemaining(),
      ]);

      console.log("Storage State:");
      console.log("- Access Token:", accessToken ? "Present" : "Missing");
      console.log("- Refresh Token:", refreshToken ? "Present" : "Missing");
      console.log("- User Data:", user ? "Present" : "Missing");
      console.log("- Has Valid Session:", hasSession);
      console.log("- Is Token Expired:", isExpired);
      console.log(
        "- Time Remaining:",
        timeRemaining > 0 ? `${Math.floor(timeRemaining / 60000)}m` : "Expired"
      );

      if (user) {
        console.log("User:", { id: user.id, email: user.email });
      }

      console.log(
        "localStorage keys:",
        Object.keys(localStorage).filter(
          (key) =>
            key.includes("elemo") ||
            key.includes("auth") ||
            key.includes("token")
        )
      );
    } catch (error) {
      console.error("Debug failed:", error);
    }
    console.groupEnd();
  };

  if (loading) {
    return (
      <div className="rounded border border-yellow-200 bg-yellow-50 p-4">
        <p>Loading debug state...</p>
      </div>
    );
  }

  return (
    <div className="space-y-4 rounded border border-gray-200 bg-gray-50 p-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">🔐 Auth Debug State</h3>
        <div className="space-x-2">
          <Button onClick={loadDebugState}>Refresh</Button>
          <Button onClick={handleCleanup} variant="destructive">
            Clean All
          </Button>
          <Button onClick={handleDebugConsole} variant="secondary">
            Console Log
          </Button>
          <Button
            onClick={() => {
              console.log("🔍 Immediate Storage Check:");
              console.log("localStorage keys:", Object.keys(localStorage));
              console.log("cookies:", document.cookie);
              console.log("Auth context:", {
                isAuthenticated: auth.isAuthenticated,
                isLoading: auth.isLoading,
                user: auth.user
                  ? { id: auth.user.id, email: auth.user.email }
                  : null,
              });
            }}
            variant="success"
          >
            Quick Check
          </Button>
        </div>
      </div>

      {debugState && (
        <div className="space-y-2 text-sm">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h4 className="mb-2 font-medium">Storage State:</h4>
              <div className="space-y-1">
                <div>
                  Access Token:{" "}
                  <span
                    className={
                      debugState.hasAccessToken
                        ? "text-green-600"
                        : "text-red-600"
                    }
                  >
                    {debugState.hasAccessToken ? "✅ Present" : "❌ Missing"}
                  </span>
                </div>
                <div>
                  Refresh Token:{" "}
                  <span
                    className={
                      debugState.hasRefreshToken
                        ? "text-green-600"
                        : "text-red-600"
                    }
                  >
                    {debugState.hasRefreshToken ? "✅ Present" : "❌ Missing"}
                  </span>
                </div>
                <div>
                  User Data:{" "}
                  <span
                    className={
                      debugState.hasUser ? "text-green-600" : "text-red-600"
                    }
                  >
                    {debugState.hasUser ? "✅ Present" : "❌ Missing"}
                  </span>
                </div>
                <div>
                  Valid Session:{" "}
                  <span
                    className={
                      debugState.hasValidSession
                        ? "text-green-600"
                        : "text-red-600"
                    }
                  >
                    {debugState.hasValidSession ? "✅ Yes" : "❌ No"}
                  </span>
                </div>
              </div>
            </div>

            <div>
              <h4 className="mb-2 font-medium">Token State:</h4>
              <div className="space-y-1">
                <div>
                  Token Expired:{" "}
                  <span
                    className={
                      debugState.isTokenExpired
                        ? "text-red-600"
                        : "text-green-600"
                    }
                  >
                    {debugState.isTokenExpired ? "❌ Yes" : "✅ No"}
                  </span>
                </div>
                <div>
                  Time Remaining:{" "}
                  <span className="font-mono">
                    {formatTime(debugState.timeRemaining)}
                  </span>
                </div>
                <div>
                  Auth State:{" "}
                  <span
                    className={
                      debugState.authState.isAuthenticated
                        ? "text-green-600"
                        : "text-red-600"
                    }
                  >
                    {debugState.authState.isAuthenticated
                      ? "✅ Authenticated"
                      : "❌ Not Authenticated"}
                  </span>
                </div>
                <div>
                  Loading:{" "}
                  <span>
                    {debugState.authState.isLoading ? "⏳ Yes" : "❌ No"}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {debugState.user && (
            <div>
              <h4 className="mb-1 font-medium">User Info:</h4>
              <div className="rounded border bg-white p-2">
                <pre className="text-xs">
                  {JSON.stringify(debugState.user, null, 2)}
                </pre>
              </div>
            </div>
          )}

          {debugState.authState.error && (
            <div>
              <h4 className="mb-1 font-medium text-red-600">Error:</h4>
              <div className="rounded border border-red-200 bg-red-50 p-2 text-red-700">
                {debugState.authState.error}
              </div>
            </div>
          )}

          <div className="border-t pt-2 text-xs text-gray-500">
            <p>
              localStorage keys:{" "}
              {Object.keys(localStorage)
                .filter(
                  (key) =>
                    key.includes("elemo") ||
                    key.includes("auth") ||
                    key.includes("token")
                )
                .join(", ") || "none"}
            </p>
            <p>Last updated: {new Date().toLocaleTimeString()}</p>
          </div>
        </div>
      )}
    </div>
  );
}
