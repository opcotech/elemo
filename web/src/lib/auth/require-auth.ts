import { redirect } from "@tanstack/react-router";

import { hasValidSession } from "./session";

export function requireAuthBeforeLoad(ctx: { location: { href: string } }) {
  // Only run client-side to avoid SSR issues
  if (typeof window === "undefined") {
    return;
  }

  // Use a try-catch to handle any cookie/storage access issues
  try {
    const hasSession = hasValidSession();
    console.log("🔐 Route guard check:", {
      url: ctx.location.href,
      hasSession,
      cookies: typeof document !== "undefined" ? document.cookie : "N/A",
      timestamp: new Date().toISOString(),
    });

    if (!hasSession) {
      console.log("❌ No session found, redirecting to login");
      throw redirect({
        to: "/login",
        search: {
          redirect: ctx.location.href,
        },
      });
    }

    console.log("✅ Session valid, allowing access");
  } catch (error) {
    // If session check fails, allow through and let AuthProvider handle it
    console.warn("Session check failed in route guard:", error);
  }
}

export function redirectIfAuthenticated() {
  if (hasValidSession()) {
    throw redirect({
      to: "/dashboard",
    });
  }
}
