import {
  ArrowRightIcon,
  Building2Icon,
  Layers3Icon,
  ShieldIcon,
  UsersIcon,
} from "lucide-react";
import { useMemo } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { NamespaceEntitySubtitle } from "@/components/namespaces";
import { AppEmptyState } from "@/components/shared/app-feedback";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { PropertyList } from "@/components/shared/property-list";
import { Section } from "@/components/shared/section";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ExternalLink } from "@/components/ui/external-link";
import { InternalLink } from "@/components/ui/internal-link";
import { UserAvatarCompact } from "@/components/ui/user-avatar";
import { useAuth } from "@/hooks/use-auth";
import { can } from "@/lib/auth/permissions";
import { zUserStatus } from "@/lib/client/zod.gen";
import { formatDate } from "@/lib/format-date";
import { internalPath } from "@/lib/internal-url";
import { sortOrganizationMembers } from "@/lib/organization-members";
import type { OrganizationWorkspaceData } from "@/lib/organization-workspace";
import { pluralize } from "@/lib/utils";

export function OrganizationOverviewPage({
  data,
}: {
  data: OrganizationWorkspaceData;
}) {
  const { user } = useAuth();
  const currentUserId = user?.id ?? null;
  const {
    organization,
    members,
    namespaces,
    roles,
    permissions,
    hasReadAccess,
  } = data;
  const organizationId = organization.id;
  const hasOrgWritePermission = can(permissions, "write");

  const sortedMembers = useMemo(() => {
    return sortOrganizationMembers(members);
  }, [members]);

  const sortedNamespaces = useMemo(() => {
    return [...namespaces].sort((a, b) => a.name.localeCompare(b.name));
  }, [namespaces]);

  const sortedRoles = useMemo(() => {
    return [...roles].sort((a, b) => a.name.localeCompare(b.name));
  }, [roles]);

  const settingsHref = `/settings/organizations/${organizationId}`;

  return (
    <ContentWidth width="overview" className="space-y-7">
      <EntityHeader
        type="organization"
        title={organization.name}
        description={
          organization.email || "Organization members, roles, and namespaces."
        }
        imageUrl={organization.logo}
        actions={
          <PageActions
            secondary={[
              ...(hasOrgWritePermission
                ? [
                    {
                      label: "Edit organization",
                      href: `${settingsHref}/edit`,
                    },
                  ]
                : []),
            ]}
          />
        }
      />

      <div className="grid gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <div className="space-y-8">
          {hasReadAccess ? (
            <>
              <Section
                title="Namespaces"
                action={
                  <Button
                    variant="ghost"
                    size="sm"
                    render={
                      <InternalLink
                        to="/namespaces"
                        search={{ organization: organizationId }}
                      />
                    }
                  >
                    View all <ArrowRightIcon />
                  </Button>
                }
              >
                {sortedNamespaces.length > 0 ? (
                  <AppList>
                    {sortedNamespaces.slice(0, 6).map((namespace) => (
                      <EntityLink
                        key={namespace.id}
                        type="namespace"
                        href={`/namespaces/${namespace.id}`}
                        title={namespace.name}
                        subtitle={
                          <NamespaceEntitySubtitle
                            description={namespace.description}
                            projectCount={namespace.projects.length}
                            documentCount={namespace.documents.length}
                          />
                        }
                      />
                    ))}
                  </AppList>
                ) : (
                  <AppEmptyState
                    compact
                    icon={<Layers3Icon />}
                    title="No namespaces"
                    description="Namespaces for this organization will appear here."
                    action={
                      hasOrgWritePermission ? (
                        <Button
                          variant="outline"
                          render={
                            <InternalLink
                              to={internalPath(
                                `${settingsHref}/namespaces/new`
                              )}
                            />
                          }
                        >
                          Create namespace
                        </Button>
                      ) : undefined
                    }
                  />
                )}
              </Section>

              <Section
                title="Members"
                action={
                  <Button
                    variant="ghost"
                    size="sm"
                    render={<InternalLink to={internalPath(settingsHref)} />}
                  >
                    View all <ArrowRightIcon />
                  </Button>
                }
              >
                {sortedMembers.length > 0 ? (
                  <AppList>
                    {sortedMembers.slice(0, 6).map((member) => {
                      const fullName = `${member.first_name} ${member.last_name}`;
                      const isCurrentUser = currentUserId === member.id;
                      return (
                        <div
                          key={member.id}
                          role="listitem"
                          className="flex min-w-0 items-center gap-2.5 px-3 py-2.5"
                        >
                          <UserAvatarCompact
                            firstName={member.first_name}
                            lastName={member.last_name}
                            picture={member.picture}
                            size="sm"
                          />
                          <div className="min-w-0 flex-1">
                            <div className="flex min-w-0 items-center gap-2">
                              <p className="truncate text-sm font-medium">
                                {fullName}
                              </p>
                              {isCurrentUser && (
                                <Badge variant="secondary">You</Badge>
                              )}
                              {member.status !== zUserStatus.enum.active && (
                                <Badge
                                  variant={
                                    member.status === zUserStatus.enum.deleted
                                      ? "destructive"
                                      : "outline"
                                  }
                                >
                                  {member.status}
                                </Badge>
                              )}
                            </div>
                            <p className="text-muted-foreground truncate text-xs">
                              {member.roles.length > 0
                                ? member.roles.join(", ")
                                : member.email}
                            </p>
                          </div>
                        </div>
                      );
                    })}
                  </AppList>
                ) : (
                  <AppEmptyState
                    compact
                    icon={<UsersIcon />}
                    title="No members"
                    description="Members will appear here once they join this organization."
                    action={
                      hasOrgWritePermission ? (
                        <Button
                          variant="outline"
                          render={
                            <InternalLink to={internalPath(settingsHref)} />
                          }
                        >
                          Manage in Settings
                        </Button>
                      ) : undefined
                    }
                  />
                )}
              </Section>

              <Section
                title="Roles"
                action={
                  <Button
                    variant="ghost"
                    size="sm"
                    render={<InternalLink to={internalPath(settingsHref)} />}
                  >
                    View all <ArrowRightIcon />
                  </Button>
                }
              >
                {sortedRoles.length > 0 ? (
                  <AppList>
                    {sortedRoles.slice(0, 6).map((role) => (
                      <div
                        key={role.id}
                        role="listitem"
                        className="flex min-w-0 items-center gap-2.5 px-3 py-2.5"
                      >
                        <div className="bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-lg">
                          <ShieldIcon className="size-4" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-sm font-medium">
                            {role.name}
                          </p>
                          <p className="text-muted-foreground truncate text-xs">
                            {role.description ||
                              `${role.members.length} ${pluralize(
                                role.members.length,
                                "member",
                                "members"
                              )}`}
                          </p>
                        </div>
                      </div>
                    ))}
                  </AppList>
                ) : (
                  <AppEmptyState
                    compact
                    icon={<ShieldIcon />}
                    title="No roles"
                    description="Roles organize permissions and member access."
                    action={
                      hasOrgWritePermission ? (
                        <Button
                          variant="outline"
                          render={
                            <InternalLink
                              to={internalPath(`${settingsHref}/roles/new`)}
                            />
                          }
                        >
                          Create role
                        </Button>
                      ) : undefined
                    }
                  />
                )}
              </Section>
            </>
          ) : (
            <AppEmptyState
              icon={<Building2Icon />}
              title="Limited access"
              description="You can view basic organization details, but members, roles, and namespaces require read permission."
            />
          )}
        </div>

        <div className="space-y-8">
          <Section title="Details">
            <PropertyList
              items={[
                {
                  label: "Status",
                  value: <StatusIndicator status={organization.status} />,
                },
                {
                  label: "Email",
                  value: organization.email ? (
                    <ExternalLink href={`mailto:${organization.email}`}>
                      {organization.email}
                    </ExternalLink>
                  ) : (
                    "—"
                  ),
                },
                {
                  label: "Website",
                  value: organization.website ? (
                    <ExternalLink href={organization.website} />
                  ) : (
                    "—"
                  ),
                },
                {
                  label: "Created",
                  value: formatDate(organization.created_at),
                },
              ]}
            />
          </Section>
        </div>
      </div>
    </ContentWidth>
  );
}
