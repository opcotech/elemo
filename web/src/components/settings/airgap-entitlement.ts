import { isPermissionDenied } from "@/lib/api/errors";
import type {
  SystemAirGapEntitlement,
  SystemEntitlements,
} from "@/lib/api/types";

const warningStates = new Set<SystemAirGapEntitlement["state"]>([
  "grace",
  "expired",
  "invalid",
  "missing",
  "not_yet_valid",
]);

export function airGapEntitlementWarning(
  data: SystemEntitlements | undefined,
  error?: unknown
): { title: string; description: string } | null {
  if (error && isPermissionDenied(error)) {
    return null;
  }
  if (data?.deployment_mode !== "airgap" || !data.airgap) {
    return null;
  }
  if (!warningStates.has(data.airgap.state)) {
    return null;
  }
  return warningCopy(data.airgap.state);
}

export function warningCopy(state: SystemAirGapEntitlement["state"]): {
  title: string;
  description: string;
} {
  switch (state) {
    case "grace":
      return {
        title: "License is in grace period",
        description:
          "This installation can still write data and activate human users until the grace period ends. Replace the license file and restart to restore a valid entitlement.",
      };
    case "expired":
      return {
        title: "License has expired",
        description:
          "This installation is read-only. Reads, login, password reset, and user deactivation remain available. Domain writes and new human user activations are blocked until a valid license is installed and the server is restarted.",
      };
    case "not_yet_valid":
      return {
        title: "License is not yet valid",
        description:
          "This installation is read-only. Reads, login, password reset, and user deactivation remain available. Domain writes and new human user activations are blocked until the license not-before time. Check the system clock or replace the license file.",
      };
    case "missing":
      return {
        title: "License is missing",
        description:
          "This installation is read-only. Reads, login, password reset, and user deactivation remain available. Domain writes and new human user activations are blocked. Mount a vendor-signed license file and restart this AirGap node.",
      };
    default:
      return {
        title: "License is invalid",
        description:
          "This installation is read-only. Reads, login, password reset, and user deactivation remain available. Domain writes and new human user activations are blocked. Replace the license file with a vendor-signed license bound to this installation and restart.",
      };
  }
}
