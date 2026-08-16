import { createFileRoute } from "@tanstack/react-router";

import { WorkSurface } from "@/components/work/work-surface";
import { useAuth } from "@/hooks/use-auth";
import { workRouteSearchSchema } from "@/lib/work-route-search";

export const Route = createFileRoute("/_authenticated/my-work")({
  staticData: { breadcrumb: "My Work" },
  validateSearch: workRouteSearchSchema,
  component: MyWorkRoute,
});

function MyWorkRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { user } = useAuth();

  if (!user) {
    return null;
  }

  return (
    <WorkSurface
      title="My Work"
      description="What you own, what is blocked, and what is coming up across Elemo."
      scope={{ type: "person", personId: user.id }}
      search={search}
      onSearchChange={(patch) =>
        void navigate({
          search: (previous) => ({ ...previous, ...patch }),
          replace: true,
        })
      }
    />
  );
}
