import { useQuery } from "@tanstack/react-query";

import { v1PermissionResourceGetOptions } from "@/lib/api";

export function usePermissions(
  resourceId: string,
  hasSSRResponse: boolean = false
) {
  return useQuery({
    enabled: !hasSSRResponse,
    ...v1PermissionResourceGetOptions({
      path: {
        resourceId,
      },
    }),
  });
}
