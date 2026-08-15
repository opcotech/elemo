import { PageNotice } from "@/components/ui/page-notice";

export function OrganizationNotFound() {
  return (
    <PageNotice
      title="Organization Details"
      message="Organization not found. Please check the URL and try again."
    />
  );
}
