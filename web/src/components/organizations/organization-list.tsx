import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Building2, Plus } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationRow } from "./organization-row";

import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TableSkeleton } from "@/components/ui/table-skeleton";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import {
  ResourceType,
  usePermissions,
  usePermissionsByResourceId,
  withResourceType,
} from "@/hooks/use-permissions";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1OrganizationsGetOptions } from "@/lib/api/query-options";
import { zOrganizationStatus } from "@/lib/api/schemas";
import { Action, can } from "@/lib/auth/permissions";

const organizationTableSkeletonColumns = [
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Email", skeletonClassName: "h-5 w-40" },
  { header: "Website", skeletonClassName: "h-5 w-48" },
  { header: "Members", skeletonClassName: "h-6 w-16" },
  { header: "Status", skeletonClassName: "h-5 w-24" },
  {
    header: "Actions",
    skeletonClassName: "h-5 w-8",
    headerClassName: "text-right",
    cellClassName: "text-right",
    srOnly: true,
    count: 3,
  },
] as const;

export function OrganizationList() {
  const [searchTerm, setSearchTerm] = useState("");
  const pageNav = useCursorPageNav({ resetKey: searchTerm });
  const {
    data: organizationsPage,
    isLoading,
    error,
  } = useQuery(
    v1OrganizationsGetOptions({
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const organizations = organizationsPage?.items ?? [];

  const { data: systemPermissions } = usePermissions(
    withResourceType(ResourceType.Installation)
  );
  const organizationIds = organizations.map((organization) => organization.id);
  const organizationPermissionsById = usePermissionsByResourceId(
    ResourceType.Organization,
    organizationIds
  );
  const canCreate = can(systemPermissions, Action.OrganizationCreate);

  const sortedOrganizations = useMemo(() => {
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
    <Button
      variant="outline"
      size="sm"
      render={<Link to="/settings/organizations/new" />}
    >
      <Plus className="size-4" />
      Create Organization
    </Button>
  ) : undefined;

  return (
    <SettingsResourceTable
      dataSection="organizations"
      title="Organizations"
      description="View and manage organizations."
      isLoading={isLoading}
      error={error}
      actionButton={createButton}
      search={{
        value: searchTerm,
        onChange: setSearchTerm,
        placeholder: "Search organizations...",
        itemCount: organizations.length,
      }}
      empty={{
        icon: <Building2 />,
        title: "No organizations available",
        description: "Get started by creating your first organization.",
        action: createButton,
        searchTitle: "No organizations found",
        searchDescription:
          "No organizations match your search criteria. Try adjusting your search.",
        hasItems: organizations.length > 0,
        hasFilteredItems: filteredOrganizations.length > 0,
      }}
      skeleton={<TableSkeleton columns={organizationTableSkeletonColumns} />}
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
        </TableBody>
      </Table>
      <CursorPaginator {...cursorPaginatorProps(organizationsPage, pageNav)} />
    </SettingsResourceTable>
  );
}
