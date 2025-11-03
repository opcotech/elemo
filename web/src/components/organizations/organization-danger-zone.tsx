import { Trash2 } from "lucide-react";
import { useState } from "react";

import { OrganizationDangerZoneSkeleton } from "./organization-danger-zone-skeleton";
import { OrganizationDeleteDialog } from "./organization-delete-dialog";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Organization } from "@/lib/api";
import { can } from "@/lib/auth/permissions";


interface OrganizationDangerZoneProps {
  organization: Organization;
}

export function OrganizationDangerZone({
  organization,
}: OrganizationDangerZoneProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const { data: permissions, isLoading: isPermissionsLoading } = usePermissions(
    withResourceType(ResourceType.Organization, organization.id)
  );

  const hasDeletePermission = can(permissions, "delete");

  // Only render if organization is active
  if (organization.status !== "active") {
    return null;
  }

  // Show skeleton while loading permissions
  if (isPermissionsLoading) {
    return <OrganizationDangerZoneSkeleton />;
  }

  // Only show button if user has delete permission
  if (!hasDeletePermission) {
    return null;
  }

  return (
    <>
      <Card className="border-destructive bg-transparent">
        <CardHeader>
          <CardTitle className="text-destructive">Danger Zone</CardTitle>
          <CardDescription>
            Irreversible actions for this organization
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">
              Deleting an organization will mark it as deleted and hide it from
              listings. This action cannot be undone.
            </p>
            <p className="text-sm font-medium">Consequences:</p>
            <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
              <li>All organization members will lose access</li>
              <li>Organization data will be hidden from search and listings</li>
              <li>The organization will be marked as deleted</li>
              <li>This action is permanent and cannot be reversed</li>
            </ul>
          </div>
          <div className="flex justify-end">
            <Button
              variant="destructive"
              onClick={() => setDeleteDialogOpen(true)}
            >
              <Trash2 className="size-4" />
              Delete Organization
            </Button>
          </div>
        </CardContent>
      </Card>

      <OrganizationDeleteDialog
        organization={organization}
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
      />
    </>
  );
}

