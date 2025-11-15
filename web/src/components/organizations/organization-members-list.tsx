import { UserMinus, UserPlus, Users, X } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationMemberInviteDialog } from "./organization-member-invite-dialog";
import { OrganizationMemberInviteRevokeDialog } from "./organization-member-invite-revoke-dialog";
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
  const [inviteDialogOpen, setInviteDialogOpen] = useState(false);
  const [revokeInviteDialogOpen, setRevokeInviteDialogOpen] = useState(false);
  const [selectedMember, setSelectedMember] =
    useState<OrganizationMember | null>(null);

  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasOrgReadPermission = can(orgPermissions, "read");
  const hasOrgWritePermission = can(orgPermissions, "write");
  const isPermissionsLoading = isOrgPermissionsLoading;

  if (!isPermissionsLoading && !hasOrgReadPermission) {
    return null;
  }

  const handleRemoveClick = (member: OrganizationMember) => {
    setSelectedMember(member);
    setRemoveMemberDialogOpen(true);
  };

  const handleRemoveSuccess = () => {
    setRemoveMemberDialogOpen(false);
    setSelectedMember(null);
  };

  const handleInviteSuccess = () => {
    setInviteDialogOpen(false);
  };

  const handleRevokeInviteClick = (member: OrganizationMember) => {
    setSelectedMember(member);
    setRevokeInviteDialogOpen(true);
  };

  const handleRevokeInviteSuccess = () => {
    setRevokeInviteDialogOpen(false);
    setSelectedMember(null);
  };

  const sortedMembers = useMemo(() => {
    if (!members) return [];
    return [...members].sort((a, b) => {
      // Pending members come first
      if (a.status === "pending" && b.status !== "pending") return -1;
      if (a.status !== "pending" && b.status === "pending") return 1;

      // Deleted members come last
      if (a.status === "deleted" && b.status !== "deleted") return 1;
      if (a.status !== "deleted" && b.status === "deleted") return -1;

      // Within same status, sort alphabetically
      const aName = `${a.first_name} ${a.last_name}`.toLowerCase();
      const bName = `${b.first_name} ${b.last_name}`.toLowerCase();
      return aName.localeCompare(bName);
    });
  }, [members]);

  const filteredMembers = useMemo(() => {
    if (!sortedMembers || !searchTerm.trim()) return sortedMembers || [];
    const term = searchTerm.toLowerCase();
    return sortedMembers.filter(
      (member) =>
        member.first_name.toLowerCase().includes(term) ||
        member.last_name.toLowerCase().includes(term) ||
        member.email.toLowerCase().includes(term) ||
        member.roles.some((role) => role.toLowerCase().includes(term))
    );
  }, [sortedMembers, searchTerm]);

  const emptyState =
    !sortedMembers || sortedMembers.length === 0
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

  const shouldShowSearch =
    (sortedMembers && sortedMembers.length > 0) || searchTerm.trim() !== "";

  const inviteButton = hasOrgWritePermission ? (
    <Button
      variant="outline"
      size="sm"
      onClick={() => setInviteDialogOpen(true)}
    >
      <UserPlus className="size-4" />
      Invite Member
    </Button>
  ) : undefined;

  return (
    <>
      <ListContainer
        data-section="organization-members"
        title="Members"
        description="Organization members and their roles."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        actionButton={inviteButton}
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
                          <div className="flex items-center justify-end gap-2">
                            {member.status === "pending" && (
                              <Button
                                variant="destructive-ghost"
                                size="sm"
                                onClick={() => handleRevokeInviteClick(member)}
                                title="Revoke invitation"
                              >
                                <X className="size-4" />
                                <span className="sr-only">
                                  Revoke invitation
                                </span>
                              </Button>
                            )}
                            {member.status !== "pending" && (
                              <Button
                                variant="destructive-ghost"
                                size="sm"
                                onClick={() => handleRemoveClick(member)}
                                title="Remove member"
                              >
                                <UserMinus className="size-4" />
                                <span className="sr-only">Remove member</span>
                              </Button>
                            )}
                          </div>
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
        <>
          <OrganizationMemberRemoveDialog
            member={selectedMember}
            organizationId={organizationId}
            open={removeMemberDialogOpen}
            onOpenChange={setRemoveMemberDialogOpen}
            onSuccess={handleRemoveSuccess}
          />
          <OrganizationMemberInviteRevokeDialog
            member={selectedMember}
            organizationId={organizationId}
            open={revokeInviteDialogOpen}
            onOpenChange={setRevokeInviteDialogOpen}
            onSuccess={handleRevokeInviteSuccess}
          />
        </>
      )}
      <OrganizationMemberInviteDialog
        organizationId={organizationId}
        open={inviteDialogOpen}
        onOpenChange={setInviteDialogOpen}
        onSuccess={handleInviteSuccess}
      />
    </>
  );
}
