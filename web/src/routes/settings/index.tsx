import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/settings/")({
  beforeLoad: requireAuthBeforeLoad,
  component: ProfileSettings,
});

function ProfileSettings() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Profile & Account",
        isNavigatable: false,
      },
    ]);
  }, []);

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Profile & Account</h1>
        <p className="mt-2 text-gray-600">
          Manage your personal information and preferences.
        </p>
      </div>
    </div>
  );
}
