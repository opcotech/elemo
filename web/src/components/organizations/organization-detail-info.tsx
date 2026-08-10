import { Link } from "@tanstack/react-router";
import { Edit } from "lucide-react";

import { DetailField } from "@/components/shared/detail-field";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
    <Card data-section="organization-info">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle>Organization Information</CardTitle>
            <CardDescription>
              Details about the organization and its status.
            </CardDescription>
          </div>
          {hasWritePermission && (
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
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
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
        </div>
      </CardContent>
    </Card>
  );
}
