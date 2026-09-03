import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { PuzzleIcon, Trash2Icon, UploadIcon } from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";
import { v1PluginDeleteMutation } from "@/lib/api/mutation-options";
import { uploadPluginPackageFn } from "@/lib/api/plugin-upload";
import { v1PluginsGetOptions } from "@/lib/api/query-options";
import type { Plugin } from "@/lib/api/types";

async function fileToBase64(file: File): Promise<string> {
  const bytes = new Uint8Array(await file.arrayBuffer());
  let binary = "";
  const chunk = 0x8000;
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk));
  }
  return btoa(binary);
}

export function PluginCatalog({ canInstall }: { canInstall: boolean }) {
  const queryClient = useQueryClient();
  const [pendingDelete, setPendingDelete] = useState<Plugin | null>(null);
  const installInput = useRef<HTMLInputElement>(null);
  const upgradeInput = useRef<HTMLInputElement>(null);
  const [upgradeTarget, setUpgradeTarget] = useState<string | null>(null);
  const list = useQuery(v1PluginsGetOptions());

  const invalidate = () =>
    queryClient.invalidateQueries({
      predicate: (query) => JSON.stringify(query.queryKey).includes("v1Plugin"),
    });

  const remove = useMutation({
    ...v1PluginDeleteMutation(),
    onSuccess: async () => {
      toast.success("Plugin uninstalled");
      await invalidate();
    },
    onError: (error) => {
      toast.error(
        error instanceof ApiError ? error.message : "Failed to uninstall plugin"
      );
    },
  });

  async function onUpload(file: File, pluginId?: string) {
    const bytes = await fileToBase64(file);
    const result = await uploadPluginPackageFn({
      data: { filename: file.name, bytes, pluginId },
    });
    if (result.status >= 400) {
      let message = "Failed to upload plugin";
      try {
        const parsed = JSON.parse(result.body) as { message?: string };
        if (parsed.message) {
          message = parsed.message;
        }
      } catch {
        // keep default
      }
      toast.error(message);
      return;
    }
    toast.success(pluginId ? "Plugin upgraded" : "Plugin installed");
    await invalidate();
  }

  const plugins = list.data ?? [];

  return (
    <div className="space-y-4">
      {canInstall ? (
        <div className="flex items-center gap-2">
          <input
            ref={installInput}
            type="file"
            accept=".zip,application/zip"
            className="sr-only"
            onChange={(event) => {
              const file = event.target.files?.[0];
              event.target.value = "";
              if (file) {
                void onUpload(file);
              }
            }}
          />
          <Button
            variant="outline"
            size="sm"
            onClick={() => installInput.current?.click()}
          >
            <UploadIcon className="size-4" />
            Install package
          </Button>
        </div>
      ) : null}

      {list.isPending ? (
        <div className="space-y-3 rounded-xl border p-4" aria-busy="true">
          <Skeleton className="h-5 w-40" />
          <Skeleton className="h-4 w-64" />
        </div>
      ) : list.isError ? (
        <Alert variant="destructive">
          <AlertDescription>
            {list.error instanceof ApiError
              ? list.error.message
              : "Failed to load plugins. Please try again later."}
          </AlertDescription>
        </Alert>
      ) : plugins.length === 0 ? (
        <EmptyState
          icon={<PuzzleIcon />}
          title="No plugins installed"
          description="Upload a plugin zip to add capabilities without restarting Elemo."
        />
      ) : (
        <ul className="divide-border divide-y rounded-xl border">
          {plugins.map((plugin) => (
            <li
              key={plugin.plugin_id}
              className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div className="min-w-0 space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-medium">{plugin.name}</span>
                  <Badge variant="outline">{plugin.version}</Badge>
                  <Badge
                    variant={
                      plugin.status === "failed" ? "destructive" : "secondary"
                    }
                  >
                    {plugin.status}
                  </Badge>
                </div>
                <p className="text-muted-foreground font-mono text-xs">
                  {plugin.plugin_id}
                </p>
                {plugin.error ? (
                  <p className="text-destructive text-sm">{plugin.error}</p>
                ) : null}
                <p className="text-muted-foreground text-xs">
                  Capabilities: {plugin.capabilities.join(", ") || "none"}
                </p>
              </div>
              {canInstall ? (
                <div className="flex shrink-0 gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setUpgradeTarget(plugin.plugin_id);
                      upgradeInput.current?.click();
                    }}
                  >
                    Upgrade
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPendingDelete(plugin)}
                  >
                    <Trash2Icon className="size-4" />
                    Uninstall
                  </Button>
                </div>
              ) : null}
            </li>
          ))}
        </ul>
      )}

      <input
        ref={upgradeInput}
        type="file"
        accept=".zip,application/zip"
        className="sr-only"
        onChange={(event) => {
          const file = event.target.files?.[0];
          event.target.value = "";
          if (file && upgradeTarget) {
            void onUpload(file, upgradeTarget);
          }
          setUpgradeTarget(null);
        }}
      />

      <DeleteConfirmationDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open) {
            setPendingDelete(null);
          }
        }}
        title={`Uninstall ${pendingDelete?.name ?? "plugin"}?`}
        description="Uninstalling deletes plugin graph nodes and namespaced relations. This cannot be undone."
        consequences={[
          "The package is removed from this instance",
          "Extension graph data for this plugin is deleted",
        ]}
        deleteButtonText="Uninstall"
        isPending={remove.isPending}
        onConfirm={() => {
          if (!pendingDelete) {
            return;
          }
          remove.mutate({ path: { pluginId: pendingDelete.plugin_id } });
          setPendingDelete(null);
        }}
      />
    </div>
  );
}
