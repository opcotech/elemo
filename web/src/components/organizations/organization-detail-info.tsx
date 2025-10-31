import { OrganizationDetailField } from "./organization-detail-field";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { Organization } from "@/lib/api";
import { formatDate } from "@/lib/utils";

export function OrganizationDetailInfo({
  organization,
}: {
  organization: Organization;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Organization Information</CardTitle>
        <CardDescription>
          Details about the organization and its status.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <OrganizationDetailField label="Name" value={organization.name} />

          <OrganizationDetailField label="Email" value={organization.email} />

          <OrganizationDetailField label="Website">
            {organization.website ? (
              <a
                href={organization.website}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline"
              >
                {organization.website}
              </a>
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
