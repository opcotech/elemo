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
        <h1 className="font-bold text-2xl">Profile & Account</h1>
        <p className="mt-2 text-muted-foreground">
          Manage your personal information and preferences.
        </p>
      </div>
    </div>
  );
}
