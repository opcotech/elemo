import { createFileRoute } from "@tanstack/react-router";

import { OrganizationInviteAcceptForm } from "@/components/organizations/organization-invite-accept-form";

export const Route = createFileRoute("/organizations/join")({
  validateSearch: (search: Record<string, unknown>) => ({
    organization: (search.organization as string) || undefined,
    token: (search.token as string) || undefined,
  }),
  component: OrganizationJoinPage,
});

function OrganizationJoinPage() {
  return <OrganizationInviteAcceptForm />;
}
