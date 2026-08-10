import {
  ArrowRightIcon,
  FileTextIcon,
  FolderKanbanIcon,
  SearchIcon,
} from "lucide-react";
import { useMemo, useState } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { CreateButton } from "@/components/shared/create-button";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Section } from "@/components/shared/section";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { InternalLink } from "@/components/ui/internal-link";
import { CompactWorkList } from "@/components/work/work-list";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Namespace, Organization } from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";
import { internalPath } from "@/lib/internal-url";
import {
  selectAttentionSignals,
  selectDocumentBodies,
  selectWorkItems,
} from "@/lib/mock-data";

export function NamespaceOverviewPage({
  namespace,
  organization,
}: {
  namespace: Namespace;
  organization?: Organization;
}) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );
  const mayWrite = can(permissions, "write");
  const scopedWork = selectWorkItems({
    scope: { type: "namespace", namespaceId: namespace.id },
  });
  const work = scopedWork;
  const attention = selectAttentionSignals({
    scope: { type: "namespace", namespaceId: namespace.id },
  });
  const fixtureDocuments = selectDocumentBodies({
    type: "namespace",
    namespaceId: namespace.id,
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
              mayWrite ? (
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
              {
                label: "Create project",
                href: organization
                  ? `/settings/organizations/${organization.id}/namespaces/${namespace.id}/projects/new`
                  : undefined,
              },
              {
                label: "View relationships",
                href: `/relations/namespace/${namespace.id}`,
              },
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
                    to={internalPath(`/namespaces/${namespace.id}/projects`)}
                  />
                }
              >
                View all <ArrowRightIcon />
              </Button>
            }
          >
            {namespace.projects.length > 0 ? (
              <AppList>
                {namespace.projects.slice(0, 6).map((project) => (
                  <EntityLink
                    key={project.id}
                    type="project"
                    href={`/namespaces/${namespace.id}/projects/${project.id}`}
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
              <AppEmptyState
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

          <Section title="Recent documents">
            {namespace.documents.length > 0 ? (
              <AppList>
                {namespace.documents.slice(0, 5).map((document) => (
                  <EntityLink
                    key={document.id}
                    type="document"
                    href={`/documents/${document.id}`}
                    title={document.name}
                    subtitle={document.excerpt || "No summary"}
                  />
                ))}
              </AppList>
            ) : fixtureDocuments.length > 0 ? (
              <>
                <MockDataAlert
                  title="Illustrative document summaries"
                  className="mb-3"
                >
                  This fallback list uses centralized document fixtures.
                </MockDataAlert>
                {fixtureDocuments.map((document) => (
                  <EntityLink
                    key={document.documentId}
                    type="document"
                    href={`/documents/${document.documentId}`}
                    title={document.title}
                    subtitle={document.excerpt}
                  />
                ))}
              </>
            ) : (
              <AppEmptyState
                compact
                icon={<FileTextIcon />}
                title="No documents"
                description="Decisions, specifications, and knowledge will appear here."
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
  namespace: Namespace;
  organization?: Organization;
}) {
  const [query, setQuery] = useState("");
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespace.id)
  );
  const mayWrite = can(permissions, "write");
  const projects = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return namespace.projects;
    return namespace.projects.filter((project) =>
      [project.name, project.key, project.description]
        .filter(Boolean)
        .join(" ")
        .toLowerCase()
        .includes(normalized)
    );
  }, [namespace.projects, query]);

  return (
    <ContentWidth width="overview" className="space-y-6">
      <EntityHeader
        type="project"
        eyebrow={namespace.name}
        title="Projects"
        description={`Real undertakings in ${namespace.name}.`}
        showIcon={false}
        actions={
          mayWrite && organization ? (
            <CreateButton
              href={`/settings/organizations/${organization.id}/namespaces/${namespace.id}/projects/new`}
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
      {projects.length > 0 ? (
        <AppList>
          {projects.map((project, index) => (
            <InternalLink
              key={project.id}
              to={internalPath(
                `/namespaces/${namespace.id}/projects/${project.id}`
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
                {index + 1} of {namespace.projects.length}
              </div>
            </InternalLink>
          ))}
        </AppList>
      ) : (
        <AppEmptyState
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
}: {
  namespace: Namespace;
}) {
  const [query, setQuery] = useState("");
  const documents = namespace.documents.filter((document) =>
    [document.name, document.excerpt]
      .filter(Boolean)
      .join(" ")
      .toLowerCase()
      .includes(query.trim().toLowerCase())
  );

  return (
    <ContentWidth width="overview" className="space-y-6">
      <EntityHeader
        type="document"
        eyebrow={namespace.name}
        title="Documents"
        description="Decisions, specifications, and knowledge connected to this namespace."
        showIcon={false}
      />
      <div className="bg-background sticky top-0 z-10 py-3">
        <div className="relative">
          <SearchIcon className="text-muted-foreground absolute top-2.5 left-3 size-4" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search documents..."
            className="pl-9"
          />
        </div>
      </div>
      {documents.length > 0 ? (
        <AppList>
          {documents.map((document) => (
            <EntityLink
              key={document.id}
              type="document"
              href={`/documents/${document.id}`}
              title={document.name}
              subtitle={
                <span>
                  {document.excerpt || "No summary"} ·{" "}
                  {document.created_at
                    ? new Intl.DateTimeFormat(undefined, {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      }).format(new Date(document.created_at))
                    : "Date unavailable"}
                </span>
              }
              className="px-4 py-3"
            />
          ))}
        </AppList>
      ) : (
        <AppEmptyState
          icon={<FileTextIcon />}
          title={query ? "No document matches" : "No documents yet"}
          description={
            query
              ? "Try a different title or summary."
              : "Documents hold decisions and knowledge without requiring folders or projects."
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
