import { Outlet, createFileRoute } from "@tanstack/react-router";

import { SettingsLayout } from "@/components/layout/settings-layout";

export const Route = createFileRoute("/_authenticated/settings")({
  staticData: {
    breadcrumb: "Settings",
  },
  component: SettingsRoot,
});

function SettingsRoot() {
  return (
    <SettingsLayout>
      <Outlet />
    </SettingsLayout>
  );
}
