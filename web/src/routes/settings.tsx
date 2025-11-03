import { Outlet, createFileRoute } from "@tanstack/react-router";

import { AuthenticatedLayout } from "@/components/layout/authenticated-layout";
import { SettingsLayout } from "@/components/layout/settings-layout";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/settings")({
  beforeLoad: requireAuthBeforeLoad,
  component: SettingsRoot,
});

function SettingsRoot() {
  return (
    <AuthenticatedLayout>
      <SettingsLayout>
        <Outlet />
      </SettingsLayout>
    </AuthenticatedLayout>
  );
}
