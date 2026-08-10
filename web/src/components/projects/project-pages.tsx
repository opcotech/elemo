import {
  ActivityIcon,
  ArrowRightIcon,
  FileTextIcon,
  SearchIcon,
} from "lucide-react";
import { useMemo, useState } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
import { ActivityFeed } from "@/components/shared/activity-feed";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { CreateButton } from "@/components/shared/create-button";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { PropertyList } from "@/components/shared/property-list";
import { RelationList } from "@/components/shared/relation-list";
import { Section } from "@/components/shared/section";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { InternalLink } from "@/components/ui/internal-link";
import { Progress } from "@/components/ui/progress";
import { CompactWorkList } from "@/components/work/work-list";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import type { Namespace, Project } from "@/lib/api/types";
import { can } from "@/lib/auth/permissions";
import { internalPath } from "@/lib/internal-url";
import {
  selectActivity,
  selectRelations,
  selectWorkItems,
} from "@/lib/mock-data";

export function ProjectOverviewPage({
  namespace,
  project,
}: {
  namespace: Namespace;
  project: Project;
}) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Project, project.id)
  );
  const mayWrite = can(permissions, "write");
  const scopedWork = selectWorkItems({
    scope: {
      type: "project",
      namespaceId: namespace.id,
      projectId: project.id,
    },
  });
  const work = scopedWork;
  const projectRelations = selectRelations({
    entity: { id: work[0]?.id ?? "", type: "work-item" },
  });
  const progress = project.issues.length
    ? Math.min(92, 28 + project.issues.length * 9)
    : 0;

  return (
    <ContentWidth width="overview" className="space-y-7">
      <EntityHeader
        type="project"
        eyebrow={`${namespace.name} / ${project.key}`}
        title={project.name}
        description={
          project.description || "No project outcome has been described yet."
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
                <Button disabled size="sm" title="Write permission required">
                  Create work
                </Button>
              )
            }
            secondary={[
              {
                label: "Open project work",
                href: `/namespaces/${namespace.id}/projects/${project.id}/work`,
              },
              {
                label: "View relationships",
                href: `/relations/project/${project.id}`,
              },
            ]}
          />
        }
      />

      <div className="grid gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <div className="space-y-8">
          <Section title="Current state">
            <div className="rounded-xl border p-4">
              <div className="flex items-center gap-3">
                <Progress value={progress} className="flex-1" />
                <span className="text-muted-foreground text-sm tabular-nums">
                  {project.issues.length
                    ? `${progress}% illustrative`
                    : "No issue progress"}
                </span>
              </div>
              <p className="text-muted-foreground mt-3 text-sm">
                {project.issues.length
                  ? `${project.issues.length} linked issues are associated with this project.`
                  : "No linked issues are associated with this project yet."}
              </p>
              {project.issues.length > 0 && (
                <MockDataAlert title="Illustrative progress" className="mt-3">
                  Progress is estimated from linked issues and does not reflect
                  verified completion.
                </MockDataAlert>
              )}
            </div>
          </Section>

          <Section
            title="Active work"
            action={
              <Button
                variant="ghost"
                size="sm"
                render={
                  <InternalLink
                    to={internalPath(
                      `/namespaces/${namespace.id}/projects/${project.id}/work`
                    )}
                  />
                }
              >
                View all <ArrowRightIcon />
              </Button>
            }
          >
            <MockDataAlert title="Illustrative project work" className="mb-3">
              {scopedWork.length
                ? "Work detail is provided by centralized fixtures."
                : "No fixture work matches this project. Unrelated examples are not shown."}
            </MockDataAlert>
            <CompactWorkList items={work} limit={5} />
          </Section>

          <Section
            title="Documents"
            action={
              <Button
                variant="ghost"
                size="sm"
                render={
                  <InternalLink
                    to={internalPath(
                      `/namespaces/${namespace.id}/projects/${project.id}/documents`
                    )}
                  />
                }
              >
                View all <ArrowRightIcon />
              </Button>
            }
          >
            {project.documents.length > 0 ? (
              <AppList>
                {project.documents.slice(0, 5).map((document) => (
                  <EntityLink
                    key={document.id}
                    type="document"
                    href={`/documents/${document.id}`}
                    title={document.name}
                    subtitle={document.excerpt || "No summary"}
                  />
                ))}
              </AppList>
            ) : (
              <AppEmptyState
                compact
                icon={<FileTextIcon />}
                title="No project documents"
                description="Documents can explain outcomes and decisions without becoming folders."
              />
            )}
          </Section>
        </div>

        <div className="space-y-8">
          <Section title="Details">
            <PropertyList
              items={[
                { label: "Key", value: project.key },
                {
                  label: "State",
                  value: <StatusIndicator status={project.status} />,
                },
                {
                  label: "Namespace",
                  value: (
                    <InternalLink
                      to={internalPath(`/namespaces/${namespace.id}`)}
                      className="text-primary hover:underline"
                    >
                      {namespace.name}
                    </InternalLink>
                  ),
                },
                {
                  label: "Issues",
                  value: `${project.issues.length} embedded IDs`,
                },
                {
                  label: "Documents",
                  value: project.documents.length,
                },
              ]}
            />
          </Section>
          <Section title="Relations">
            <MockDataAlert title="Illustrative relationships" className="mb-3">
              Project relationships are illustrative examples linked to the work
              shown above.
            </MockDataAlert>
            <RelationList
              relations={projectRelations}
              entity={{
                id: work[0]?.id ?? project.id,
                type: "work-item",
              }}
            />
          </Section>
        </div>
      </div>
    </ContentWidth>
  );
}

export function ProjectDocumentsPage({
  namespace,
  project,
}: {
  namespace: Namespace;
  project: Project;
}) {
  const [query, setQuery] = useState("");
  const documents = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return project.documents;
    return project.documents.filter((document) =>
      [document.name, document.excerpt]
        .filter(Boolean)
        .join(" ")
        .toLowerCase()
        .includes(normalized)
    );
  }, [project.documents, query]);

  return (
    <ContentWidth width="overview" className="space-y-6">
      <EntityHeader
        type="document"
        eyebrow={`${namespace.name} / ${project.name}`}
        title="Documents"
        description="Documents linked to this project."
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
                  {document.excerpt || "No summary"}
                  {document.created_at
                    ? ` · ${new Intl.DateTimeFormat(undefined, {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      }).format(new Date(document.created_at))}`
                    : ""}
                </span>
              }
              className="px-4 py-3"
            />
          ))}
        </AppList>
      ) : (
        <AppEmptyState
          icon={<FileTextIcon />}
          title={query ? "No document matches" : "No project documents"}
          description={
            query
              ? "Try a different title or summary."
              : "Documents added to this project will appear here."
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

export function ProjectActivityPage({
  namespace,
  project,
}: {
  namespace: Namespace;
  project: Project;
}) {
  const activity = selectActivity({ limit: 20 });
  return (
    <ContentWidth width="overview" className="space-y-6">
      <EntityHeader
        type="project"
        eyebrow={`${namespace.name} / ${project.name}`}
        title="Activity"
        description="A bounded history of meaningful project changes."
        showIcon={false}
      />
      <MockDataAlert title="Illustrative activity">
        Activity entries below are illustrative examples and may not reflect
        this project’s live history.
      </MockDataAlert>
      {activity.length > 0 ? (
        <ActivityFeed entries={activity} />
      ) : (
        <AppEmptyState
          icon={<ActivityIcon />}
          title="No activity"
          description="Meaningful changes will appear here as activity becomes available."
        />
      )}
    </ContentWidth>
  );
}
