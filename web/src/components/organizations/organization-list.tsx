import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Plus, Search } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationRow } from "./organization-row";
import { OrganizationTableSkeletonRows } from "./organization-table-skeleton";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
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

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Organizations</CardTitle>
          <CardDescription>View and manage organizations.</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="destructive">
            <AlertDescription>
              Failed to load organizations. Please try again later.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle>Organizations</CardTitle>
            <CardDescription>View and manage organizations.</CardDescription>
          </div>
          {canCreate && (
            <Button asChild>
              <Link to="/settings/organizations/new">
                <Plus className="size-4" />
                Create Organization
              </Link>
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
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
            ) : filteredOrganizations.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="py-4 text-center">
                  {searchTerm
                    ? "No organizations found matching your search."
                    : "No organizations available."}
                </TableCell>
              </TableRow>
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
      </CardContent>
    </Card>
  );
}
