import { useQueries, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Edit, Folder, Plus, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";

import { NamespaceDeleteDialog } from "./namespace-delete-dialog";

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
import type { Namespace } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import {
  v1OrganizationsGetOptions,
  v1PermissionResourceGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { pluralize } from "@/lib/utils";

interface NamespaceWithOrganization extends Namespace {
  organizationId: string;
  organizationName: string;
}

function AllNamespacesListSkeleton() {
  return (
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
        {Array.from({ length: 5 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-5 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-48" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-6 w-16" />
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

interface AllNamespaceRowProps {
  namespace: NamespaceWithOrganization;
  onDeleteClick: (namespace: NamespaceWithOrganization) => void;
}

function AllNamespaceRow({ namespace, onDeleteClick }: AllNamespaceRowProps) {
  const projectCount = namespace.projects?.length || 0;
  const documentCount = namespace.documents?.length || 0;

  const {
    data: namespacePermissions,
    isLoading: isNamespacePermissionsLoading,
  } = usePermissions(withResourceType(ResourceType.Namespace, namespace.id));

  const hasNamespaceReadPermission = can(namespacePermissions, "read");
  const hasNamespaceWritePermission = can(namespacePermissions, "write");
  const hasNamespaceDeletePermission = can(namespacePermissions, "delete");
  const isPermissionsLoading = isNamespacePermissionsLoading;

  return (
    <TableRow>
      <TableCell className="font-medium">
        <ConditionalLink
          to="/settings/organizations/$organizationId/namespaces/$namespaceId"
          params={{
            organizationId: namespace.organizationId,
            namespaceId: namespace.id,
          }}
          condition={hasNamespaceReadPermission}
        >
          {namespace.name}
        </ConditionalLink>
      </TableCell>
      <TableCell>
        <Link
          to="/settings/organizations/$organizationId"
          params={{ organizationId: namespace.organizationId }}
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
        <Badge variant="secondary">
          {projectCount} {pluralize(projectCount, "project", "projects")}
        </Badge>
      </TableCell>
      <TableCell>
        <Badge variant="secondary">
          {documentCount} {pluralize(documentCount, "document", "documents")}
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
            {hasNamespaceWritePermission && (
              <Button variant="ghost" size="sm" asChild>
                <Link
                  to="/settings/organizations/$organizationId/namespaces/$namespaceId/edit"
                  params={{
                    organizationId: namespace.organizationId,
                    namespaceId: namespace.id,
                  }}
                >
                  <Edit className="size-4" />
                  <span className="sr-only">Edit namespace</span>
                </Link>
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
  namespaces: NamespaceWithOrganization[];
  isLoading: boolean;
  error: unknown;
}

export function AllNamespacesList({
  namespaces,
  isLoading,
  error,
}: AllNamespacesListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedNamespace, setSelectedNamespace] =
    useState<NamespaceWithOrganization | null>(null);

  // Check if user can create namespaces (has write permission on any organization)
  const { data: organizations } = useQuery(v1OrganizationsGetOptions());
  const permissionQueries = useQueries({
    queries:
      organizations && organizations.length > 0
        ? organizations.map((org) =>
            v1PermissionResourceGetOptions({
              path: {
                resourceId: withResourceType(ResourceType.Organization, org.id),
              },
            })
          )
        : [],
  });

  const canCreateNamespace = useMemo(() => {
    if (!organizations) return false;
    return organizations.some((org, index) => {
      const permissions = permissionQueries[index]?.data;
      return can(permissions, "write");
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

  const emptyState =
    namespaces.length === 0
      ? {
          icon: <Folder />,
          title: "No namespaces found",
          description:
            "You don't have access to any namespaces yet. Namespaces help organize projects and documents within organizations.",
        }
      : filteredNamespaces.length === 0 && searchTerm.trim()
        ? {
            icon: <Folder />,
            title: "No namespaces found",
            description:
              "No namespaces match your search criteria. Try adjusting your search.",
          }
        : undefined;

  const shouldShowSearch = namespaces.length > 0 || searchTerm.trim() !== "";

  const createButton = canCreateNamespace ? (
    <Button asChild>
      <Link to={"/settings/namespaces/new" as any}>
        <Plus className="size-4" />
        Create Namespace
      </Link>
    </Button>
  ) : undefined;

  return (
    <>
      <ListContainer
        data-section="all-namespaces"
        title="Namespaces"
        description="All namespaces you have access to across organizations."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        actionButton={createButton}
        searchInput={
          shouldShowSearch ? (
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Search namespaces or organizations..."
              disabled={isLoading}
            />
          ) : undefined
        }
      >
        {isLoading ? (
          <AllNamespacesListSkeleton />
        ) : (
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
              {filteredNamespaces.map((namespace) => (
                <AllNamespaceRow
                  key={namespace.id}
                  namespace={namespace}
                  onDeleteClick={handleDeleteClick}
                />
              ))}
            </TableBody>
          </Table>
        )}
      </ListContainer>

      {selectedNamespace && (
        <NamespaceDeleteDialog
          namespace={selectedNamespace}
          organizationId={selectedNamespace.organizationId}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
