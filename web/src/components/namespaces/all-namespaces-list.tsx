import { useQueries, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Edit, Folder, Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import { NamespaceDeleteDialog } from "./namespace-delete-dialog";

import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { Button } from "@/components/ui/button";
import { ConditionalLink } from "@/components/ui/conditional-link";
import { CountBadge } from "@/components/ui/count-badge";
import { InternalLink } from "@/components/ui/internal-link";
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
  withResourceType,
} from "@/hooks/use-permissions";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1NamespacesGetOptions,
  v1PermissionResourceGetOptions,
} from "@/lib/api/query-options";
import type { EffectiveActions, Namespace } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";

interface NamespaceWithOrganization extends Namespace {
  organizationId: string;
  organizationSlug: string;
  organizationName: string;
}

const allNamespacesListSkeletonColumns = [
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Organization", skeletonClassName: "h-4 w-32" },
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

interface AllNamespaceRowProps {
  namespace: NamespaceWithOrganization;
  permissions: EffectiveActions | undefined;
  isPermissionsLoading: boolean;
  onDeleteClick: (namespace: NamespaceWithOrganization) => void;
}

function AllNamespaceRow({
  namespace,
  permissions,
  isPermissionsLoading,
  onDeleteClick,
}: AllNamespaceRowProps) {
  const projectCount = namespace.project_count ?? 0;
  const documentCount = namespace.document_count ?? 0;

  const hasNamespaceReadPermission = can(permissions, Action.NamespaceRead);
  const hasNamespaceWritePermission = can(permissions, Action.NamespaceUpdate);
  const hasNamespaceDeletePermission = can(permissions, Action.NamespaceDelete);

  return (
    <TableRow>
      <TableCell className="font-medium">
        <ConditionalLink
          to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug"
          params={{
            organizationSlug: namespace.organizationSlug,
            namespaceSlug: namespace.slug,
          }}
          condition={hasNamespaceReadPermission}
        >
          {namespace.name}
        </ConditionalLink>
      </TableCell>
      <TableCell>
        <Link
          to="/settings/organizations/$organizationSlug"
          params={{ organizationSlug: namespace.organizationSlug }}
          className="text-primary hover:underline"
        >
          {namespace.organizationName}
        </Link>
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
                    to="/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/edit"
                    params={{
                      organizationSlug: namespace.organizationSlug,
                      namespaceSlug: namespace.slug,
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

interface AllNamespacesListProps {
  organizations: { id: string; name: string }[];
}

export function AllNamespacesList({ organizations }: AllNamespacesListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedNamespace, setSelectedNamespace] =
    useState<NamespaceWithOrganization | null>(null);
  const pageNav = useCursorPageNav({ resetKey: searchTerm });
  const {
    data: namespacesPage,
    isLoading,
    error,
  } = useQuery(
    v1NamespacesGetOptions({
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const namespaces = useMemo(
    () =>
      (namespacesPage?.items ?? []).map((namespace) => ({
        ...namespace,
        organizationId: namespace.organization.id,
        organizationSlug: namespace.organization.slug,
        organizationName: namespace.organization.name,
      })),
    [namespacesPage?.items]
  );

  const permissionQueries = useQueries({
    queries:
      organizations.length > 0
        ? organizations.map((org) =>
            v1PermissionResourceGetOptions({
              path: {
                resourceId: withResourceType(ResourceType.Organization, org.id),
              },
            })
          )
        : [],
  });
  const namespacePermissionsById = usePermissionsByResourceId(
    ResourceType.Namespace,
    namespaces.map((namespace) => namespace.id)
  );

  const canCreateNamespace = useMemo(() => {
    return organizations.some((org, index) => {
      const permissions = permissionQueries[index]?.data;
      return can(permissions, Action.NamespaceCreate);
    });
  }, [organizations, permissionQueries]);

  const handleDeleteClick = (namespace: NamespaceWithOrganization) => {
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
              namespace.organizationName.toLowerCase().includes(term) ||
              (namespace.description &&
                namespace.description.toLowerCase().includes(term))
          );
        })();
    return [...filtered].sort((a, b) =>
      a.name.localeCompare(b.name, undefined, { sensitivity: "base" })
    );
  }, [namespaces, searchTerm]);

  const createButton = canCreateNamespace ? (
    <Button render={<InternalLink to="/settings/namespaces/new" />}>
      <Plus className="size-4" />
      Create Namespace
    </Button>
  ) : undefined;

  return (
    <>
      <SettingsResourceTable
        dataSection="all-namespaces"
        title="Namespaces"
        description="All namespaces you have access to across organizations."
        isLoading={isLoading}
        error={error}
        actionButton={createButton}
        search={{
          value: searchTerm,
          onChange: setSearchTerm,
          placeholder: "Search namespaces or organizations...",
          itemCount: namespaces.length,
        }}
        empty={{
          icon: <Folder />,
          title: "No namespaces found",
          description:
            "You don't have access to any namespaces yet. Namespaces help organize projects and documents within organizations.",
          searchTitle: "No namespaces found",
          searchDescription:
            "No namespaces match your search criteria. Try adjusting your search.",
          hasItems: namespaces.length > 0,
          hasFilteredItems: filteredNamespaces.length > 0,
        }}
        skeleton={<TableSkeleton columns={allNamespacesListSkeletonColumns} />}
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Organization</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Projects</TableHead>
              <TableHead>Documents</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredNamespaces.map((namespace) => {
              const permissionQuery = namespacePermissionsById.get(
                namespace.id
              );
              return (
                <AllNamespaceRow
                  key={namespace.id}
                  namespace={namespace}
                  permissions={permissionQuery?.data}
                  isPermissionsLoading={permissionQuery?.isLoading ?? true}
                  onDeleteClick={handleDeleteClick}
                />
              );
            })}
          </TableBody>
        </Table>
        <CursorPaginator {...cursorPaginatorProps(namespacesPage, pageNav)} />
      </SettingsResourceTable>

      {selectedNamespace && (
        <NamespaceDeleteDialog
          namespace={selectedNamespace}
          organizationId={selectedNamespace.organizationId}
          organizationSlug={selectedNamespace.organizationSlug}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
