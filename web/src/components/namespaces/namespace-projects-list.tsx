import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Edit, Folder, Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import { ProjectDeleteDialog } from "@/components/projects/project-delete-dialog";
import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
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
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import {
  ResourceType,
  usePermissionsByResourceId,
} from "@/hooks/use-permissions";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1NamespacesProjectsGetOptions } from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
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
  organizationSlug: string;
  namespaceSlug: string;
  onDeleteClick: (project: PartialProject) => void;
}

function ProjectRow({
  project,
  permissions,
  isPermissionsLoading,
  organizationSlug,
  namespaceSlug,
  onDeleteClick,
}: ProjectRowProps) {
  const hasProjectReadPermission = can(permissions, Action.ProjectRead);
  const hasProjectWritePermission = can(permissions, Action.ProjectUpdate);
  const hasProjectDeletePermission = can(permissions, Action.ProjectDelete);

  return (
    <TableRow>
      <TableCell className="font-mono text-sm">
        <ConditionalLink
          to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey"
          params={{
            organizationSlug,
            namespaceSlug,
            projectKey: project.key,
          }}
          condition={hasProjectReadPermission}
        >
          {project.key}
        </ConditionalLink>
      </TableCell>
      <TableCell className="font-medium">
        <ConditionalLink
          to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey"
          params={{
            organizationSlug,
            namespaceSlug,
            projectKey: project.key,
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
                    to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey/edit"
                    params={{
                      organizationSlug,
                      namespaceSlug,
                      projectKey: project.key,
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
  organizationId: string;
  organizationSlug: string;
  namespaceId: string;
  namespaceSlug: string;
  namespacePermissions: EffectiveActions | undefined;
}

export function NamespaceProjectsList({
  organizationId,
  organizationSlug,
  namespaceId,
  namespaceSlug,
  namespacePermissions,
}: NamespaceProjectsListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedProject, setSelectedProject] = useState<PartialProject | null>(
    null
  );
  const pageNav = useCursorPageNav({ resetKey: searchTerm });
  const {
    data: projectsPage,
    isLoading,
    error,
  } = useQuery(
    v1NamespacesProjectsGetOptions({
      path: namespaceRefPath(organizationId, namespaceId),
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const projects = projectsPage?.items ?? [];
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
          to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/new"
          params={{ organizationSlug, namespaceSlug }}
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
                  organizationSlug={organizationSlug}
                  namespaceSlug={namespaceSlug}
                  onDeleteClick={handleDeleteClick}
                />
              );
            })}
          </TableBody>
        </Table>
        <CursorPaginator {...cursorPaginatorProps(projectsPage, pageNav)} />
      </SettingsResourceTable>

      {selectedProject && (
        <ProjectDeleteDialog
          project={selectedProject}
          organizationId={organizationId}
          organizationSlug={organizationSlug}
          namespaceId={namespaceId}
          namespaceSlug={namespaceSlug}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
