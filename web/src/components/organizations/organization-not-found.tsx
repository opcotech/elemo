import { OrganizationDetailHeader } from "./organization-detail-header";

import { Alert, AlertDescription } from "@/components/ui/alert";

export function OrganizationNotFound() {
  return (
    <div className="space-y-6">
      <OrganizationDetailHeader title="Organization Details" />

      <Alert variant="destructive">
        <AlertDescription>
          Organization not found. Please check the URL and try again.
        </AlertDescription>
      </Alert>
    </div>
  );
}
