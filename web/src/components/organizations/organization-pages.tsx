import { useQuery } from "@tanstack/react-query";
import {
  ArrowRightIcon,
  Building2Icon,
  FileTextIcon,
  Layers3Icon,
  ShieldIcon,
  UsersIcon,
} from "lucide-react";
import { useMemo } from "react";

import { DocumentLibraryPage } from "@/components/documents/document-library-page";
import { DocumentSummaryList } from "@/components/documents/document-summary-list";
import { ContentWidth } from "@/components/layout/content-width";
import { NamespaceEntitySubtitle } from "@/components/namespaces/namespace-entity-subtitle";
import { openQuickCreate } from "@/components/quick-create/open";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { AddButton, CreateButton } from "@/components/ui/create-button";
import { EmptyState } from "@/components/ui/empty-state";
import { ExternalLink } from "@/components/ui/external-link";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { PropertyList } from "@/components/ui/property-list";
import { Section } from "@/components/ui/section";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { UserAvatarCompact } from "@/components/ui/user-avatar";
import { useAuth } from "@/hooks/use-auth";
import { v1OrganizationsDocumentsGetOptions } from "@/lib/api/query-options";
import { zUserStatus } from "@/lib/api/schemas";
import { Action, can } from "@/lib/auth/permissions";
import type { DocumentLibrarySearch } from "@/lib/documents/library";
import { formatDate } from "@/lib/format-date";
import { internalPath } from "@/lib/internal-url";
import { sortOrganizationMembers } from "@/lib/organization-members";
import type { OrganizationWorkspaceData } from "@/lib/organization-workspace";
import {
  namespacePath,
  organizationDocumentsPath,
  settingsNamespaceNewPath,
  settingsOrganizationEditPath,
  settingsOrganizationPath,
} from "@/lib/paths";
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
  const hasOrgUpdatePermission = can(permissions, Action.OrganizationUpdate);
  const hasNamespaceCreatePermission = can(permissions, Action.NamespaceCreate);
  const hasDocumentCreatePermission = can(permissions, Action.DocumentCreate);
  const hasMembersManagePermission = can(
    permissions,
    Action.OrganizationMembersManage
  );
  const hasRoleManagePermission = can(permissions, Action.RoleManage);
  const documentCount = organization.document_count ?? 0;
  const { data: documentsPage, isLoading: isDocumentsLoading } = useQuery({
    ...v1OrganizationsDocumentsGetOptions({
      path: { organizationRef: organizationId },
      query: { page_size: 5, all: true },
    }),
    enabled: hasReadAccess,
  });
  const documents = documentsPage?.items ?? [];

  const sortedMembers = useMemo(() => {
    return sortOrganizationMembers(members);
  }, [members]);

  const sortedNamespaces = useMemo(() => {
    return [...namespaces].sort((a, b) => a.name.localeCompare(b.name));
  }, [namespaces]);

  const sortedRoles = useMemo(() => {
    return [...roles].sort((a, b) => a.name.localeCompare(b.name));
  }, [roles]);

  const settingsHref = settingsOrganizationPath({
    organizationSlug: organization.slug,
  });

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
              ...(hasOrgUpdatePermission
                ? [
                    {
                      label: "Edit organization",
                      href: settingsOrganizationEditPath({
                        organizationSlug: organization.slug,
                      }),
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
                        href={namespacePath({
                          organizationSlug: organization.slug,
                          namespaceSlug: namespace.slug,
                        })}
                        title={namespace.name}
                        subtitle={
                          <NamespaceEntitySubtitle
                            description={namespace.description}
                            projectCount={namespace.project_count ?? 0}
                            documentCount={namespace.document_count ?? 0}
                          />
                        }
                      />
                    ))}
                  </AppList>
                ) : (
                  <EmptyState
                    compact
                    icon={<Layers3Icon />}
                    title="No namespaces"
                    description="Namespaces for this organization will appear here."
                    action={
                      hasNamespaceCreatePermission ? (
                        <Button
                          variant="outline"
                          render={
                            <InternalLink
                              to={internalPath(
                                settingsNamespaceNewPath({
                                  organizationSlug: organization.slug,
                                })
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
                title="Documents"
                data-section="documents"
                description={
                  documentCount
                    ? `${documentCount} ${pluralize(documentCount, "document", "documents")}`
                    : undefined
                }
                action={
                  <div className="flex items-center gap-2">
                    {hasDocumentCreatePermission ? (
                      <AddButton
                        size="sm"
                        onClick={() => openQuickCreate("document")}
                      />
                    ) : null}
                    <Button
                      variant="ghost"
                      size="sm"
                      render={
                        <InternalLink
                          to={internalPath(
                            organizationDocumentsPath({
                              organizationSlug: organization.slug,
                            })
                          )}
                        />
                      }
                    >
                      View all <ArrowRightIcon />
                    </Button>
                  </div>
                }
              >
                {isDocumentsLoading ? (
                  <ListSkeleton rows={4} />
                ) : documents.length > 0 ? (
                  <DocumentSummaryList documents={documents} />
                ) : (
                  <EmptyState
                    compact
                    icon={<FileTextIcon />}
                    title="No documents"
                    description="Documents that live in this organization will appear here."
                    action={
                      hasDocumentCreatePermission ? (
                        <CreateButton
                          onClick={() => openQuickCreate("document")}
                        />
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
                  <EmptyState
                    compact
                    icon={<UsersIcon />}
                    title="No members"
                    description="Members will appear here once they join this organization."
                    action={
                      hasMembersManagePermission ? (
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
                              `${role.member_count ?? 0} ${pluralize(
                                role.member_count ?? 0,
                                "member",
                                "members"
                              )}`}
                          </p>
                        </div>
                      </div>
                    ))}
                  </AppList>
                ) : (
                  <EmptyState
                    compact
                    icon={<ShieldIcon />}
                    title="No roles"
                    description="Roles organize permissions and member access."
                    action={
                      hasRoleManagePermission ? (
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
            <EmptyState
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
                {
                  label: "Documents",
                  value: documentCount,
                },
              ]}
            />
          </Section>
        </div>
      </div>
    </ContentWidth>
  );
}

export function OrganizationDocumentsPage({
  data,
  search,
}: {
  data: OrganizationWorkspaceData;
  search: DocumentLibrarySearch;
}) {
  const { organization, permissions, hasReadAccess } = data;
  const hasDocumentWritePermission =
    can(permissions, Action.DocumentCreate) ||
    can(permissions, Action.FolderCreate);

  return (
    <DocumentLibraryPage
      kind="organization"
      libraryId={organization.id}
      organizationId={organization.id}
      organizationSlug={organization.slug}
      libraryName={organization.name}
      search={search}
      mayWrite={hasDocumentWritePermission}
      hasReadAccess={hasReadAccess}
      documentCount={organization.document_count ?? 0}
      limitedAccessTitle="Limited access"
      limitedAccessDescription="Documents require read permission on this organization."
    />
  );
}
