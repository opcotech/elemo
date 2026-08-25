import {
  createCsrfMiddleware,
  createMiddleware,
  createStart,
} from "@tanstack/react-start";

import { cacheControlForRequestPath } from "@/lib/http-cache";

const csrfMiddleware = createCsrfMiddleware({
  filter: (ctx) => ctx.handlerType === "serverFn",
});

const noStoreCacheMiddleware = createMiddleware().server(
  async ({ next, request }) => {
    const result = await next();
    const cacheControl = cacheControlForRequestPath(
      new URL(request.url).pathname
    );
    if (cacheControl) {
      result.response.headers.set("Cache-Control", cacheControl);
    }
    return result;
  }
);

export const startInstance = createStart(() => ({
  requestMiddleware: [noStoreCacheMiddleware, csrfMiddleware],
}));
