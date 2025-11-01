import { redirect } from "@tanstack/react-router";

import { hasValidSession } from "./session";

export function requireAuthBeforeLoad(ctx: { location: { href: string } }) {
  // Only run client-side to avoid SSR issues
  if (typeof window === "undefined") {
    return;
  }

  // Use a try-catch to handle any cookie/storage access issues
  try {
    if (!hasValidSession()) {
      throw redirect({
        to: "/login",
        search: {
          redirect: ctx.location.href,
        },
      });
    }
  } catch (error) {
    console.warn("Session check failed in route guard:", error);
  }
}

export function redirectIfAuthenticated() {
  if (hasValidSession()) {
    // If there's a redirect parameter, use it, otherwise go to dashboard
    const urlParams = new URLSearchParams(window.location.search);
    const redirectTo = urlParams.get("redirect");

    // Only redirect if the redirect target is not the login page itself
    if (redirectTo && !redirectTo.includes("/login")) {
      throw redirect({
        to: redirectTo,
      });
    } else {
      throw redirect({
        to: "/dashboard",
      });
    }
  }
}
