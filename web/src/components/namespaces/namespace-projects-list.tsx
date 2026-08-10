import { Link } from "@tanstack/react-router";
import { Edit, Folder, Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import { ProjectDeleteDialog } from "@/components/projects/project-delete-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConditionalLink } from "@/components/ui/conditional-link";
import { ListContainer } from "@/components/ui/list-container";
import { SearchInput } from "@/components/ui/search-input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { PartialProject, Permission } from "@/lib/api";
import { can } from "@/lib/auth/permissions";

function NamespaceProjectsListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Key</TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Description</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-5 w-16" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-5 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-48" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-6 w-16" />
            </TableCell>
            <TableCell className="text-right">
              <div className="flex justify-end gap-1">
                <Skeleton className="h-8 w-8" />
                <Skeleton className="h-8 w-8" />
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

interface ProjectRowProps {
  project: PartialProject;
  organizationId: string;
  namespaceId: string;
  onDeleteClick: (project: PartialProject) => void;
}

function ProjectRow({
  project,
  organizationId,
  namespaceId,
  onDeleteClick,
}: ProjectRowProps) {
  const { data: projectPermissions, isLoading: isProjectPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Project, project.id));
  const hasProjectReadPermission = can(projectPermissions, "read");
  const hasProjectWritePermission = can(projectPermissions, "write");
  const hasProjectDeletePermission = can(projectPermissions, "delete");

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
        {isProjectPermissionsLoading ? (
          <div className="flex justify-end gap-1">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        ) : (
          <div className="flex items-center justify-end gap-x-1">
            {hasProjectWritePermission && (
              <Button variant="ghost" size="sm" asChild>
                <Link
                  to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId/edit"
                  params={{
                    organizationId,
                    namespaceId,
                    projectId: project.id,
                  }}
                >
                  <Edit className="size-4" />
                  <span className="sr-only">Edit project</span>
                </Link>
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
  namespacePermissions: Permission[] | undefined;
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
  const hasCreatePermission = can(namespacePermissions, "write");

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

  const emptyState =
    projects.length === 0
      ? {
          icon: <Folder />,
          title: "No projects found",
          description: "Create a project to organize work in this namespace.",
          action: hasCreatePermission ? (
            <Button variant="outline" size="sm" asChild>
              <Link
                to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/new"
                params={{ organizationId, namespaceId }}
              >
                <Plus className="size-4" />
                Create Project
              </Link>
            </Button>
          ) : undefined,
        }
      : filteredProjects.length === 0 && searchTerm.trim()
        ? {
            icon: <Folder />,
            title: "No projects found",
            description:
              "No projects match your search criteria. Try adjusting your search.",
          }
        : undefined;

  const shouldShowSearch = projects.length > 0 || searchTerm.trim() !== "";
  const createButton = hasCreatePermission ? (
    <Button variant="outline" size="sm" asChild>
      <Link
        to="/settings/organizations/$organizationId/namespaces/$namespaceId/projects/new"
        params={{ organizationId, namespaceId }}
      >
        <Plus className="size-4" />
        Create Project
      </Link>
    </Button>
  ) : undefined;

  return (
    <>
      <ListContainer
        data-section="namespace-projects"
        title="Projects"
        description="Projects in this namespace."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        actionButton={createButton}
        searchInput={
          shouldShowSearch ? (
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Search projects..."
              disabled={isLoading}
            />
          ) : undefined
        }
      >
        {isLoading ? (
          <NamespaceProjectsListSkeleton />
        ) : (
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
              {filteredProjects.map((project) => (
                <ProjectRow
                  key={project.id}
                  project={project}
                  organizationId={organizationId}
                  namespaceId={namespaceId}
                  onDeleteClick={handleDeleteClick}
                />
              ))}
            </TableBody>
          </Table>
        )}
      </ListContainer>

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
