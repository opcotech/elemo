import type { QueryKey } from "@tanstack/react-query";
import type { z } from "zod";

import {
  v1IssuesDocumentsGetOptions,
  v1NamespacesDocumentsGetOptions,
  v1OrganizationsDocumentsGetOptions,
  v1ProjectsDocumentsGetOptions,
} from "@/lib/api/query-options";
import {
  namespaceRefPath,
  organizationRefPath,
  projectIdPath,
} from "@/lib/api/refs";
import { zDocumentCreate } from "@/lib/api/schemas";
import type { DocumentCreate } from "@/lib/api/types";

export const documentCreateFormSchema = zDocumentCreate.pick({
  title: true,
});

export type DocumentCreateFormValues = z.infer<typeof documentCreateFormSchema>;

export const documentCreateFormDefaults: DocumentCreateFormValues = {
  title: "",
};

export type DocumentListParentType =
  "organization" | "namespace" | "project" | "issue";

export interface DocumentCreateParent {
  type: Exclude<DocumentListParentType, "issue">;
  id: string;
  organizationId?: string;
}

export function documentCreateBody(values: { title: string }): DocumentCreate {
  return {
    title: values.title.trim(),
  };
}

export function documentCreateParentFromNavigation(navigation: {
  type: string;
  organizationId?: string;
  namespaceId?: string;
  projectId?: string;
}): DocumentCreateParent | null {
  if (navigation.type === "project" && navigation.projectId) {
    return { type: "project", id: navigation.projectId };
  }
  if (
    navigation.type === "namespace" &&
    navigation.namespaceId &&
    navigation.organizationId
  ) {
    return {
      type: "namespace",
      id: navigation.namespaceId,
      organizationId: navigation.organizationId,
    };
  }
  if (navigation.type === "organization" && navigation.organizationId) {
    return { type: "organization", id: navigation.organizationId };
  }
  return null;
}

export function documentCreateContextCopy(input: {
  type: string;
  organizationName?: string;
  namespaceName?: string;
  projectName?: string;
}): string {
  if (input.type === "project") {
    const namespace = input.namespaceName ?? "this namespace";
    const project = input.projectName ?? "this project";
    return `Lives in ${namespace}. Related to ${project}.`;
  }
  if (input.type === "namespace") {
    return `Lives in ${input.namespaceName ?? "this namespace"}.`;
  }
  if (input.type === "organization") {
    return `Lives in ${input.organizationName ?? "this organization"}.`;
  }
  return "Global context";
}

export function documentListQueryKey(
  type: DocumentListParentType,
  id: string,
  organizationId?: string
): QueryKey {
  switch (type) {
    case "organization":
      return v1OrganizationsDocumentsGetOptions({
        path: organizationRefPath(id),
      }).queryKey;
    case "namespace":
      return v1NamespacesDocumentsGetOptions({
        path: namespaceRefPath(organizationId ?? id, id),
      }).queryKey;
    case "project":
      return v1ProjectsDocumentsGetOptions({
        path: projectIdPath(id),
      }).queryKey;
    case "issue":
      return v1IssuesDocumentsGetOptions({ path: { id } }).queryKey;
  }
}
