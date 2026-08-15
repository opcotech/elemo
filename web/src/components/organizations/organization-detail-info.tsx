import { Link } from "@tanstack/react-router";
import { Edit } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DetailCard } from "@/components/ui/detail-card";
import { DetailField } from "@/components/ui/detail-field";
import { ExternalLink } from "@/components/ui/external-link";
import type { Organization, Permission } from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";
import { zOrganizationStatus } from "@/lib/client/zod.gen";
import { formatDate } from "@/lib/format-date";

export function OrganizationDetailInfo({
  organization,
  permissions,
}: {
  organization: Organization;
  permissions: Permission[];
}) {
  const hasWritePermission = can(permissions, "write");

  return (
    <DetailCard
      data-section="organization-info"
      title="Organization Information"
      description="Details about the organization and its status."
      actions={
        hasWritePermission ? (
          <Button
            variant="outline"
            size="sm"
            render={
              <Link
                to="/settings/organizations/$organizationId/edit"
                params={{ organizationId: organization.id }}
              />
            }
          >
            <Edit className="size-4" />
            Edit
          </Button>
        ) : null
      }
    >
      <DetailField label="Name" value={organization.name} />

      <DetailField label="Email" value={organization.email} />

      <DetailField label="Website">
        {organization.website ? (
          <ExternalLink href={organization.website} />
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </DetailField>

      <DetailField label="Status">
        {organization.status === zOrganizationStatus.enum.active ? (
          <Badge variant="success">Active</Badge>
        ) : (
          <Badge variant="destructive">Deleted</Badge>
        )}
      </DetailField>

      <DetailField
        label="Created At"
        value={formatDate(organization.created_at)}
      />
    </DetailCard>
  );
}
