import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { OrganizationInviteAcceptForm } from "@/components/organizations/organization-invite-accept-form";

export const Route = createFileRoute("/organizations/join")({
  validateSearch: z.object({
    organization: z.string().min(1).max(128),
    token: z.string().min(1).max(4096),
  }),
  component: OrganizationJoinPage,
});

function OrganizationJoinPage() {
  return <OrganizationInviteAcceptForm />;
}
