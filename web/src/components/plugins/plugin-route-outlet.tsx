import { Component, useSyncExternalStore } from "react";
import type { ErrorInfo, ReactNode } from "react";

import {
  matchPluginRoute,
  pluginPageState,
  subscribePluginRegistry,
} from "@/lib/plugins/registry";

class PluginRouteErrorBoundary extends Component<
  { children: ReactNode },
  { error: Error | null }
> {
  state = { error: null as Error | null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("Plugin route crashed", error, info);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="text-destructive p-6 text-sm">
          This plugin page failed to load.
        </div>
      );
    }
    return this.props.children;
  }
}

export function PluginRouteOutlet({
  pluginId,
  splat,
  context,
}: {
  pluginId: string;
  splat: string;
  context?: Record<string, unknown>;
}) {
  const state = useSyncExternalStore(
    subscribePluginRegistry,
    () => pluginPageState(pluginId, splat),
    () => pluginPageState(pluginId, splat)
  );
  const match = useSyncExternalStore(
    subscribePluginRegistry,
    () => matchPluginRoute(pluginId, splat),
    () => matchPluginRoute(pluginId, splat)
  );

  if (state === "loading") {
    return (
      <div
        data-testid="plugin-page-loading"
        className="text-muted-foreground p-6 text-sm"
      >
        Loading plugin page…
      </div>
    );
  }

  if (!match) {
    return (
      <div className="text-muted-foreground p-6 text-sm">
        This plugin page is not available.
      </div>
    );
  }

  return (
    <PluginRouteErrorBoundary>
      <match.component {...(context ?? {})} />
    </PluginRouteErrorBoundary>
  );
}
