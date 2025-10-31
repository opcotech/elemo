import { OrganizationDetailHeader } from "./organization-detail-header";

import { Alert, AlertDescription } from "@/components/ui/alert";

export function OrganizationDetailError() {
  return (
    <div className="space-y-6">
      <OrganizationDetailHeader title="Organization Details" />

      <Alert variant="destructive">
        <AlertDescription>
          Failed to load organization details. Please try again later.
        </AlertDescription>
      </Alert>
    </div>
  );
}
