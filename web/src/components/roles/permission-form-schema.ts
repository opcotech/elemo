import { z } from "zod";

import { zPermissionKind, zResourceType } from "@/lib/client/zod.gen";

export const organizationScopedResourceType = zResourceType.extract([
  "Organization",
  "Namespace",
  "Document",
  "Project",
  "Role",
]);

export const permissionFormSchema = z.object({
  resourceType: organizationScopedResourceType,
  resourceId: z.string().min(1, "Resource ID is required"),
  kind: zPermissionKind,
});

export type PermissionFormValues = z.infer<typeof permissionFormSchema>;
