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
import { pluralize } from "@/lib/utils";

function NamespacesListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
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

interface NamespaceRowProps {
  namespace: Namespace;
  organizationId: string;
  onDeleteClick: (namespace: Namespace) => void;
}

function NamespaceRow({
  namespace,
  organizationId,
  onDeleteClick,
}: NamespaceRowProps) {
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
                    organizationId,
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

interface NamespacesListProps {
  namespaces: Namespace[];
  isLoading: boolean;
  error: unknown;
  organizationId: string;
}

export function NamespacesList({
  namespaces,
  isLoading,
  error,
  organizationId,
}: NamespacesListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedNamespace, setSelectedNamespace] = useState<Namespace | null>(
    null
  );

  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasOrgReadPermission = can(orgPermissions, "read");
  const hasOrgWritePermission = can(orgPermissions, "write");
  const hasCreatePermission = hasOrgWritePermission;
  const isPermissionsLoading = isOrgPermissionsLoading;

  if (!isPermissionsLoading && !hasOrgReadPermission) {
    return null;
  }

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

  const emptyState =
    namespaces.length === 0
      ? {
          icon: <Folder />,
          title: "No namespaces found",
          description:
            "Namespaces help organize projects and documents. Create a namespace to get started.",
          action: hasCreatePermission ? (
            <Button variant="outline" size="sm" asChild>
              <Link
                to="/settings/organizations/$organizationId/namespaces/new"
                params={{ organizationId }}
              >
                <Plus className="size-4" />
                Create Namespace
              </Link>
            </Button>
          ) : undefined,
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

  const createButton =
    !isPermissionsLoading && hasCreatePermission ? (
      <Button variant="outline" size="sm" asChild>
        <Link
          to="/settings/organizations/$organizationId/namespaces/new"
          params={{ organizationId }}
        >
          <Plus className="size-4" />
          Create Namespace
        </Link>
      </Button>
    ) : undefined;

  return (
    <>
      <ListContainer
        data-section="organization-namespaces"
        title="Namespaces"
        description="Organization namespaces and their resources."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        actionButton={createButton}
        searchInput={
          shouldShowSearch ? (
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Search namespaces..."
              disabled={isLoading}
            />
          ) : undefined
        }
      >
        {isLoading ? (
          <NamespacesListSkeleton />
        ) : (
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
              {filteredNamespaces.map((namespace) => (
                <NamespaceRow
                  key={namespace.id}
                  namespace={namespace}
                  organizationId={organizationId}
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
          organizationId={organizationId}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={handleDeleteSuccess}
        />
      )}
    </>
  );
}
