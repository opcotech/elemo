import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Building2, Plus } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationCardSkeleton } from "./organization-card";
import { OrganizationRow } from "./organization-row";

import { Button } from "@/components/ui/button";
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
import { v1OrganizationsGetOptions } from "@/lib/api";
import { can } from "@/lib/auth/permissions";

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

export function OrganizationListSkeleton({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, i) => (
        <OrganizationCardSkeleton key={i} />
      ))}
    </div>
  );
}

export function OrganizationList() {
  const [searchTerm, setSearchTerm] = useState("");

  const {
    data: organizations,
    isLoading,
    error,
  } = useQuery(v1OrganizationsGetOptions());

  const { data: systemPermissions } = usePermissions(
    withResourceType(ResourceType.Organization)
  );
  const canCreate = can(systemPermissions, "create");

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

  const createButton = canCreate ? (
    <Button asChild>
      <Link to="/settings/organizations/new">
        <Plus className="size-4" />
        Create Organization
      </Link>
    </Button>
  ) : undefined;

  const emptyState =
    !organizations || organizations.length === 0
      ? {
          icon: <Building2 />,
          title: "No organizations available",
          description: "Get started by creating your first organization.",
          action: canCreate ? (
            <Button variant="outline" size="sm" asChild>
              <Link to="/settings/organizations/new">
                <Plus className="size-4" />
                Create Organization
              </Link>
            </Button>
          ) : undefined,
        }
      : filteredOrganizations.length === 0 && searchTerm.trim()
        ? {
            icon: <Building2 />,
            title: "No organizations found",
            description:
              "No organizations match your search criteria. Try adjusting your search.",
          }
        : undefined;

  // Show search input only when there's data to search through OR when search is active
  const shouldShowSearch =
    (organizations && organizations.length > 0) || searchTerm.trim() !== "";

  return (
    <ListContainer
      title="Organizations"
      description="View and manage organizations."
      isLoading={isLoading}
      error={error}
      emptyState={emptyState}
      actionButton={createButton}
      searchInput={
        shouldShowSearch ? (
          <SearchInput
            value={searchTerm}
            onChange={setSearchTerm}
            placeholder="Search organizations..."
            disabled={isLoading}
          />
        ) : undefined
      }
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Website</TableHead>
            <TableHead>Members</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>
              <span className="sr-only">Actions</span>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <OrganizationTableSkeletonRows />
          ) : (
            <>
              {filteredOrganizations.map((organization) => (
                <OrganizationRow
                  key={organization.id}
                  organization={organization}
                />
              ))}
            </>
          )}
        </TableBody>
      </Table>
    </ListContainer>
  );
}
