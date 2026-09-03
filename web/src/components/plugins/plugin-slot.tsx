import type { PluginSlotName } from "@elemo/plugin-sdk";
import { Component, useSyncExternalStore } from "react";
import type { ErrorInfo, ReactNode } from "react";

import {
  getSlotContributions,
  subscribePluginRegistry,
} from "@/lib/plugins/registry";

export class PluginErrorBoundary extends Component<
  { pluginId: string; children: ReactNode },
  { error: Error | null }
> {
  state = { error: null as Error | null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error(`Plugin ${this.props.pluginId} crashed`, error, info);
  }

  render() {
    if (this.state.error) {
      return (
        <div
          data-slot="plugin-error"
          className="text-destructive border-destructive/30 rounded-md border p-3 text-sm"
        >
          Plugin {this.props.pluginId} failed to render.
        </div>
      );
    }
    return this.props.children;
  }
}

export function PluginSlot({
  name,
  context,
}: {
  name: PluginSlotName;
  context?: Record<string, unknown>;
}) {
  const contributions = useSyncExternalStore(
    subscribePluginRegistry,
    () => getSlotContributions(name),
    () => getSlotContributions(name)
  );

  if (contributions.length === 0) {
    return null;
  }

  return (
    <div data-plugin-slot={name} className="space-y-3">
      {contributions.map((item) => (
        <PluginErrorBoundary
          key={`${item.pluginId}:${item.index}`}
          pluginId={item.pluginId}
        >
          <item.component {...(context ?? {})} />
        </PluginErrorBoundary>
      ))}
    </div>
  );
}
