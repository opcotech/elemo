import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";
import {
  v1PluginConfigPatchMutation,
  v1PluginDisableMutation,
  v1PluginEnableMutation,
} from "@/lib/api/mutation-options";
import { v1PluginsGetOptions } from "@/lib/api/query-options";
import type {
  Plugin,
  PluginConfigField,
  PluginGraphForeignSummary,
  PluginGraphKindSummary,
  ResourceType,
} from "@/lib/api/types";

type GraphBinding = { plugin_id: string; kind: string };

function asRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function graphBindingValue(value: unknown): GraphBinding | null {
  const record = asRecord(value);
  const pluginId = record.plugin_id;
  const kind = record.kind;
  if (
    typeof pluginId === "string" &&
    typeof kind === "string" &&
    pluginId &&
    kind
  ) {
    return { plugin_id: pluginId, kind };
  }
  return null;
}

function kindMatchesForeign(
  kind: PluginGraphKindSummary,
  foreign: PluginGraphForeignSummary
): boolean {
  if (kind.parent !== foreign.parent) {
    return false;
  }
  for (const want of foreign.properties ?? []) {
    const got = kind.properties?.find(
      (property) => property.name === want.name
    );
    if (!got || got.type !== want.type) {
      return false;
    }
  }
  return true;
}

