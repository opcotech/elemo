import { describe, expect, it } from "vitest";

import { ApiError } from "@/lib/api/errors";
import {
  accessDeniedRouteData,
  isAccessDeniedRouteData,
  withRouteErrors,
} from "@/lib/route-errors";

describe("route error handling", () => {
  it("maps API not-found errors to the router not-found boundary", async () => {
    await expect(
      withRouteErrors(() => Promise.reject(new ApiError(404, "missing")))
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("redirects operational permission failures", async () => {
    await expect(
      withRouteErrors(
        () => Promise.reject(new ApiError(403, "forbidden")),
        "redirect"
      )
    ).rejects.toMatchObject({
      options: { to: "/permission-denied" },
    });
  });

  it("returns serializable access-denied data for settings detail routes", async () => {
    await expect(
      withRouteErrors(
        () => Promise.reject(new ApiError(403, "forbidden")),
        "data"
      )
    ).resolves.toBe(accessDeniedRouteData);

    expect(isAccessDeniedRouteData(accessDeniedRouteData)).toBe(true);
    expect(isAccessDeniedRouteData({ accessDenied: false })).toBe(false);
  });

  it("returns successful route data unchanged", async () => {
    const routeData = { organizationId: "organization-1" };

    await expect(
      withRouteErrors(() => Promise.resolve(routeData))
    ).resolves.toBe(routeData);
  });

  it("preserves permission errors when no mapping is requested", async () => {
    const error = new ApiError(403, "forbidden");

    await expect(withRouteErrors(() => Promise.reject(error))).rejects.toBe(
      error
    );
  });

  it("does not remap unrelated loader failures", async () => {
    const error = new Error("loader failed");

    await expect(
      withRouteErrors(() => Promise.reject(error), "data")
    ).rejects.toBe(error);
  });

  it("only recognizes the explicit access-denied marker", () => {
    expect(isAccessDeniedRouteData(null)).toBe(false);
    expect(isAccessDeniedRouteData({ accessDenied: "true" })).toBe(false);
    expect(isAccessDeniedRouteData({ accessDenied: true })).toBe(true);
  });
});
