import { Link } from "@tanstack/react-router";
import { Edit, Folder, Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import { NamespaceDeleteDialog } from "./namespace-delete-dialog";

import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import { Button } from "@/components/ui/button";
import { ConditionalLink } from "@/components/ui/conditional-link";
import { CountBadge } from "@/components/ui/count-badge";
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
import type { Namespace, Permission } from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";

const namespacesListSkeletonColumns = [
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Description", skeletonClassName: "h-4 w-48" },
  { header: "Projects", skeletonClassName: "h-6 w-16" },
  { header: "Documents", skeletonClassName: "h-6 w-16" },
  {
    header: "Actions",
    skeletonClassName: "h-8 w-8",
    headerClassName: "text-right",
    cellClassName: "text-right",
    count: 2,
  },
] as const;

interface NamespaceRowProps {
  namespace: Namespace;
  permissions: Permission[] | undefined;
  isPermissionsLoading: boolean;
  organizationId: string;
  onDeleteClick: (namespace: Namespace) => void;
}

function NamespaceRow({
  namespace,
  permissions,
  isPermissionsLoading,
  organizationId,
  onDeleteClick,
}: NamespaceRowProps) {
  const projectCount = namespace.project_count ?? 0;
  const documentCount = namespace.document_count ?? 0;

  const hasNamespaceReadPermission = can(permissions, "read");
  const hasNamespaceWritePermission = can(permissions, "write");
  const hasNamespaceDeletePermission = can(permissions, "delete");

  return (
    <TableRow>
      <TableCell className="font-medium">
        <ConditionalLink
          to="/settings/organizations/$organizationId/namespaces/$namespaceId"
          params={{ organizationId, namespaceId: namespace.id }}
          condition={hasNamespaceReadPermission}
        >
          {namespace.name}
        </ConditionalLink>
      </TableCell>
      <TableCell>
        <span className="text-muted-foreground text-sm">
          {namespace.description || "—"}
        </span>
      </TableCell>
      <TableCell>
        <CountBadge count={projectCount} singular="project" plural="projects" />
      </TableCell>
      <TableCell>
        <CountBadge
          count={documentCount}
          singular="document"
          plural="documents"
        />
      </TableCell>
      <TableCell className="text-right">
        {isPermissionsLoading ? (
          <div className="flex justify-end gap-1">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        ) : (
          <div className="flex items-center justify-end gap-x-1">
            {hasNamespaceWritePermission && (
              <Button
                variant="ghost"
                size="sm"
                render={
                  <Link
                    to="/settings/organizations/$organizationId/namespaces/$namespaceId/edit"
                    params={{
                      organizationId,
                      namespaceId: namespace.id,
                    }}
                  />
                }
              >
                <Edit className="size-4" />
                <span className="sr-only">Edit namespace</span>
              </Button>
            )}
            {hasNamespaceDeletePermission && (
              <Button
                variant="destructive-ghost"
                size="sm"
                onClick={() => onDeleteClick(namespace)}
              >
                <Trash2 className="size-4" />
                <span className="sr-only">Delete namespace</span>
              </Button>
            )}
          </div>
        )}
      </TableCell>
    </TableRow>
  );
}

interface NamespacesListProps {
  namespaces: Namespace[];
  isLoading: boolean;
  error: unknown;
  organizationId: string;
  organizationPermissions: Permission[];
}

export function NamespacesList({
  namespaces,
  isLoading,
  error,
  organizationId,
  organizationPermissions,
}: NamespacesListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedNamespace, setSelectedNamespace] = useState<Namespace | null>(
    null
  );

  const hasOrgWritePermission = can(organizationPermissions, "write");
  const hasCreatePermission = hasOrgWritePermission;
  const namespacePermissionsById = usePermissionsByResourceId(
    ResourceType.Namespace,
    namespaces.map((namespace) => namespace.id)
  );

  const handleDeleteClick = (namespace: Namespace) => {
    setSelectedNamespace(namespace);
    setDeleteDialogOpen(true);
  };

  const handleDeleteSuccess = () => {
    setDeleteDialogOpen(false);
    setSelectedNamespace(null);
  };

  const filteredNamespaces = useMemo(() => {
    const filtered = !searchTerm.trim()
      ? namespaces
      : (() => {
          const term = searchTerm.toLowerCase();
          return namespaces.filter(
            (namespace) =>
              namespace.name.toLowerCase().includes(term) ||
              (namespace.description &&
                namespace.description.toLowerCase().includes(term))
          );
        })();
    return [...filtered].sort((a, b) =>
      a.name.localeCompare(b.name, undefined, { sensitivity: "base" })
    );
  }, [namespaces, searchTerm]);

  const createButton = hasCreatePermission ? (
    <Button
      variant="outline"
      size="sm"
      render={
        <Link
          to="/settings/organizations/$organizationId/namespaces/new"
          params={{ organizationId }}
        />
      }
    >
      <Plus className="size-4" />
      Create Namespace
    </Button>
  ) : undefined;

  return (
    <>
      <SettingsResourceTable
        dataSection="organization-namespaces"
        title="Namespaces"
        description="Organization namespaces and their resources."
        isLoading={isLoading}
        error={error}
        actionButton={createButton}
        search={{
          value: searchTerm,
          onChange: setSearchTerm,
          placeholder: "Search namespaces...",
          itemCount: namespaces.length,
        }}
        empty={{
          icon: <Folder />,
          title: "No namespaces found",
          description:
            "Namespaces help organize projects and documents. Create a namespace to get started.",
          action: createButton ? (
            <Button
              variant="outline"
              size="sm"
              render={
                <Link
                  to="/settings/organizations/$organizationId/namespaces/new"
                  params={{ organizationId }}
                />
              }
            >
              <Plus className="size-4" />
              Create Namespace
            </Button>
          ) : undefined,
          searchTitle: "No namespaces found",
          searchDescription:
            "No namespaces match your search criteria. Try adjusting your search.",
          hasItems: namespaces.length > 0,
          hasFilteredItems: filteredNamespaces.length > 0,
        }}
        skeleton={<TableSkeleton columns={namespacesListSkeletonColumns} />}
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Projects</TableHead>
              <TableHead>Documents</TableHead>
              {hasOrgWritePermission && (
                <TableHead>
                  <span className="sr-only">Actions</span>
                </TableHead>
              )}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredNamespaces.map((namespace) => {
              const permissionQuery = namespacePermissionsById.get(
                namespace.id
              );
              return (
                <NamespaceRow
                  key={namespace.id}
                  namespace={namespace}
                  permissions={permissionQuery?.data}
                  isPermissionsLoading={permissionQuery?.isLoading ?? true}
                  organizationId={organizationId}
                  onDeleteClick={handleDeleteClick}
                />
              );
            })}
          </TableBody>
        </Table>
      </SettingsResourceTable>

      {selectedNamespace && (
        <NamespaceDeleteDialog
          namespace={selectedNamespace}
          organizationId={organizationId}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
