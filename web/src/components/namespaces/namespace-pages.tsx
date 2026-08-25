import { useQuery } from "@tanstack/react-query";
import {
  ArrowRightIcon,
  FileTextIcon,
  FolderKanbanIcon,
  SearchIcon,
} from "lucide-react";
import { useMemo, useState } from "react";

import { DocumentLibraryPage } from "@/components/documents/document-library-page";
import { DocumentSummaryList } from "@/components/documents/document-summary-list";
import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
import { MockDataAlert } from "@/components/shared/app-feedback";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import { AddButton, CreateButton } from "@/components/ui/create-button";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { Section } from "@/components/ui/section";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { CompactWorkList } from "@/components/work/work-list";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1NamespacesDocumentsGetOptions,
  v1NamespacesProjectsGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
import type {
  AccessibleNamespace,
  PartialDocument,
  Project,
} from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";
import type { DocumentLibrarySearch } from "@/lib/documents/library";
import { internalPath } from "@/lib/internal-url";
import { selectAttentionSignals, selectWorkItems } from "@/lib/mock-data";
import {
  namespaceDocumentsPath,
  namespaceProjectsPath,
  projectPath,
  settingsProjectNewPath,
} from "@/lib/paths";

export function NamespaceOverviewPage({
  namespace,
  organization,
}: {
  namespace: AccessibleNamespace;
  organization: AccessibleNamespace["organization"];
}) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );
  const { data: projectsPage, isLoading: isProjectsLoading } = useQuery(
    v1NamespacesProjectsGetOptions({
      path: namespaceRefPath(organization.id, namespace.id),
      query: { page_size: 6 },
    })
  );
  const { data: documentsPage, isLoading: isDocumentsLoading } = useQuery(
    v1NamespacesDocumentsGetOptions({
      path: namespaceRefPath(organization.id, namespace.id),
      query: { page_size: 5, all: true },
    })
  );
  const projects: Project[] = projectsPage?.items ?? [];
  const documents: PartialDocument[] = documentsPage?.items ?? [];
  const mayCreateWork = can(permissions, Action.IssueCreate);
  const mayCreateDocument = can(permissions, Action.DocumentCreate);
  const mayCreateProject = can(permissions, Action.ProjectCreate);
  const mayReadDocuments = can(permissions, Action.DocumentRead);
  const mayAdminister = can(permissions, Action.NamespaceRead);
  const scopedWork = selectWorkItems({
    scope: { type: "namespace", namespaceId: namespace.id },
  });
  const work = scopedWork;
  const attention = selectAttentionSignals({
    scope: { type: "namespace", namespaceId: namespace.id },
  });

  return (
    <ContentWidth width="overview" className="space-y-7">
      <EntityHeader
        type="namespace"
        eyebrow={organization?.name ?? "Namespace"}
        title={namespace.name}
        description={
          namespace.description ||
          "An operational context for projects, work, and knowledge."
        }
        showIcon={false}
        actions={
          <PageActions
            primary={
              mayCreateWork ? (
                <CreateButton onClick={() => openQuickCreate("work")}>
                  Work
                </CreateButton>
              ) : (
                <Button size="sm" disabled title="Write permission required">
                  Create work
                </Button>
              )
            }
            secondary={[
              ...(mayCreateProject
                ? [
                    {
                      label: "Create project",
                      href: settingsProjectNewPath({
                        organizationSlug: organization.slug,
                        namespaceSlug: namespace.slug,
                      }),
                    },
                  ]
                : []),
              ...(mayAdminister
                ? [
                    {
                      label: "View relationships",
                      href: `/relations/namespace/${namespace.id}`,
                    },
                  ]
                : []),
            ]}
          />
        }
      />

      <div className="grid gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <div className="space-y-8">
          <Section
            title="Projects"
            action={
              <Button
                variant="ghost"
                size="sm"
                render={
                  <InternalLink
                    to={internalPath(
                      namespaceProjectsPath({
                        organizationSlug: organization.slug,
                        namespaceSlug: namespace.slug,
                      })
                    )}
                  />
                }
              >
                View all <ArrowRightIcon />
              </Button>
            }
          >
            {isProjectsLoading ? (
              <ListSkeleton rows={4} />
            ) : projects.length > 0 ? (
              <AppList>
                {projects.map((project) => (
                  <EntityLink
                    key={project.id}
                    type="project"
                    href={projectPath({
                      organizationSlug: organization.slug,
                      namespaceSlug: namespace.slug,
                      projectKey: project.key,
                    })}
                    title={project.name}
                    imageUrl={project.logo}
                    subtitle={
                      <span className="flex items-center gap-2">
                        <StatusIndicator status={project.status} />
                        <span className="truncate">
                          {project.description || project.key}
                        </span>
                      </span>
                    }
                  />
                ))}
              </AppList>
            ) : (
              <EmptyState
                compact
                icon={<FolderKanbanIcon />}
                title="No projects yet"
                description="Projects are optional. Work can exist directly in this namespace."
              />
            )}
          </Section>

          <Section
            title="Your work"
            action={
              <Button
                variant="ghost"
                size="sm"
                render={<InternalLink to="/my-work" />}
              >
                View all <ArrowRightIcon />
              </Button>
            }
          >
            <MockDataAlert title="Illustrative namespace work" className="mb-3">
              {scopedWork.length
                ? "This work region is powered by centralized fixtures."
                : "No fixture work matches this namespace. Unrelated examples are not shown."}
            </MockDataAlert>
            <CompactWorkList items={work} limit={5} />
          </Section>
        </div>

        <div className="space-y-8">
          <Section title="Needs attention">
            <MockDataAlert
              title="Illustrative attention signals"
              className="mb-3"
            >
              Attention and risk signals are illustrative. Namespace and project
              details elsewhere on this page reflect live workspace data.
            </MockDataAlert>
            {attention.length > 0 ? (
              <div className="space-y-2">
                {attention.slice(0, 4).map((signal) => (
                  <div
                    key={signal.id}
                    className="rounded-lg border px-3 py-2.5"
                  >
                    <p className="text-sm font-medium">{signal.summary}</p>
                    <p className="text-muted-foreground mt-1 text-xs capitalize">
                      {signal.reason.replaceAll("-", " ")}
                    </p>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground rounded-lg border p-4 text-sm">
                No fixture attention signals map to this namespace.
              </p>
            )}
          </Section>

          <Section
            title="Recent documents"
            data-section="documents"
            action={
              <div className="flex items-center gap-2">
                {mayCreateDocument ? (
                  <AddButton
                    size="sm"
                    onClick={() => openQuickCreate("document")}
                  />
                ) : null}
                {mayReadDocuments ? (
                  <Button
                    variant="ghost"
                    size="sm"
                    render={
                      <InternalLink
                        to={internalPath(
                          namespaceDocumentsPath({
                            organizationSlug: organization.slug,
                            namespaceSlug: namespace.slug,
                          })
                        )}
                      />
                    }
                  >
                    View all <ArrowRightIcon />
                  </Button>
                ) : null}
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
                description="Documents that live in this namespace will appear here."
                action={
                  mayCreateDocument ? (
                    <CreateButton onClick={() => openQuickCreate("document")} />
                  ) : undefined
                }
              />
            )}
          </Section>
        </div>
      </div>
    </ContentWidth>
  );
}

export function NamespaceProjectsPage({
  namespace,
  organization,
}: {
  namespace: AccessibleNamespace;
  organization: AccessibleNamespace["organization"];
}) {
  const [query, setQuery] = useState("");
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );
  const pageNav = useCursorPageNav({ resetKey: query });
  const { data: projectsPage, isLoading } = useQuery(
    v1NamespacesProjectsGetOptions({
      path: namespaceRefPath(organization.id, namespace.id),
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const mayCreateProject = can(permissions, Action.ProjectCreate);
  const projects = useMemo(() => {
    const items: Project[] = projectsPage?.items ?? [];
    const normalized = query.trim().toLowerCase();
    if (!normalized) return items;
    return items.filter((project) =>
      [project.name, project.key, project.description]
        .filter(Boolean)
        .join(" ")
        .toLowerCase()
        .includes(normalized)
    );
  }, [projectsPage?.items, query]);

  return (
    <ContentWidth width="overview" className="space-y-6">
      <EntityHeader
        type="project"
        eyebrow={namespace.name}
        title="Projects"
        description={`Real undertakings in ${namespace.name}.`}
        showIcon={false}
        actions={
          mayCreateProject ? (
            <CreateButton
              href={settingsProjectNewPath({
                organizationSlug: organization.slug,
                namespaceSlug: namespace.slug,
              })}
            >
              New project
            </CreateButton>
          ) : undefined
        }
      />
      <div className="bg-background sticky top-0 z-10 flex flex-wrap gap-2 py-3">
        <div className="relative min-w-60 flex-1">
          <SearchIcon className="text-muted-foreground absolute top-2.5 left-3 size-4" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search projects..."
            className="pl-9"
          />
        </div>
        <Button variant="outline">State</Button>
        <Button variant="outline">Sort: Updated</Button>
      </div>
      <MockDataAlert title="Limited project signals">
        Project name, status, and description are shown from your workspace.
        Progress, lead, target, and risk are unavailable for these projects
        right now.
      </MockDataAlert>
      {isLoading ? (
        <ListSkeleton />
      ) : projects.length > 0 ? (
        <>
          <AppList>
            {projects.map((project, index) => (
              <InternalLink
                key={project.id}
                to={internalPath(
                  projectPath({
                    organizationSlug: organization.slug,
                    namespaceSlug: namespace.slug,
                    projectKey: project.key,
                  })
                )}
                className="hover:bg-muted/40 focus-visible:ring-ring grid gap-3 px-4 py-4 outline-none focus-visible:ring-2 focus-visible:ring-inset sm:grid-cols-[minmax(0,1fr)_9rem_10rem]"
              >
                <div className="flex min-w-0 gap-3">
                  {project.logo ? (
                    <img
                      src={project.logo}
                      alt=""
                      className="bg-muted size-10 shrink-0 rounded-lg object-cover"
                    />
                  ) : (
                    <span className="bg-muted text-muted-foreground flex size-10 shrink-0 items-center justify-center rounded-lg">
                      <FolderKanbanIcon className="size-5" />
                    </span>
                  )}
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <h2 className="truncate font-medium">{project.name}</h2>
                      <span className="text-muted-foreground font-mono text-xs">
                        {project.key}
                      </span>
                    </div>
                    <p className="text-muted-foreground mt-1 line-clamp-2 text-sm">
                      {project.description ||
                        "No project outcome has been added."}
                    </p>
                  </div>
                </div>
                <div className="text-sm">
                  <span className="text-muted-foreground block text-xs">
                    State
                  </span>
                  <StatusIndicator status={project.status} />
                </div>
                <div className="text-sm">
                  <span className="text-muted-foreground block text-xs">
                    Position
                  </span>
                  {index + 1} of {projectsPage?.items.length ?? projects.length}
                </div>
              </InternalLink>
            ))}
          </AppList>
          <CursorPaginator {...cursorPaginatorProps(projectsPage, pageNav)} />
        </>
      ) : (
        <EmptyState
          icon={<FolderKanbanIcon />}
          title={query ? "No project matches" : "No projects yet"}
          description={
            query
              ? "Try a different name, key, or description."
              : "Projects represent coordinated outcomes; namespace work does not require one."
          }
          action={
            query ? (
              <Button variant="outline" onClick={() => setQuery("")}>
                Clear search
              </Button>
            ) : undefined
          }
        />
      )}
    </ContentWidth>
  );
}

export function NamespaceDocumentsPage({
  namespace,
  organization,
  search,
}: {
  namespace: AccessibleNamespace;
  organization: AccessibleNamespace["organization"];
  search: DocumentLibrarySearch;
}) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );
  const mayWrite =
    can(permissions, Action.DocumentCreate) ||
    can(permissions, Action.FolderCreate);

  return (
    <DocumentLibraryPage
      kind="namespace"
      libraryId={namespace.id}
      organizationId={organization.id}
      organizationSlug={organization.slug}
      namespaceId={namespace.id}
      namespaceSlug={namespace.slug}
      libraryName={namespace.name}
      search={search}
      mayWrite={mayWrite}
      documentCount={namespace.document_count ?? 0}
      limitedAccessTitle="Limited access"
      limitedAccessDescription="Documents require read permission on this namespace."
    />
  );
}
