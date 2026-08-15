import { PageNotice } from "@/components/ui/page-notice";

export function OrganizationDetailError() {
  return (
    <PageNotice
      title="Organization Details"
      message="Failed to load organization details. Please try again later."
    />
  );
}
