import { useQuery } from "@tanstack/react-query";
import {
  ArrowRightIcon,
  CheckCircle2Icon,
  CircleAlertIcon,
  ListChecksIcon,
  SparklesIcon,
} from "lucide-react";
import { useMemo } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { AttentionIcon } from "@/components/shared/attention-icon";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { groupTodosByDueDate } from "@/components/todo/grouping";
import { TodoPriorityRibbon } from "@/components/todo/todo-priority-ribbon";
import { Button } from "@/components/ui/button";
import { CreateButton } from "@/components/ui/create-button";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { PageHeader } from "@/components/ui/page-header";
import { Section } from "@/components/ui/section";
import { MarkdownContent } from "@/components/work/markdown-content";
import { workItemPath } from "@/components/work/utils";
import { CompactWorkList } from "@/components/work/work-list";
import { useAuth } from "@/hooks/use-auth";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1TodosGetOptions,
  v1UsersIssuesGetOptions,
} from "@/lib/api/query-options";
import { v1UsersIssuesGet } from "@/lib/api/sdk";
import type { Todo } from "@/lib/api/types";
import { internalPath } from "@/lib/internal-url";
import { resolveDemoPerson, selectAttentionSignals } from "@/lib/mock-data";
import { namespacePath } from "@/lib/paths";
import { recentEntityLinkType } from "@/lib/recent-entity";
import { uiActions, useUiSelector } from "@/lib/ui-store";
import { issuesToWorkItemsWithNamespaces } from "@/lib/work/issue-adapter";
import { queryWorkItems } from "@/lib/work/query";

const HOME_TODO_PREVIEW_LIMIT = 5;

