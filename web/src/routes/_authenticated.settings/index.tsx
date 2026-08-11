import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/settings/")({
  staticData: {
    breadcrumb: "Profile & account",
  },
  component: ProfileSettings,
});

function ProfileSettings() {
  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Profile & Account</h1>
        <p className="text-muted-foreground mt-2">
          Manage your personal information and preferences.
        </p>
      </div>
    </div>
  );
}
