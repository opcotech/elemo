import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Building2, Plus } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationRow } from "./organization-row";

import { Button } from "@/components/ui/button";
import { ListContainer } from "@/components/ui/list-container";
import { SearchInput } from "@/components/ui/search-input";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TableSkeletonRows } from "@/components/ui/table-skeleton";
import {
  ResourceType,
  usePermissions,
  usePermissionsByResourceId,
  withResourceType,
} from "@/hooks/use-permissions";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1OrganizationsGetOptions } from "@/lib/api/query-options";
import { v1OrganizationsGet } from "@/lib/api/sdk";
import { Action, can } from "@/lib/auth/permissions";
import { zOrganizationStatus } from "@/lib/client/zod.gen";

const organizationTableSkeletonColumns = [
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Email", skeletonClassName: "h-5 w-40" },
  { header: "Website", skeletonClassName: "h-5 w-48" },
  { header: "Members", skeletonClassName: "h-6 w-16" },
  { header: "Status", skeletonClassName: "h-5 w-24" },
  {
    header: "Actions",
    skeletonClassName: "h-5 w-8",
    cellClassName: "text-right",
    srOnly: true,
    count: 3,
  },
] as const;

export function OrganizationList() {
  const [searchTerm, setSearchTerm] = useState("");

  const listOptions = v1OrganizationsGetOptions();
  const {
    data: organizationsPage,
    isLoading,
    error,
  } = useQuery(
    collectedListQuery(listOptions, async (pageToken, signal) => {
      const { data } = await v1OrganizationsGet({
        query: cursorPageQuery(pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    })
  );
  const organizations = organizationsPage?.items;

  const { data: systemPermissions } = usePermissions(
    withResourceType(ResourceType.Installation)
  );
  const organizationIds = (organizations || []).map(
    (organization) => organization.id
  );
  const organizationPermissionsById = usePermissionsByResourceId(
    ResourceType.Organization,
    organizationIds
  );
  const canCreate = can(systemPermissions, Action.OrganizationCreate);

  const sortedOrganizations = useMemo(() => {
    if (!organizations) return [];
    return [...organizations].sort((a, b) => {
      if (a.status !== b.status) {
        return a.status === zOrganizationStatus.enum.active ? -1 : 1;
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
    <Button render={<Link to="/settings/organizations/new" />}>
      <Plus className="size-4" />
      Create Organization
    </Button>
  ) : undefined;

  const emptyState =
    !organizations || organizations.length === 0
      ? {
          icon: <Building2 />,
          title: "No organizations available",
          description: "Get started by creating your first organization.",
          action: canCreate ? (
            <Button
              variant="outline"
              size="sm"
              render={<Link to="/settings/organizations/new" />}
            >
              <Plus className="size-4" />
              Create Organization
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

  const shouldShowSearch =
    (organizations && organizations.length > 0) || searchTerm.trim() !== "";

  return (
    <ListContainer
      data-section="organizations"
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
            <TableSkeletonRows columns={organizationTableSkeletonColumns} />
          ) : (
            <>
              {filteredOrganizations.map((organization) => {
                const permissionQuery = organizationPermissionsById.get(
                  organization.id
                );
                return (
                  <OrganizationRow
                    key={organization.id}
                    organization={organization}
                    permissions={permissionQuery?.data}
                    isPermissionsLoading={permissionQuery?.isLoading ?? true}
                  />
                );
              })}
            </>
          )}
        </TableBody>
      </Table>
    </ListContainer>
  );
}
