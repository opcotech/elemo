import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { WorkSurface } from "@/components/work";
import { useAuth } from "@/hooks/use-auth";
import { resolveDemoPerson } from "@/lib/mock-data";
import {
  workLayoutSchema,
  workRouteSearchSchema,
} from "@/lib/work-route-search";

const myWorkSearchSchema = workRouteSearchSchema.extend({
  layout: workLayoutSchema.catch("list"),
  group: z.enum(["status", "priority", "assignee", "none"]).catch("status"),
});

export const Route = createFileRoute("/_authenticated/my-work")({
  staticData: { breadcrumb: "My Work" },
  validateSearch: myWorkSearchSchema,
  component: MyWorkRoute,
});

function MyWorkRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { user } = useAuth();
  const demoPerson = resolveDemoPerson(user);

  return (
    <WorkSurface
      title="My Work"
      description="What you own, what is blocked, and what is coming up across Elemo."
      scope={{ type: "person", personId: demoPerson.id }}
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
