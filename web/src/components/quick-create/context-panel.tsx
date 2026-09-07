import { useQuery } from "@tanstack/react-query";

import { useNavigationContext } from "@/hooks/use-navigation-context";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { v1ProjectGetOptions } from "@/lib/api/query-options";
import { projectIdPath } from "@/lib/api/refs";
import {
  documentCreateContextCopy,
  documentCreateParentFromNavigation,
} from "@/lib/documents/create";

export function QuickCreateContext() {
  const context = useNavigationContext();
  const parent = documentCreateParentFromNavigation(context);
  const { data: workspace } = useAccessibleNamespaces();
  const { data: project } = useQuery({
    ...v1ProjectGetOptions({ path: projectIdPath(context.projectId ?? "") }),
    enabled: context.type === "project" && Boolean(context.projectId),
  });
  const organization = workspace?.organizations.find(
    (item) => item.id === context.organizationId
  );
  const namespace = workspace?.namespaces.find(
    (item) => item.id === context.namespaceId
  );
  const contextLabel = documentCreateContextCopy({
    type: context.type,
    organizationName: organization?.name,
    namespaceName: namespace?.name,
    projectName: project?.name,
  });

  return (
    <div className="rounded-lg border bg-muted/50 px-3 py-2.5">
      <p className="font-medium text-xs uppercase tracking-wide">Context</p>
      <p className="mt-1 text-muted-foreground text-sm">{contextLabel}</p>
      <p className="mt-1 text-muted-foreground text-xs">
        {parent
          ? "Create stays in this context. Folders can be chosen later."
          : "Inherited from the current route."}
      </p>
    </div>
  );
}
