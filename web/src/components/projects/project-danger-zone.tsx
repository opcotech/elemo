import { Trash2 } from "lucide-react";
import { useState } from "react";

import { ProjectDeleteDialog } from "./project-delete-dialog";

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
import type { Project } from "@/lib/api";
import { can } from "@/lib/auth/permissions";

export function ProjectDangerZoneSkeleton() {
  return (
    <Card className="border-destructive bg-transparent">
      <CardHeader>
        <CardTitle className="text-destructive">Danger Zone</CardTitle>
        <CardDescription>Irreversible actions for this project</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
        <div className="flex justify-end">
          <Skeleton className="h-10 w-36" />
        </div>
      </CardContent>
    </Card>
  );
}

interface ProjectDangerZoneProps {
  project: Project;
  organizationId: string;
  namespaceId: string;
}

export function ProjectDangerZone({
  project,
  organizationId,
  namespaceId,
}: ProjectDangerZoneProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const { data: permissions, isLoading: isPermissionsLoading } = usePermissions(
    withResourceType(ResourceType.Project, project.id)
  );

  const hasDeletePermission = can(permissions, "delete");

  if (isPermissionsLoading) {
    return <ProjectDangerZoneSkeleton />;
  }

  if (!hasDeletePermission) {
    return null;
  }

  return (
    <>
      <Card
        data-section="project-danger-zone"
        className="border-destructive bg-transparent"
      >
        <CardHeader>
          <CardTitle className="text-destructive">Danger Zone</CardTitle>
          <CardDescription>
            Irreversible actions for this project
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <p className="text-muted-foreground text-sm">
              Deleting a project permanently removes it from this namespace.
              This action cannot be undone.
            </p>
            <p className="text-sm font-medium">Consequences:</p>
            <ul className="text-muted-foreground list-inside list-disc space-y-1 text-sm">
              <li>The project will be permanently deleted</li>
              <li>
                Documents and issues will remain but will no longer be
                associated with the project
              </li>
              <li>You will be redirected to the namespace details page</li>
            </ul>
          </div>
          <div className="flex justify-end">
            <Button
              variant="destructive"
              onClick={() => setDeleteDialogOpen(true)}
            >
              <Trash2 className="size-4" />
              Delete Project
            </Button>
          </div>
        </CardContent>
      </Card>

      <ProjectDeleteDialog
        project={project}
        organizationId={organizationId}
        namespaceId={namespaceId}
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        navigateOnSuccess
      />
    </>
  );
}