export function HomePage() {
  const { user } = useAuth();
  const userId = user?.id;
  const demoPerson = resolveDemoPerson(user);
  const { data: accessibleWorkspace, isLoading: namespacesLoading } =
    useAccessibleNamespaces();
  const namespaces = accessibleWorkspace?.namespaces ?? [];
  const { data: todosPage, isLoading: todosLoading } =
    useQuery(v1TodosGetOptions());
  const userIssuesOptions = v1UsersIssuesGetOptions({
    path: { id: userId ?? "" },
    query: cursorPageQuery(),
  });
  const {
    data: issuesPage,
    error: issuesError,
    isLoading: issuesLoading,
  } = useQuery({
    ...collectedListQuery(userIssuesOptions, async (pageToken, signal) => {
      const { data } = await v1UsersIssuesGet({
        path: { id: userId ?? "" },
        query: cursorPageQuery(pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    }),
    enabled: Boolean(userId),
  });
  const todos: Todo[] = todosPage?.items ?? [];
  const userWorkItems = useMemo(
    () => issuesToWorkItemsWithNamespaces(issuesPage?.items ?? [], namespaces),
    [issuesPage?.items, namespaces]
  );
  const userWorkItemById = useMemo(
    () =>
      new Map(
        userWorkItems.flatMap((item) => [
          [item.id, item] as const,
          [item.key, item] as const,
        ])
      ),
    [userWorkItems]
  );
  const recentEntities = useUiSelector((state) => state.recentEntities);
  const attention = useMemo(() => {
    const entries: {
      signal: ReturnType<typeof selectAttentionSignals>[number];
      item: (typeof userWorkItems)[number];
    }[] = [];
    for (const signal of selectAttentionSignals({ personId: demoPerson.id })) {
      const item = userWorkItemById.get(signal.workItemId);
      if (!item) {
        continue;
      }
      entries.push({ signal, item });
    }
    return entries;
  }, [demoPerson.id, userWorkItemById]);
  const continueWorking = useMemo(
    () =>
      queryWorkItems(userWorkItems, {
        filters: { statuses: ["in progress"] },
        sort: [{ field: "priority", direction: "desc" }],
      }),
    [userWorkItems]
  );
  const openTodoGroups = useMemo(() => {
    const openTodos = todos.filter((todo) => !todo.completed);
    const groups = groupTodosByDueDate(openTodos);
    let remaining = HOME_TODO_PREVIEW_LIMIT;

    return groups.flatMap((group) => {
      if (remaining <= 0) {
        return [];
      }

      const previewTodos = group.todos.slice(0, remaining);
      remaining -= previewTodos.length;

      return [{ ...group, todos: previewTodos }];
    });
  }, [todos]);
  const hasOpenTodos = openTodoGroups.length > 0;

  return (
    <ContentWidth width="overview" className="space-y-8">
      <PageHeader
        title="Home"
        description="Your work across Elemo"
        actions={
          <CreateButton onClick={() => openQuickCreate()}>Create</CreateButton>
        }
      />

      <div className="grid gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <div className="space-y-8">
          <Section
            title="Needs attention"
            action={
              <Button
                variant="ghost"
                size="sm"
                render={<InternalLink to="/my-work" />}
              >
                View My Work <ArrowRightIcon />
              </Button>
            }
          >
            <MockDataAlert
              title="Illustrative attention signals"
              className="mb-3"
            >
              Attention reasons and work details are centralized fixtures. Real
              notifications remain in Inbox and are not duplicated here.
            </MockDataAlert>
            {attention.length > 0 ? (
              <AppList>
                {attention.slice(0, 5).map(({ signal, item }) => {
                  return (
                    <InternalLink
                      key={signal.id}
                      to={internalPath(workItemPath(item))}
                      className="hover:bg-muted/50 flex items-center gap-3 px-3 py-2.5"
                    >
                      <AttentionIcon severity={signal.severity} />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-medium">
                          {`${item.key} ${item.title}`}
                        </span>
                        <span className="text-muted-foreground block truncate text-xs">
                          {signal.summary}
                        </span>
                      </span>
                      <span className="text-muted-foreground text-xs capitalize">
                        {signal.reason.replaceAll("-", " ")}
                      </span>
                    </InternalLink>
                  );
                })}
              </AppList>
            ) : (
              <EmptyState
                compact
                icon={<CheckCircle2Icon />}
                title="You’re clear"
                description="There are no unacknowledged fixture attention signals."
              />
            )}
          </Section>

          <Section
            title="Continue working"
            data-section="home-continue-working"
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
            {issuesLoading ? (
              <ListSkeleton rows={6} />
            ) : issuesError ? (
              <EmptyState
                compact
                icon={<CircleAlertIcon />}
                title="Couldn't load work"
                description="Assigned in-progress work could not be loaded. Try again later."
              />
            ) : (
              <CompactWorkList
                items={continueWorking}
                limit={6}
                showAssignees={false}
                showLabels={false}
                emptyTitle="No in-progress work"
                emptyDescription="Assigned work that is in progress will appear here."
              />
            )}
          </Section>

          <Section
            title="Personal todos"
            data-section="home-personal-todos"
            description="Your personal todos"
            action={
              hasOpenTodos ? (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={uiActions.openTodoSheet}
                >
                  View all <ArrowRightIcon />
                </Button>
              ) : undefined
            }
          >
            {todosLoading ? (
              <ListSkeleton rows={4} />
            ) : hasOpenTodos ? (
              <div className="flex flex-col gap-5">
                {openTodoGroups.map((group) => (
                  <section key={group.id} className="min-w-0">
                    <h3 className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                      {group.label}
                    </h3>
                    <AppList aria-label={`${group.label} todos`}>
                      {group.todos.map((todo) => (
                        <button
                          type="button"
                          key={todo.id}
                          className="hover:bg-muted/50 flex w-full items-center gap-3 px-3 py-2.5 text-left"
                          onClick={uiActions.openTodoSheet}
                        >
                          <ListChecksIcon className="text-muted-foreground size-4" />
                          <span className="min-w-0 flex-1">
                            <span className="block truncate text-sm font-medium">
                              {todo.title}
                            </span>
                            <MarkdownContent
                              markdown={todo.description}
                              size="xs"
                              className="block truncate"
                              empty={
                                <span className="text-muted-foreground block truncate text-xs">
                                  No description
                                </span>
                              }
                            />
                          </span>
                          <TodoPriorityRibbon
                            priority={todo.priority}
                            className="gap-1"
                            iconClassName="size-3"
                            labelClassName="text-xs font-medium"
                          />
                        </button>
                      ))}
                    </AppList>
                  </section>
                ))}
              </div>
            ) : (
              <EmptyState
                compact
                icon={<ListChecksIcon />}
                title="No open todos"
                description="Use Quick Create to add a real personal todo."
                action={
                  <Button variant="outline" onClick={() => openQuickCreate()}>
                    Add todo
                  </Button>
                }
              />
            )}
          </Section>
        </div>

        <div className="space-y-8">
          <Section title="Recent context">
            {namespacesLoading ? (
              <ListSkeleton rows={4} />
            ) : namespaces.length > 0 ? (
              <AppList>
                {namespaces.slice(0, 5).map((namespace) => (
                  <EntityLink
                    key={namespace.id}
                    type="namespace"
                    href={namespacePath({
                      organizationSlug: namespace.organizationSlug,
                      namespaceSlug: namespace.slug,
                    })}
                    title={namespace.name}
                    subtitle={namespace.organizationName}
                  />
                ))}
              </AppList>
            ) : (
              <EmptyState
                compact
                icon={<SparklesIcon />}
                title="No namespaces"
                description="Join or create a namespace from Settings to establish an operational context."
                action={
                  <Button
                    variant="outline"
                    render={<InternalLink to="/namespaces" />}
                  >
                    Browse namespaces
                  </Button>
                }
              />
            )}
          </Section>

          <Section title="Recent entities">
            {recentEntities.length > 0 ? (
              <AppList>
                {recentEntities.slice(0, 6).map((entity) => (
                  <EntityLink
                    key={`${entity.type}:${entity.id}`}
                    type={recentEntityLinkType(entity.type)}
                    href={entity.href}
                    title={entity.label}
                  />
                ))}
              </AppList>
            ) : (
              <p className="text-muted-foreground rounded-lg border p-4 text-sm">
                Open a namespace, project, work item, or document to build your
                recent history.
              </p>
            )}
          </Section>
        </div>
      </div>
    </ContentWidth>
  );
}
