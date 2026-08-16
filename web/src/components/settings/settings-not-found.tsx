import { PageNotice } from "@/components/ui/page-notice";

export function SettingsNotFound() {
  return (
    <PageNotice
      title="Not found"
      message="The requested resource was not found. Please check the URL and try again."
    />
  );
}
