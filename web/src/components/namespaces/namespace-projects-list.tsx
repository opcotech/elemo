import { Link } from "@tanstack/react-router";
import { Edit, Folder, Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import { ProjectDeleteDialog } from "@/components/projects/project-delete-dialog";
import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConditionalLink } from "@/components/ui/conditional-link";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TableSkeleton } from "@/components/ui/table-skeleton";
import {
  ResourceType,
  usePermissionsByResourceId,
} from "@/hooks/use-permissions";
import type { EffectiveActions, PartialProject } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";

const namespaceProjectsListSkeletonColumns = [
  { header: "Key", skeletonClassName: "h-5 w-16" },
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Description", skeletonClassName: "h-4 w-48" },
  { header: "Status", skeletonClassName: "h-6 w-16" },
  {
    header: "Actions",
    skeletonClassName: "h-8 w-8",
    headerClassName: "text-right",
    cellClassName: "text-right",
    count: 2,
  },
] as const;

interface ProjectRowProps {
  project: PartialProject;
  permissions: EffectiveActions | undefined;
  isPermissionsLoading: boolean;
  organizationId: string;
  namespaceId: string;
  onDeleteClick: (project: PartialProject) => void;
}

function ProjectRow({
  project,
  permissions,
  isPermissionsLoading,
  organizationId,
  namespaceId,
  onDeleteClick,
}: ProjectRowProps) {
  const hasProjectReadPermission = can(permissions, Action.ProjectRead);
  const hasProjectWritePermission = can(permissions, Action.ProjectUpdate);
  const hasProjectDeletePermission = can(permissions, Action.ProjectDelete);

  return (
    <TableRow>
      <TableCell className="font-mono text-sm">
        <ConditionalLink
          to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId"
          params={{
            organizationId,
            namespaceId,
            projectId: project.id,
          }}
          condition={hasProjectReadPermission}
        >
          {project.key}
        </ConditionalLink>
      </TableCell>
      <TableCell className="font-medium">
        <ConditionalLink
          to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId"
          params={{
            organizationId,
            namespaceId,
            projectId: project.id,
          }}
          condition={hasProjectReadPermission}
        >
          {project.name}
        </ConditionalLink>
      </TableCell>
      <TableCell>
        <span className="text-muted-foreground text-sm">
          {project.description || "—"}
        </span>
      </TableCell>
      <TableCell>
        <Badge variant={project.status === "active" ? "success" : "secondary"}>
          {project.status}
        </Badge>
      </TableCell>
      <TableCell className="text-right">
        {isPermissionsLoading ? (
          <div className="flex justify-end gap-1">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        ) : (
          <div className="flex items-center justify-end gap-x-1">
            {hasProjectWritePermission && (
              <Button
                variant="ghost"
                size="sm"
                render={
                  <Link
                    to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId/edit"
                    params={{
                      organizationId,
                      namespaceId,
                      projectId: project.id,
                    }}
                  />
                }
              >
                <Edit className="size-4" />
                <span className="sr-only">Edit project</span>
              </Button>
            )}
            {hasProjectDeletePermission && (
              <Button
                variant="destructive-ghost"
                size="sm"
                onClick={() => onDeleteClick(project)}
              >
                <Trash2 className="size-4" />
                <span className="sr-only">Delete project</span>
              </Button>
            )}
          </div>
        )}
      </TableCell>
    </TableRow>
  );
}

interface NamespaceProjectsListProps {
  projects: PartialProject[];
  isLoading: boolean;
  error: unknown;
  organizationId: string;
  namespaceId: string;
  namespacePermissions: EffectiveActions | undefined;
}

export function NamespaceProjectsList({
  projects,
  isLoading,
  error,
  organizationId,
  namespaceId,
  namespacePermissions,
}: NamespaceProjectsListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedProject, setSelectedProject] = useState<PartialProject | null>(
    null
  );
  const hasCreatePermission = can(namespacePermissions, Action.ProjectCreate);
  const projectPermissionsById = usePermissionsByResourceId(
    ResourceType.Project,
    projects.map((project) => project.id)
  );

  const filteredProjects = useMemo(() => {
    if (!searchTerm.trim()) return projects;
    const term = searchTerm.toLowerCase();
    return projects.filter(
      (project) =>
        project.name.toLowerCase().includes(term) ||
        project.key.toLowerCase().includes(term) ||
        (project.description &&
          project.description.toLowerCase().includes(term))
    );
  }, [projects, searchTerm]);

  const handleDeleteClick = (project: PartialProject) => {
    setSelectedProject(project);
    setDeleteDialogOpen(true);
  };

  const handleDeleteSuccess = () => {
    setSelectedProject(null);
  };

  const createButton = hasCreatePermission ? (
    <Button
      variant="outline"
      size="sm"
      render={
        <Link
          to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/new"
          params={{ organizationId, namespaceId }}
        />
      }
    >
      <Plus className="size-4" />
      Create Project
    </Button>
  ) : undefined;

  return (
    <>
      <SettingsResourceTable
        dataSection="namespace-projects"
        title="Projects"
        description="Projects in this namespace."
        isLoading={isLoading}
        error={error}
        actionButton={createButton}
        search={{
          value: searchTerm,
          onChange: setSearchTerm,
          placeholder: "Search projects...",
          itemCount: projects.length,
        }}
        empty={{
          icon: <Folder />,
          title: "No projects found",
          description: "Create a project to organize work in this namespace.",
          action: createButton,
          searchTitle: "No projects found",
          searchDescription:
            "No projects match your search criteria. Try adjusting your search.",
          hasItems: projects.length > 0,
          hasFilteredItems: filteredProjects.length > 0,
        }}
        skeleton={
          <TableSkeleton columns={namespaceProjectsListSkeletonColumns} />
        }
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Key</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredProjects.map((project) => {
              const permissionQuery = projectPermissionsById.get(project.id);
              return (
                <ProjectRow
                  key={project.id}
                  project={project}
                  permissions={permissionQuery?.data}
                  isPermissionsLoading={permissionQuery?.isLoading ?? true}
                  organizationId={organizationId}
                  namespaceId={namespaceId}
                  onDeleteClick={handleDeleteClick}
                />
              );
            })}
          </TableBody>
        </Table>
      </SettingsResourceTable>

      {selectedProject && (
        <ProjectDeleteDialog
          project={selectedProject}
          organizationId={organizationId}
          namespaceId={namespaceId}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
