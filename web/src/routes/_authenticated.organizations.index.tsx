import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Building2Icon, PlusIcon } from "lucide-react";
import { useMemo } from "react";
import { z } from "zod";

import { ContentWidth } from "@/components/layout/content-width";
import { AppEmptyState } from "@/components/shared/app-feedback";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { PageHeader } from "@/components/shared/page-header";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Button } from "@/components/ui/button";
import { InternalLink } from "@/components/ui/internal-link";
import { SearchInput } from "@/components/ui/search-input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { v1OrganizationsGetOptions } from "@/lib/api/query-options";
import { can } from "@/lib/auth/permissions";
import { zOrganizationStatus } from "@/lib/client/zod.gen";
import { loadOrganizations } from "@/lib/route-data";
import { pluralize } from "@/lib/utils";

const organizationsListSearchSchema = z.object({
  q: z.string().optional().catch(undefined),
  status: z
    .union([z.literal("all"), zOrganizationStatus])
    .optional()
    .catch(undefined),
});

export const Route = createFileRoute("/_authenticated/organizations/")({
  validateSearch: organizationsListSearchSchema,
  loader: ({ context }) => loadOrganizations(context.queryClient),
  component: OrganizationsListPage,
});

function OrganizationsListPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: organizations, isLoading } = useQuery(
    v1OrganizationsGetOptions()
  );
  const { data: systemPermissions } = usePermissions(
    withResourceType(ResourceType.Organization)
  );
  const canCreate = can(systemPermissions, "create");
  const statusFilter = search.status ?? zOrganizationStatus.enum.active;

  const filteredOrganizations = useMemo(() => {
    if (!organizations) return [];
    const query = search.q?.trim().toLowerCase() ?? "";

    return organizations
      .filter((organization) => {
        if (statusFilter !== "all" && organization.status !== statusFilter) {
          return false;
        }
        if (!query) return true;
        return [organization.name, organization.email]
          .filter(Boolean)
          .join(" ")
          .toLowerCase()
          .includes(query);
      })
      .sort((a, b) => {
        if (a.status !== b.status) {
          return a.status === zOrganizationStatus.enum.active ? -1 : 1;
        }
        return a.name.localeCompare(b.name);
      });
  }, [organizations, search.q, statusFilter]);

  const hasActiveFilters = Boolean(
    search.q?.trim() || statusFilter !== zOrganizationStatus.enum.active
  );

  return (
    <ContentWidth width="overview" className="space-y-6">
      <PageHeader
        title="Organizations"
        description="Organizations you can open as operational context."
        actions={
          canCreate ? (
            <Button
              size="sm"
              render={<InternalLink to="/settings/organizations/new" />}
            >
              <PlusIcon />
              Create organization
            </Button>
          ) : undefined
        }
      />

      <div className="bg-background sticky top-0 z-10 flex flex-wrap items-center gap-2 py-3">
        <SearchInput
          value={search.q ?? ""}
          onChange={(value) =>
            void navigate({
              search: (previous) => ({
                ...previous,
                q: value || undefined,
              }),
              replace: true,
            })
          }
          placeholder="Search organizations..."
          className="min-w-60"
        />
        <Select
          value={statusFilter}
          onValueChange={(value) =>
            void navigate({
              search: (previous) => ({
                ...previous,
                status:
                  !value || value === zOrganizationStatus.enum.active
                    ? undefined
                    : value,
              }),
              replace: true,
            })
          }
          items={{
            all: "All statuses",
            ...Object.fromEntries(
              zOrganizationStatus.options.map((status) => [
                status,
                status.charAt(0).toUpperCase() + status.slice(1),
              ])
            ),
          }}
        >
          <SelectTrigger className="w-44">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All statuses</SelectItem>
            {zOrganizationStatus.options.map((status) => (
              <SelectItem key={status} value={status}>
                {status.charAt(0).toUpperCase() + status.slice(1)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <p className="text-muted-foreground text-sm">Loading organizations…</p>
      ) : filteredOrganizations.length > 0 ? (
        <AppList>
          {filteredOrganizations.map((organization) => (
            <EntityLink
              key={organization.id}
              type="organization"
              href={`/organizations/${organization.id}`}
              title={organization.name}
              imageUrl={organization.logo}
              subtitle={
                <span className="flex items-center gap-2">
                  <StatusIndicator status={organization.status} />
                  <span className="truncate">
                    {organization.email ||
                      `${organization.members.length} ${pluralize(
                        organization.members.length,
                        "member",
                        "members"
                      )}`}
                  </span>
                </span>
              }
            />
          ))}
        </AppList>
      ) : (
        <AppEmptyState
          icon={<Building2Icon />}
          title={
            hasActiveFilters ? "No matching organizations" : "No organizations"
          }
          description={
            hasActiveFilters
              ? "Try a different search or status filter."
              : "Join or create an organization from Settings to establish workspace context."
          }
          action={
            hasActiveFilters ? (
              <Button
                variant="outline"
                onClick={() =>
                  void navigate({
                    search: { q: undefined, status: undefined },
                    replace: true,
                  })
                }
              >
                Clear filters
              </Button>
            ) : canCreate ? (
              <Button
                variant="outline"
                render={<InternalLink to="/settings/organizations/new" />}
              >
                <PlusIcon />
                Create organization
              </Button>
            ) : (
              <Button
                variant="outline"
                render={<InternalLink to="/settings/organizations" />}
              >
                Manage organizations
              </Button>
            )
          }
        />
      )}
    </ContentWidth>
  );
}
