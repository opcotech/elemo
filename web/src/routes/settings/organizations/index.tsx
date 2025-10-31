import { useQuery } from "@tanstack/react-query";
import { Link, createFileRoute } from "@tanstack/react-router";
import { Edit, Eye, Search, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Organization } from "@/lib/api";
import { v1OrganizationsGetOptions } from "@/lib/api";
import { can } from "@/lib/auth/permissions";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";
import { formatDate } from "@/lib/utils";

export const Route = createFileRoute("/settings/organizations/")({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationsPage,
});

function OrganizationTableSkeletonRow() {
  return (
    <TableRow>
      <TableCell>
        <Skeleton className="h-5 w-32" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-5 w-40" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-5 w-48" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-6 w-16" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-5 w-24" />
      </TableCell>
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-x-1">
          <Skeleton className="h-5 w-8" />
          <Skeleton className="h-5 w-8" />
          <Skeleton className="h-5 w-8" />
        </div>
      </TableCell>
    </TableRow>
  );
}

function OrganizationTableSkeletonRows({ count = 5 }: { count?: number }) {
  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <OrganizationTableSkeletonRow key={i} />
      ))}
    </>
  );
}

function OrganizationRow({ organization }: { organization: Organization }) {
  const { data: permissions, isLoading: isPermissionsLoading } = usePermissions(
    withResourceType(ResourceType.Organization, organization.id)
  );

  const hasReadPermission = can(permissions, "read");
  const hasWritePermission = can(permissions, "write");
  const hasDeletePermission = can(permissions, "delete");

  return (
    <TableRow>
      <TableCell className="font-medium">
        <Link
          to="/settings/organizations/$organizationId"
          params={{ organizationId: organization.id }}
          className="text-primary hover:underline"
        >
          {organization.name}
        </Link>
      </TableCell>
      <TableCell>{organization.email}</TableCell>
      <TableCell>
        {organization.website ? (
          <a
            href={organization.website}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:underline"
          >
            {organization.website}
          </a>
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </TableCell>
      <TableCell>
        {organization.status === "active" ? (
          <Badge variant="success">Active</Badge>
        ) : (
          <Badge variant="destructive">Deleted</Badge>
        )}
      </TableCell>
      <TableCell>{formatDate(organization.created_at)}</TableCell>
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-x-1">
          {isPermissionsLoading ? (
            <div className="flex items-center gap-x-2 py-1.5">
              <Skeleton className="h-5 w-8" />
              <Skeleton className="h-5 w-8" />
              <Skeleton className="h-5 w-8" />
            </div>
          ) : (
            <>
              {hasReadPermission && (
                <Button variant="ghost" size="sm" asChild>
                  <Link
                    to="/settings/organizations/$organizationId"
                    params={{ organizationId: organization.id }}
                  >
                    <Eye className="size-4" />
                    <span className="sr-only">View organization</span>
                  </Link>
                </Button>
              )}
              {hasWritePermission && (
                <Button variant="ghost" size="sm" disabled>
                  <Edit className="size-4" />
                  <span className="sr-only">Edit organization</span>
                </Button>
              )}
              {hasDeletePermission && (
                <Button variant="ghost" size="sm" disabled>
                  <Trash2 className="size-4" />
                  <span className="sr-only">Delete organization</span>
                </Button>
              )}
            </>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}

function OrganizationRows({
  organizations,
  searchTerm,
  isLoading,
}: {
  organizations: Organization[];
  searchTerm: string;
  isLoading: boolean;
}) {
  if (isLoading) {
    return <OrganizationTableSkeletonRows />;
  }

  if (organizations.length === 0) {
    return (
      <TableRow>
        <TableCell colSpan={6} className="py-4 text-center">
          {searchTerm
            ? "No organizations found matching your search."
            : "No organizations available."}
        </TableCell>
      </TableRow>
    );
  }

  return (
    <>
      {organizations.map((organization) => (
        <OrganizationRow key={organization.id} organization={organization} />
      ))}
    </>
  );
}

function OrganizationsPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();
  const [searchTerm, setSearchTerm] = useState("");

  const {
    data: organizations,
    isLoading,
    error,
  } = useQuery(v1OrganizationsGetOptions());

  const sortedOrganizations = useMemo(() => {
    if (!organizations) return [];
    return [...organizations].sort((a, b) => {
      if (a.status !== b.status) {
        return a.status === "active" ? -1 : 1;
      }
      return a.name.localeCompare(b.name);
    });
  }, [organizations]);

  const filteredOrganizations = useMemo(() => {
    if (!searchTerm.trim()) return sortedOrganizations;
    const term = searchTerm.toLowerCase();
    return sortedOrganizations.filter((org) =>
      org.name.toLowerCase().includes(term)
    );
  }, [sortedOrganizations, searchTerm]);

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Organizations",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems]);

  if (error) {
    return (
      <div className="space-y-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold">Organizations</h1>
          <p className="mt-2 text-gray-600">View and manage organizations.</p>
        </div>
        <Alert variant="destructive">
          <AlertDescription>
            Failed to load organizations. Please try again later.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Organizations</h1>
        <p className="mt-2 text-gray-600">View and manage organizations.</p>
      </div>

      <div className="space-y-4">
        <div className="relative max-w-md flex-1">
          <Search className="text-muted-foreground absolute top-3 left-2 h-4 w-4" />
          <Input
            placeholder="Search organizations..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            disabled={isLoading}
            className="pl-8"
          />
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Website</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created At</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <OrganizationRows
              organizations={filteredOrganizations}
              searchTerm={searchTerm}
              isLoading={isLoading}
            />
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
