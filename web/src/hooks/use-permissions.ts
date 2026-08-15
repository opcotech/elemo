import { useQueries, useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import { v1PermissionResourceGetOptions } from "@/lib/api/query-options";
import type { ResourceType } from "@/lib/auth/permissions";
import { withResourceType } from "@/lib/auth/permissions";

export { ResourceType, withResourceType } from "@/lib/auth/permissions";

export function usePermissions(resourceId: string, disabled: boolean = false) {
  return useQuery({
    enabled: !disabled,
    ...v1PermissionResourceGetOptions({
      path: {
        resourceId,
      },
    }),
  });
}

export function usePermissionsByResourceId(
  resourceType: ResourceType,
  ids: readonly string[]
) {
  const queries = useQueries({
    queries: ids.map((id) =>
      v1PermissionResourceGetOptions({
        path: {
          resourceId: withResourceType(resourceType, id),
        },
      })
    ),
  });

  return useMemo(
    () => new Map(ids.map((id, index) => [id, queries[index]])),
    [ids, queries]
  );
}
