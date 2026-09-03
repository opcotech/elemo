import { MessageSquareIcon } from "lucide-react";
import type { ReactNode } from "react";
import { useSyncExternalStore } from "react";

import { PluginErrorBoundary } from "@/components/plugins/plugin-slot";
import { EmptyState } from "@/components/ui/empty-state";
import { SectionAccordion } from "@/components/ui/section";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  getSlotContributions,
  subscribePluginRegistry,
} from "@/lib/plugins/registry";

const ACTIVITY_SLOT = "issue.activity" as const;
const HOST_TAB = "activity";

export function WorkItemActivity({
  context,
  children,
}: {
  context?: Record<string, unknown>;
  children?: ReactNode;
}) {
  const contributions = useSyncExternalStore(
    subscribePluginRegistry,
    () => getSlotContributions(ACTIVITY_SLOT),
    () => getSlotContributions(ACTIVITY_SLOT)
  );

  const hasPlugins = contributions.length > 0;
  const hostActivity = children ?? (
    <EmptyState
      compact
      icon={<MessageSquareIcon />}
      title="No activity yet"
      description="Issue activity is not available from the API yet."
    />
  );

  const tabCount = contributions.length + 1;

  return (
    <SectionAccordion
      key={hasPlugins ? "activity-extended" : "activity"}
      title="Activity"
      value="activity"
      defaultOpen={hasPlugins}
    >
      {hasPlugins ? (
        <Tabs defaultValue={HOST_TAB} className="w-full">
          <TabsList
            className="grid w-full"
            style={{
              gridTemplateColumns: `repeat(${tabCount}, minmax(0, 1fr))`,
            }}
          >
            <TabsTrigger value={HOST_TAB}>Activity</TabsTrigger>
            {contributions.map((item) => (
              <TabsTrigger
                key={`${item.pluginId}:${item.index}`}
                value={activityTabValue(item.pluginId, item.index)}
              >
                {item.title || item.pluginId}
              </TabsTrigger>
            ))}
          </TabsList>
          <TabsContent value={HOST_TAB} className="mt-3">
            {hostActivity}
          </TabsContent>
          {contributions.map((item) => (
            <TabsContent
              key={`${item.pluginId}:${item.index}`}
              value={activityTabValue(item.pluginId, item.index)}
              className="mt-3"
            >
              <PluginErrorBoundary pluginId={item.pluginId}>
                <item.component {...(context ?? {})} />
              </PluginErrorBoundary>
            </TabsContent>
          ))}
        </Tabs>
      ) : (
        hostActivity
      )}
    </SectionAccordion>
  );
}

function activityTabValue(pluginId: string, index: number): string {
  return `plugin:${pluginId}:${index}`;
}
