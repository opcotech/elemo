import { useNavigationContext } from "@/hooks/use-navigation-context";

export function QuickCreateContext() {
  const context = useNavigationContext();
  const contextLabel =
    context.type === "project"
      ? `Namespace ${context.namespaceId} / Project ${context.projectId}`
      : context.type === "namespace"
        ? `Namespace ${context.namespaceId}`
        : "Global context";

  return (
    <div className="bg-muted/50 rounded-lg border px-3 py-2.5">
      <p className="text-xs font-medium tracking-wide uppercase">Context</p>
      <p className="text-muted-foreground mt-1 text-sm">{contextLabel}</p>
      <p className="text-muted-foreground mt-1 text-xs">
        Inherited from the current route.
      </p>
    </div>
  );
}
