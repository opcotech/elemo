import { Trash2, Users } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationMemberRemoveDialog } from "./organization-member-remove-dialog";

import { Badge } from "@/components/ui/badge";
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
import { UserAvatarCompact } from "@/components/ui/user-avatar";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { OrganizationMember } from "@/lib/api";
import { can } from "@/lib/auth/permissions";

function OrganizationMembersListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Roles</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <div className="flex items-center gap-3">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-1">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-3 w-24" />
                </div>
              </div>
            </TableCell>
            <TableCell>
              <div className="flex gap-1">
                <Skeleton className="h-6 w-16" />
                <Skeleton className="h-6 w-16" />
              </div>
            </TableCell>
            <TableCell>
              <Skeleton className="h-6 w-16" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

export function OrganizationMembersList({
  members,
  isLoading,
  error,
  currentUserId,
  organizationId,
}: {
  members: OrganizationMember[];
  isLoading: boolean;
  error: unknown;
  currentUserId?: string | null;
  organizationId: string;
}) {
  const [searchTerm, setSearchTerm] = useState("");
  const [removeMemberDialogOpen, setRemoveMemberDialogOpen] = useState(false);
  const [selectedMember, setSelectedMember] =
    useState<OrganizationMember | null>(null);

  // Check permissions for organization (write)
  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasOrgWritePermission = can(orgPermissions, "write");
  const isPermissionsLoading = isOrgPermissionsLoading;

  const handleRemoveClick = (member: OrganizationMember) => {
    setSelectedMember(member);
    setRemoveMemberDialogOpen(true);
  };

  const handleRemoveSuccess = () => {
    setRemoveMemberDialogOpen(false);
    setSelectedMember(null);
  };

  const filteredMembers = useMemo(() => {
    if (!members || !searchTerm.trim()) return members || [];
    const term = searchTerm.toLowerCase();
    return members.filter(
      (member) =>
        member.first_name.toLowerCase().includes(term) ||
        member.last_name.toLowerCase().includes(term) ||
        member.email.toLowerCase().includes(term) ||
        member.roles.some((role) => role.toLowerCase().includes(term))
    );
  }, [members, searchTerm]);

  // Only show empty state when there's no data at all (not filtered)
  // When filtered results are empty but original data exists, show search + empty state
  const emptyState =
    !members || members.length === 0
      ? {
          icon: <Users />,
          title: "No members found",
          description:
            "This organization doesn't have any members yet. Members will appear here once they are added.",
        }
      : filteredMembers.length === 0 && searchTerm.trim()
        ? {
            icon: <Users />,
            title: "No members found",
            description:
              "No members match your search criteria. Try adjusting your search.",
          }
        : undefined;

  // Show search input only when there's data to search through OR when search is active
  const shouldShowSearch =
    (members && members.length > 0) || searchTerm.trim() !== "";

  return (
    <>
      <ListContainer
        title="Members"
        description="Organization members and their roles."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        searchInput={
          shouldShowSearch ? (
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Search members..."
              disabled={isLoading}
            />
          ) : undefined
        }
      >
        {isLoading ? (
          <OrganizationMembersListSkeleton />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Roles</TableHead>
                <TableHead>Status</TableHead>
                {hasOrgWritePermission && (
                  <TableHead>
                    <span className="sr-only">Actions</span>
                  </TableHead>
                )}
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredMembers.map((member) => {
                const isCurrentUser = currentUserId === member.id;
                const fullName = `${member.first_name} ${member.last_name}`;

                return (
                  <TableRow key={member.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <UserAvatarCompact
                          firstName={member.first_name}
                          lastName={member.last_name}
                          picture={member.picture}
                        />
                        <div className="flex flex-col gap-0.5">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{fullName}</span>
                            {isCurrentUser && (
                              <Badge className="px-2 py-0.5 text-xs">You</Badge>
                            )}
                          </div>
                          <span className="text-muted-foreground text-sm">
                            {member.email}
                          </span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {member.roles.map((roleName: string) => (
                          <Badge key={roleName} variant="secondary">
                            {roleName}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell>
                      {member.status === "active" ? (
                        <Badge variant="success">Active</Badge>
                      ) : member.status === "deleted" ? (
                        <Badge variant="destructive">Deleted</Badge>
                      ) : (
                        <Badge variant="outline">{member.status}</Badge>
                      )}
                    </TableCell>
                    {hasOrgWritePermission && (
                      <TableCell className="text-right">
                        {isPermissionsLoading ? (
                          <Skeleton className="h-8 w-8" />
                        ) : (
                          <Button
                            variant="destructive-ghost"
                            size="sm"
                            onClick={() => handleRemoveClick(member)}
                          >
                            <Trash2 className="size-4" />
                            <span className="sr-only">Remove member</span>
                          </Button>
                        )}
                      </TableCell>
                    )}
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </ListContainer>

      {selectedMember && (
        <OrganizationMemberRemoveDialog
          member={selectedMember}
          organizationId={organizationId}
          open={removeMemberDialogOpen}
          onOpenChange={setRemoveMemberDialogOpen}
          onSuccess={handleRemoveSuccess}
        />
      )}
    </>
  );
}
