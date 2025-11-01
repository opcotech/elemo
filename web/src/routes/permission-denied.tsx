import { createFileRoute } from "@tanstack/react-router";

import { PermissionDenied } from "@/components/permission-denied";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/permission-denied")({
  beforeLoad: requireAuthBeforeLoad,
  component: PermissionDeniedPage,
});

function PermissionDeniedPage() {
  return <PermissionDenied />;
}
