import { Trash2 } from "lucide-react";
import { useState } from "react";

import { NamespaceDeleteDialog } from "./namespace-delete-dialog";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Namespace } from "@/lib/api";
import { can } from "@/lib/auth/permissions";

export function NamespaceDangerZoneSkeleton() {
  return (
    <Card className="border-destructive bg-transparent">
      <CardHeader>
        <CardTitle className="text-destructive">Danger Zone</CardTitle>
        <CardDescription>
          Irreversible actions for this namespace
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
        <div className="flex justify-end">
          <Skeleton className="h-10 w-40" />
        </div>
      </CardContent>
    </Card>
  );
}

interface NamespaceDangerZoneProps {
  namespace: Namespace;
  organizationId: string;
}

export function NamespaceDangerZone({
  namespace,
  organizationId,
}: NamespaceDangerZoneProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const { data: permissions, isLoading: isPermissionsLoading } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );

  const hasDeletePermission = can(permissions, "delete");

  if (isPermissionsLoading) {
    return <NamespaceDangerZoneSkeleton />;
  }

  if (!hasDeletePermission) {
    return null;
  }

  return (
    <>
      <Card
        data-section="namespace-danger-zone"
        className="border-destructive bg-transparent"
      >
        <CardHeader>
          <CardTitle className="text-destructive">Danger Zone</CardTitle>
          <CardDescription>
            Irreversible actions for this namespace
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <p className="text-muted-foreground text-sm">
              Deleting a namespace permanently removes it from this
              organization. This action cannot be undone.
            </p>
            <p className="text-sm font-medium">Consequences:</p>
            <ul className="text-muted-foreground list-inside list-disc space-y-1 text-sm">
              <li>The namespace will be permanently deleted</li>
              <li>
                Projects and documents will remain but will no longer be
                associated with the namespace
              </li>
              <li>You will be redirected to the organization details page</li>
            </ul>
          </div>
          <div className="flex justify-end">
            <Button
              variant="destructive"
              onClick={() => setDeleteDialogOpen(true)}
            >
              <Trash2 className="size-4" />
              Delete Namespace
            </Button>
          </div>
        </CardContent>
      </Card>

      <NamespaceDeleteDialog
        namespace={namespace}
        organizationId={organizationId}
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        navigateOnSuccess
      />
    </>
  );
}
