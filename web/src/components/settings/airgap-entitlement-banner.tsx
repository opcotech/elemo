import { useQuery } from "@tanstack/react-query";
import { AlertTriangle } from "lucide-react";

import { airGapEntitlementWarning } from "@/components/settings/airgap-entitlement";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { v1SystemEntitlementsOptions } from "@/lib/api/query-options";

export function AirGapEntitlementBanner() {
  const { data, error } = useQuery({
    ...v1SystemEntitlementsOptions(),
    throwOnError: false,
  });

  const copy = airGapEntitlementWarning(data, error);
  if (!copy) {
    return null;
  }

  return (
    <Alert variant="warning" className="mb-6">
      <AlertTriangle aria-hidden="true" />
      <AlertTitle role="heading" aria-level={2}>
        {copy.title}
      </AlertTitle>
      <AlertDescription>{copy.description}</AlertDescription>
    </Alert>
  );
}
