import { createMiddleware } from "@tanstack/react-start";

import { requireSessionTokens } from "./session.server";

export const authMiddleware = createMiddleware({ type: "function" }).server(
  async ({ next }) => {
    const session = await requireSessionTokens();
    return next({
      context: {
        accessToken: session.accessToken,
        refreshToken: session.refreshToken,
      },
    });
  }
);
