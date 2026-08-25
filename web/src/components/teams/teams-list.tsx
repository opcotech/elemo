import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Edit, Plus, Trash2, Users } from "lucide-react";
import { useMemo, useState } from "react";

import { TeamDeleteDialog } from "./team-delete-dialog";

import { SettingsResourceTable } from "@/components/settings/settings-resource-table";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { Button } from "@/components/ui/button";
import { CountBadge } from "@/components/ui/count-badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TableSkeleton } from "@/components/ui/table-skeleton";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1OrganizationTeamsGetOptions } from "@/lib/api/query-options";
import type { EffectiveActions, Team } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";

const teamsListSkeletonColumns = [
  { header: "Name", skeletonClassName: "h-5 w-32" },
  { header: "Description", skeletonClassName: "h-4 w-48" },
  { header: "Members", skeletonClassName: "h-6 w-16" },
  {
    header: "Actions",
    skeletonClassName: "h-8 w-8",
    headerClassName: "text-right",
    cellClassName: "text-right",
    count: 2,
  },
] as const;

interface TeamsListProps {
  organizationId: string;
  organizationSlug: string;
  organizationPermissions: EffectiveActions;
}

export function TeamsList({
  organizationId,
  organizationSlug,
  organizationPermissions,
}: TeamsListProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const pageNav = useCursorPageNav({ resetKey: searchTerm });
  const {
    data: teamsPage,
    isLoading,
    error,
  } = useQuery(
    v1OrganizationTeamsGetOptions({
      path: { organizationRef: organizationId },
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const teams = teamsPage?.items ?? [];

  const canManageTeams = can(organizationPermissions, Action.TeamManage);

  const filteredTeams = useMemo(() => {
    if (!searchTerm.trim()) return teams;
    const term = searchTerm.toLowerCase();
    return teams.filter(
      (team) =>
        team.name.toLowerCase().includes(term) ||
        (team.description && team.description.toLowerCase().includes(term))
    );
  }, [teams, searchTerm]);

  const createButton = canManageTeams ? (
    <Button
      variant="outline"
      size="sm"
      render={
        <Link
          to="/settings/organizations/$organizationSlug/teams/new"
          params={{ organizationSlug }}
        />
      }
    >
      <Plus className="size-4" />
      Create Team
    </Button>
  ) : undefined;

  return (
    <>
      <SettingsResourceTable
        dataSection="teams"
        title="Teams"
        description="Teams are principals that can hold grants. Membership is MEMBER_OF."
        isLoading={isLoading}
        error={error}
        actionButton={createButton}
        search={{
          value: searchTerm,
          onChange: setSearchTerm,
          placeholder: "Search teams...",
          itemCount: teams.length,
        }}
        empty={{
          icon: <Users />,
          title: "No teams found",
          description:
            "Create a team, add members, then grant the team actions on a scope.",
          action: createButton,
          searchTitle: "No teams found",
          searchDescription:
            "No teams match your search criteria. Try adjusting your search.",
          hasItems: teams.length > 0,
          hasFilteredItems: filteredTeams.length > 0,
        }}
        skeleton={<TableSkeleton columns={teamsListSkeletonColumns} />}
      >
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Members</TableHead>
              <TableHead>
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredTeams.map((team) => (
              <TableRow key={team.id}>
                <TableCell className="font-medium">{team.name}</TableCell>
                <TableCell>
                  <span className="text-muted-foreground text-sm">
                    {team.description || "—"}
                  </span>
                </TableCell>
                <TableCell>
                  <CountBadge
                    count={team.member_count ?? 0}
                    singular="member"
                    plural="members"
                  />
                </TableCell>
                <TableCell className="text-right">
                  {canManageTeams ? (
                    <div className="flex items-center justify-end gap-x-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        render={
                          <Link
                            to="/settings/organizations/$organizationSlug/teams/$teamId/edit"
                            params={{
                              organizationSlug,
                              teamId: team.id,
                            }}
                          />
                        }
                      >
                        <Edit className="size-4" />
                        <span className="sr-only">Edit team</span>
                      </Button>
                      <Button
                        variant="destructive-ghost"
                        size="sm"
                        onClick={() => {
                          setSelectedTeam(team);
                          setDeleteDialogOpen(true);
                        }}
                      >
                        <Trash2 className="size-4" />
                        <span className="sr-only">Delete team</span>
                      </Button>
                    </div>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <CursorPaginator {...cursorPaginatorProps(teamsPage, pageNav)} />
      </SettingsResourceTable>

      {selectedTeam && (
        <TeamDeleteDialog
          team={selectedTeam}
          organizationId={organizationId}
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          onSuccess={() => {
            setDeleteDialogOpen(false);
            setSelectedTeam(null);
          }}
        />
      )}
    </>
  );
}