function GraphBindingPicker({
  plugin,
  plugins,
  field,
  foreign,
  canManage,
  scopeId,
  scopeType,
}: {
  plugin: Plugin;
  plugins: Plugin[];
  field: PluginConfigField;
  foreign: PluginGraphForeignSummary;
  canManage: boolean;
  scopeId: string;
  scopeType: ResourceType;
}) {
  const queryClient = useQueryClient();
  const current = graphBindingValue(asRecord(plugin.config)[field.name]);
  const [pluginId, setPluginId] = useState(current?.plugin_id ?? "");
  const [kind, setKind] = useState(current?.kind ?? "");

  const candidates = useMemo(() => {
    return plugins
      .filter((item) => item.plugin_id !== plugin.plugin_id)
      .map((item) => ({
        plugin: item,
        kinds: (item.graph?.nodes ?? []).filter((node) =>
          kindMatchesForeign(node, foreign)
        ),
      }))
      .filter((item) => item.kinds.length > 0);
  }, [foreign, plugin.plugin_id, plugins]);

  const selected = candidates.find(
    (item) => item.plugin.plugin_id === pluginId
  );
  const kindItems = (selected?.kinds ?? []).map((node) => ({
    value: node.kind,
    label: node.kind,
  }));
  const pluginItems = [
    { value: "__none__", label: "None" },
    ...candidates.map((item) => ({
      value: item.plugin.plugin_id,
      label: item.plugin.name,
    })),
  ];

  const save = useMutation({
    ...v1PluginConfigPatchMutation(),
    onSuccess: async () => {
      toast.success("Plugin config saved");
      await queryClient.invalidateQueries({
        predicate: (query) =>
          JSON.stringify(query.queryKey).includes("v1Plugin"),
      });
    },
    onError: (error) => {
      toast.error(
        error instanceof ApiError
          ? error.message
          : "Failed to save plugin config"
      );
    },
  });

  if (!plugin.enabled) {
    return null;
  }

  return (
    <div className="space-y-2" data-testid={`plugin-binding-${field.name}`}>
      <Label className="text-sm font-medium">
        {field.name === "time_source" ? "Time source" : field.name}
      </Label>
      <p className="text-muted-foreground text-xs">
        Bind {foreign.name} ({foreign.parent}
        {foreign.properties?.length
          ? `, ${foreign.properties.map((property) => property.name).join(", ")}`
          : ""}
        ) to another plugin kind.
      </p>
      <div className="flex flex-col gap-2 sm:flex-row">
        <Select
          value={pluginId || "__none__"}
          onValueChange={(value) => {
            const next = value === "__none__" || !value ? "" : value;
            setPluginId(next);
            setKind("");
          }}
          items={pluginItems}
          disabled={!canManage || save.isPending}
        >
          <SelectTrigger className="sm:w-56">
            <SelectValue placeholder="Plugin" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">None</SelectItem>
            {candidates.map((item) => (
              <SelectItem
                key={item.plugin.plugin_id}
                value={item.plugin.plugin_id}
              >
                {item.plugin.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select
          value={kind}
          onValueChange={(value) => setKind(value ?? "")}
          items={kindItems}
          disabled={!canManage || save.isPending || !pluginId}
        >
          <SelectTrigger className="sm:w-44">
            <SelectValue placeholder="Kind" />
          </SelectTrigger>
          <SelectContent>
            {kindItems.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {canManage ? (
          <Button
            variant="outline"
            size="sm"
            data-testid="plugin-binding-save"
            disabled={save.isPending}
            onClick={() => {
              const next = { ...asRecord(plugin.config) };
              if (!pluginId || !kind) {
                delete next[field.name];
              } else {
                next[field.name] = { plugin_id: pluginId, kind };
              }
              save.mutate({
                path: { pluginId: plugin.plugin_id },
                query: { scope_id: scopeId, scope_type: scopeType },
                body: { config: next },
              });
            }}
          >
            Save binding
          </Button>
        ) : null}
      </div>
    </div>
  );
}

export function PluginActivationManager({
  scopeId,
  scopeType,
  canManage,
}: {
  scopeId: string;
  scopeType: ResourceType;
  canManage: boolean;
}) {
  const queryClient = useQueryClient();
  const list = useQuery(
    v1PluginsGetOptions({
      query: { scope_id: scopeId, scope_type: scopeType },
    })
  );

  const invalidate = () =>
    queryClient.invalidateQueries({
      predicate: (query) => JSON.stringify(query.queryKey).includes("v1Plugin"),
    });

  const enable = useMutation({
    ...v1PluginEnableMutation(),
    onSuccess: async () => {
      toast.success("Plugin enabled");
      await invalidate();
    },
    onError: (error) => {
      toast.error(
        error instanceof ApiError ? error.message : "Failed to enable plugin"
      );
    },
  });

  const disable = useMutation({
    ...v1PluginDisableMutation(),
    onSuccess: async () => {
      toast.success("Plugin disabled");
      await invalidate();
    },
    onError: (error) => {
      toast.error(
        error instanceof ApiError ? error.message : "Failed to disable plugin"
      );
    },
  });

  if (list.isPending) {
    return (
      <div className="space-y-3 rounded-xl border p-4" aria-busy="true">
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-4 w-64" />
        <Skeleton className="h-9 w-24" />
      </div>
    );
  }

  if (list.isError) {
    return (
      <Alert variant="destructive">
        <AlertDescription>
          {list.error instanceof ApiError
            ? list.error.message
            : "Failed to load plugins. Please try again later."}
        </AlertDescription>
      </Alert>
    );
  }

  const plugins = list.data ?? [];

  if (plugins.length === 0) {
    return (
      <EmptyState
        title="No plugins installed"
        description="Ask an administrator to install a plugin package, then enable it here."
      />
    );
  }

  return (
    <ul className="divide-border divide-y rounded-xl border">
      {plugins.map((plugin) => {
        const enabled = plugin.enabled === true;
        const bindingFields = (plugin.config_schema ?? []).filter(
          (field) => field.type === "graph_binding"
        );
        return (
          <li key={plugin.plugin_id} className="space-y-3 p-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="min-w-0 space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium">{plugin.name}</span>
                  <Badge variant={enabled ? "success" : "secondary"}>
                    {enabled ? "Enabled" : "Disabled"}
                  </Badge>
                </div>
                <p className="text-muted-foreground font-mono text-xs">
                  {plugin.plugin_id}
                </p>
              </div>
              {canManage ? (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={enable.isPending || disable.isPending}
                  onClick={() => {
                    const body = { scope_id: scopeId, scope_type: scopeType };
                    if (enabled) {
                      disable.mutate({
                        path: { pluginId: plugin.plugin_id },
                        body,
                      });
                    } else {
                      enable.mutate({
                        path: { pluginId: plugin.plugin_id },
                        body,
                      });
                    }
                  }}
                >
                  {enabled ? "Disable" : "Enable"}
                </Button>
              ) : null}
            </div>
            {bindingFields.map((field) => {
              const foreign = plugin.graph?.foreign?.find(
                (item) => item.name === field.foreign
              );
              if (!foreign) {
                return null;
              }
              return (
                <GraphBindingPicker
                  key={field.name}
                  plugin={plugin}
                  plugins={plugins}
                  field={field}
                  foreign={foreign}
                  canManage={canManage}
                  scopeId={scopeId}
                  scopeType={scopeType}
                />
              );
            })}
          </li>
        );
      })}
    </ul>
  );
}
