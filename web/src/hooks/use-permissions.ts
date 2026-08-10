import { useQuery } from "@tanstack/react-query";

import { v1PermissionResourceGetOptions } from "@/lib/api/query-options";

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
