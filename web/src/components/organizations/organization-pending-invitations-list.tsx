import { Mail, X } from "lucide-react";
import { useMemo, useState } from "react";

import { OrganizationMemberInviteRevokeDialog } from "./organization-member-invite-revoke-dialog";

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

function OrganizationPendingInvitationsListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Email</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>
            <span className="sr-only">Actions</span>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 3 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <div className="flex items-center gap-3">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="space-y-1">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-3 w-48" />
                </div>
              </div>
            </TableCell>
            <TableCell>
              <Skeleton className="h-6 w-16" />
            </TableCell>
            <TableCell className="text-right">
              <Skeleton className="h-8 w-8" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

export function OrganizationPendingInvitationsList({
  members,
  isLoading,
  error,
  organizationId,
}: {
  members: OrganizationMember[];
  isLoading: boolean;
  error: unknown;
  organizationId: string;
}) {
  const [searchTerm, setSearchTerm] = useState("");
  const [revokeInviteDialogOpen, setRevokeInviteDialogOpen] = useState(false);
  const [selectedMember, setSelectedMember] =
    useState<OrganizationMember | null>(null);

  // Filter to only pending members
  const pendingMembers = useMemo(() => {
    return (members || []).filter((member) => member.status === "pending");
  }, [members]);

  // Check permissions for organization (write)
  const { data: orgPermissions, isLoading: isOrgPermissionsLoading } =
    usePermissions(withResourceType(ResourceType.Organization, organizationId));

  const hasOrgWritePermission = can(orgPermissions, "write");
  const isPermissionsLoading = isOrgPermissionsLoading;

  const handleRevokeInviteClick = (member: OrganizationMember) => {
    setSelectedMember(member);
    setRevokeInviteDialogOpen(true);
  };

  const handleRevokeInviteSuccess = () => {
    setRevokeInviteDialogOpen(false);
    setSelectedMember(null);
  };

  const filteredMembers = useMemo(() => {
    if (!pendingMembers || !searchTerm.trim()) return pendingMembers || [];
    const term = searchTerm.toLowerCase();
    return pendingMembers.filter(
      (member) =>
        member.first_name.toLowerCase().includes(term) ||
        member.last_name.toLowerCase().includes(term) ||
        member.email.toLowerCase().includes(term)
    );
  }, [pendingMembers, searchTerm]);

  // Only show empty state when there's no data at all (not filtered)
  // When filtered results are empty but original data exists, show search + empty state
  const emptyState =
    !pendingMembers || pendingMembers.length === 0
      ? {
          icon: <Mail />,
          title: "No pending invitations",
          description:
            "All pending invitations will appear here. Invite members to get started.",
        }
      : filteredMembers.length === 0 && searchTerm.trim()
        ? {
            icon: <Mail />,
            title: "No invitations found",
            description:
              "No pending invitations match your search criteria. Try adjusting your search.",
          }
        : undefined;

  // Show search input only when there's data to search through OR when search is active
  const shouldShowSearch =
    (pendingMembers && pendingMembers.length > 0) || searchTerm.trim() !== "";

  return (
    <>
      <ListContainer
        title="Pending Invitations"
        description="Users who have been invited but haven't accepted yet."
        isLoading={isLoading}
        error={error}
        emptyState={emptyState}
        searchInput={
          shouldShowSearch ? (
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Search invitations..."
              disabled={isLoading}
            />
          ) : undefined
        }
      >
        {isLoading ? (
          <OrganizationPendingInvitationsListSkeleton />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
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
                          <span className="font-medium">{fullName}</span>
                          <span className="text-muted-foreground text-sm">
                            {member.email}
                          </span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{member.status}</Badge>
                    </TableCell>
                    {hasOrgWritePermission && (
                      <TableCell className="text-right">
                        {isPermissionsLoading ? (
                          <Skeleton className="h-8 w-8" />
                        ) : (
                          <Button
                            variant="destructive-ghost"
                            size="sm"
                            onClick={() => handleRevokeInviteClick(member)}
                            title="Revoke invitation"
                          >
                            <X className="size-4" />
                            <span className="sr-only">Revoke invitation</span>
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
        <OrganizationMemberInviteRevokeDialog
          member={selectedMember}
          organizationId={organizationId}
          open={revokeInviteDialogOpen}
          onOpenChange={setRevokeInviteDialogOpen}
          onSuccess={handleRevokeInviteSuccess}
        />
      )}
    </>
  );
}
