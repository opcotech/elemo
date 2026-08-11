import { createFileRoute } from "@tanstack/react-router";

import { PermissionDenied } from "@/components/shared/permission-denied";

export const Route = createFileRoute("/_authenticated/permission-denied")({
  staticData: {
    breadcrumb: "Permission denied",
  },
  component: PermissionDeniedPage,
});

function PermissionDeniedPage() {
  return <PermissionDenied />;
}
