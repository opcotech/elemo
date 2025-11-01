import { Link } from "@tanstack/react-router";
import { Edit } from "lucide-react";

import { OrganizationDetailField } from "./organization-detail-field";

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
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Organization } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { formatDate } from "@/lib/utils";

export function OrganizationDetailInfo({
  organization,
}: {
  organization: Organization;
}) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Organization, organization.id)
  );

  const hasWritePermission = can(permissions, "write");

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle>Organization Information</CardTitle>
            <CardDescription>
              Details about the organization and its status.
            </CardDescription>
          </div>
          {hasWritePermission && (
            <Button variant="outline" size="sm" asChild>
              <Link
                to="/settings/organizations/$organizationId/edit"
                params={{ organizationId: organization.id }}
              >
                <Edit className="mr-2 size-4" />
                Edit
              </Link>
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <OrganizationDetailField label="Name" value={organization.name} />

          <OrganizationDetailField label="Email" value={organization.email} />

          <OrganizationDetailField label="Website">
            {organization.website ? (
              <ExternalLink href={organization.website} />
            ) : (
              <span className="text-muted-foreground">—</span>
            )}
          </OrganizationDetailField>

          <OrganizationDetailField label="Status">
            {organization.status === "active" ? (
              <Badge variant="success">Active</Badge>
            ) : (
              <Badge variant="destructive">Deleted</Badge>
            )}
          </OrganizationDetailField>

          <OrganizationDetailField
            label="Created At"
            value={formatDate(organization.created_at)}
          />
        </div>
      </CardContent>
    </Card>
  );
}
