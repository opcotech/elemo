import { describe, expect, it } from "vitest";

import { airGapEntitlementWarning } from "@/components/settings/airgap-entitlement";
import { ApiError } from "@/lib/api/errors";
import type { SystemEntitlements } from "@/lib/api/types";

function entitlements(
  mode: SystemEntitlements["deployment_mode"],
  state?: NonNullable<SystemEntitlements["airgap"]>["state"]
): SystemEntitlements {
  return {
    deployment_mode: mode,
    airgap:
      state === undefined
        ? undefined
        : {
            installation_id: "11111111-1111-1111-1111-111111111111",
            seats_active: 1,
            state,
          },
  };
}

describe("airGapEntitlementWarning", () => {
  it("hides the banner for normal self-hosted deployments", () => {
    expect(airGapEntitlementWarning(entitlements("self_hosted"))).toBeNull();
  });

  it("hides the banner for a valid License", () => {
    expect(
      airGapEntitlementWarning(entitlements("airgap", "valid"))
    ).toBeNull();
  });

  it("hides the banner when the caller is not an administrator", () => {
    expect(
      airGapEntitlementWarning(undefined, new ApiError(403, "denied"))
    ).toBeNull();
  });

  it("warns during grace, expiry, and invalid states", () => {
    expect(
      airGapEntitlementWarning(entitlements("airgap", "grace"))?.title
    ).toMatch(/grace period/i);
    expect(
      airGapEntitlementWarning(entitlements("airgap", "expired"))?.description
    ).toMatch(/read-only/i);
    expect(
      airGapEntitlementWarning(entitlements("airgap", "missing"))?.description
    ).toMatch(/read-only/i);
    expect(
      airGapEntitlementWarning(entitlements("airgap", "invalid"))?.description
    ).toMatch(/read-only/i);
    expect(
      airGapEntitlementWarning(entitlements("airgap", "not_yet_valid"))
        ?.description
    ).toMatch(/read-only/i);
  });
});
