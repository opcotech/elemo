import { PageHeader } from "@/components/ui/page-header";
import { PageNotice } from "@/components/ui/page-notice";

export { DetailPageSkeleton as SettingsEntityDetailSkeleton } from "@/components/ui/detail-card";

export function SettingsEntityDetailError() {
  return (
    <PageNotice
      title="Details"
      message="Failed to load details. Please try again later."
    />
  );
}

export function SettingsAccessDenied({ resource }: { resource: string }) {
  return (
    <div className="space-y-6">
      <PageHeader title="Access Denied" />
      <div className="text-muted-foreground">
        You do not have permission to view this {resource}.
      </div>
    </div>
  );
}
