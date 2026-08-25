import { describe, expect, it } from "vitest";

import { cacheControlForRequestPath } from "./http-cache";

describe("cacheControlForRequestPath", () => {
  it.each(["/", "/login", "/dashboard", "/_serverFn/abc"])(
    "disables shared caches for %s",
    (pathname) => {
      expect(cacheControlForRequestPath(pathname)).toBe("private, no-store");
    }
  );

  it.each(["/assets", "/assets/", "/assets/index-abc.js"])(
    "leaves hashed asset path %s to Nitro/nginx",
    (pathname) => {
      expect(cacheControlForRequestPath(pathname)).toBeNull();
    }
  );
});
