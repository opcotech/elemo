import { useQuery } from "@tanstack/react-query";
import {
  ArrowRightIcon,
  CheckCircle2Icon,
  ListChecksIcon,
  SparklesIcon,
} from "lucide-react";
import { useMemo } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { AttentionIcon } from "@/components/shared/attention-icon";
import { CreateButton } from "@/components/shared/create-button";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { PageHeader } from "@/components/shared/page-header";
import { Section } from "@/components/shared/section";
import { StatusIndicator } from "@/components/shared/status-indicator";
import { groupTodosByDueDate } from "@/components/todo/grouping";
import { todoPriorityTone } from "@/components/todo/priority";
import { Button } from "@/components/ui/button";
import { InternalLink } from "@/components/ui/internal-link";
import { CompactWorkList } from "@/components/work/work-list";
import { useAuth } from "@/hooks/use-auth";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import type { Todo } from "@/lib/api/types";
import {
  getWorkItem,
  resolveDemoPerson,
  selectAttentionSignals,
  selectWorkItems,
} from "@/lib/mock-data";
import { recentEntityLinkType } from "@/lib/recent-entity";
import { uiActions, useUiSelector } from "@/lib/ui-store";

const HOME_TODO_PREVIEW_LIMIT = 5;

export function HomePage() {
  const { user } = useAuth();
  const demoPerson = resolveDemoPerson(user);
  const { data: accessibleWorkspace } = useAccessibleNamespaces();
  const namespaces = accessibleWorkspace?.namespaces ?? [];
  const { data: todosData } = useQuery(v1TodosGetOptions());
  const todos: Todo[] = todosData ?? [];
  const recentEntities = useUiSelector((state) => state.recentEntities);
  const attention = selectAttentionSignals({
    personId: demoPerson.id,
  });
  const work = selectWorkItems({
    scope: { type: "person", personId: demoPerson.id },
    filters: {
      statuses: ["planned", "in-progress", "blocked"],
    },
    sort: [{ field: "updatedAt", direction: "desc" }],
  });
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
                {attention.slice(0, 5).map((signal) => {
                  const item = getWorkItem(signal.workItemId);
                  return (
                    <InternalLink
                      key={signal.id}
                      to={item ? (`/work/${item.id}` as const) : "/my-work"}
                      className="hover:bg-muted/50 flex items-center gap-3 px-3 py-2.5"
                    >
                      <AttentionIcon severity={signal.severity} />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-medium">
                          {item ? `${item.key} ${item.title}` : signal.summary}
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
              <AppEmptyState
                compact
                icon={<CheckCircle2Icon />}
                title="You’re clear"
                description="There are no unacknowledged fixture attention signals."
              />
            )}
          </Section>

          <Section
            title="Continue working"
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
            <MockDataAlert title="Illustrative work items" className="mb-3">
              Work items shown here are illustrative examples until live work
              tracking is available in this view.
            </MockDataAlert>
            <CompactWorkList items={work} limit={6} />
          </Section>

          <Section
            title="Personal todos"
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
            {hasOpenTodos ? (
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
                            <span className="text-muted-foreground block truncate text-xs">
                              {todo.description || "No description"}
                            </span>
                          </span>
                          <StatusIndicator
                            status={todo.priority}
                            tone={todoPriorityTone[todo.priority]}
                          />
                        </button>
                      ))}
                    </AppList>
                  </section>
                ))}
              </div>
            ) : (
              <AppEmptyState
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
            {namespaces.length > 0 ? (
              <AppList>
                {namespaces.slice(0, 5).map((namespace) => (
                  <EntityLink
                    key={namespace.id}
                    type="namespace"
                    href={`/namespaces/${namespace.id}`}
                    title={namespace.name}
                    subtitle={namespace.organizationName}
                  />
                ))}
              </AppList>
            ) : (
              <AppEmptyState
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
