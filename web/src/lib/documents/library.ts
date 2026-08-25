import { z } from "zod";

import type { AccessibleWorkspace } from "@/lib/api/accessible-namespaces";
import { collectCursorPages, cursorPageQuery } from "@/lib/api/cursor-pages";
import { namespaceRefPath, organizationRefPath } from "@/lib/api/refs";
import {
  v1FolderGet,
  v1NamespacesFoldersGet,
  v1OrganizationsFoldersGet,
} from "@/lib/api/sdk";
import type { DocumentLibrary, Folder } from "@/lib/api/types";
import type { InternalPath } from "@/lib/internal-url";
import { internalPath } from "@/lib/internal-url";
import { namespaceDocumentsPath, organizationDocumentsPath } from "@/lib/paths";

export const LIBRARY_ROOT_FOLDER_VALUE = "root";
export const LIBRARY_ROOT_FOLDER_LABEL = "Library root";

export type DocumentLibraryKind = "organization" | "namespace";

export const documentLibrarySearchSchema = z
  .object({
    folder: z.string().min(1).optional(),
    all: z
      .union([z.boolean(), z.literal("true"), z.literal("false")])
      .optional(),
  })
  .transform((value) => {
    if (value.all === true || value.all === "true") {
      return { all: true as const };
    }
    if (value.folder) {
      return { folder: value.folder };
    }
    return {};
  })
  .catch({});

export type DocumentLibrarySearch = z.infer<typeof documentLibrarySearchSchema>;

export function resolveDocumentLibrarySearch(search: {
  folder?: string;
  all?: boolean | "true" | "false";
}): {
  folderId?: string;
  showAll: boolean;
} {
  if (search.all === true || search.all === "true") {
    return { showAll: true };
  }
  return {
    folderId: search.folder,
    showAll: false,
  };
}

export function documentLibrarySearchParams(state: {
  folderId?: string;
  showAll?: boolean;
}): DocumentLibrarySearch {
  if (state.showAll) {
    return { all: true };
  }
  if (state.folderId) {
    return { folder: state.folderId };
  }
  return {};
}

export function documentLibraryHref(
  library:
    | { kind: "organization"; organizationSlug: string }
    | { kind: "namespace"; organizationSlug: string; namespaceSlug: string }
): InternalPath {
  if (library.kind === "organization") {
    return internalPath(organizationDocumentsPath(library));
  }
  return internalPath(namespaceDocumentsPath(library));
}

export function documentLibraryApiPath(
  kind: "organization",
  organizationId: string,
  namespaceId?: string
): { organizationRef: string };
export function documentLibraryApiPath(
  kind: "namespace",
  organizationId: string,
  namespaceId?: string
): { organizationRef: string; namespaceRef: string };
export function documentLibraryApiPath(
  kind: DocumentLibraryKind,
  organizationId: string,
  namespaceId?: string
):
  | { organizationRef: string }
  | { organizationRef: string; namespaceRef: string };
export function documentLibraryApiPath(
  kind: DocumentLibraryKind,
  organizationId: string,
  namespaceId?: string
) {
  if (kind === "organization") {
    return organizationRefPath(organizationId);
  }
  return namespaceRefPath(organizationId, namespaceId ?? "");
}

export type DocumentLibraryTypeLabel = "Organization" | "Namespace";

export interface DocumentLibraryListItem {
  id: string;
  kind: DocumentLibraryKind;
  name: string;
  typeLabel: DocumentLibraryTypeLabel;
  href: InternalPath;
  documentCount: number;
}

export function documentLibraryTypeLabel(
  kind: DocumentLibraryKind
): DocumentLibraryTypeLabel {
  return kind === "organization" ? "Organization" : "Namespace";
}

export function documentLibraryListItem(library: {
  kind: DocumentLibraryKind;
  id: string;
  name: string;
  organizationSlug: string;
  namespaceSlug?: string;
  documentCount?: number | null;
}): DocumentLibraryListItem {
  return {
    id: library.id,
    kind: library.kind,
    name: library.name,
    typeLabel: documentLibraryTypeLabel(library.kind),
    href:
      library.kind === "organization"
        ? documentLibraryHref({
            kind: "organization",
            organizationSlug: library.organizationSlug,
          })
        : documentLibraryHref({
            kind: "namespace",
            organizationSlug: library.organizationSlug,
            namespaceSlug: library.namespaceSlug ?? "",
          }),
    documentCount: library.documentCount ?? 0,
  };
}

export function documentLibraryListItems(
  organizations: readonly {
    id: string;
    name: string;
    slug: string;
    document_count?: number | null;
  }[],
  namespaces: readonly {
    id: string;
    name: string;
    slug: string;
    organizationSlug: string;
    document_count?: number | null;
  }[]
): DocumentLibraryListItem[] {
  const byName = (
    left: DocumentLibraryListItem,
    right: DocumentLibraryListItem
  ) => left.name.localeCompare(right.name);

  return [
    ...organizations
      .map((organization) =>
        documentLibraryListItem({
          kind: "organization",
          id: organization.id,
          name: organization.name,
          organizationSlug: organization.slug,
          documentCount: organization.document_count,
        })
      )
      .sort(byName),
    ...namespaces
      .map((namespace) =>
        documentLibraryListItem({
          kind: "namespace",
          id: namespace.id,
          name: namespace.name,
          organizationSlug: namespace.organizationSlug,
          namespaceSlug: namespace.slug,
          documentCount: namespace.document_count,
        })
      )
      .sort(byName),
  ];
}

export function filterDocumentLibraryListItems(
  items: readonly DocumentLibraryListItem[],
  query: string
): DocumentLibraryListItem[] {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return [...items];
  }

  return items.filter((item) =>
    `${item.name} ${item.typeLabel}`.toLowerCase().includes(normalized)
  );
}

export function documentLibraryKindFromType(
  type: DocumentLibrary["type"]
): DocumentLibraryKind {
  return type === "Organization" ? "organization" : "namespace";
}

export function documentLibraryPageHref(
  library: DocumentLibrary,
  workspace?: AccessibleWorkspace
): InternalPath | null {
  if (library.type === "Organization") {
    const organization = workspace?.organizations.find(
      (item) => item.id === library.id
    );
    if (!organization) {
      return null;
    }
    return documentLibraryHref({
      kind: "organization",
      organizationSlug: organization.slug,
    });
  }

  const namespace = workspace?.namespaces.find(
    (item) => item.id === library.id
  );
  if (!namespace) {
    return null;
  }
  return documentLibraryHref({
    kind: "namespace",
    organizationSlug: namespace.organizationSlug,
    namespaceSlug: namespace.slug,
  });
}

export function folderPathLabel(names: readonly string[]): string {
  return names.join(" / ");
}

export interface LibraryBrowseCrumb {
  folderId?: string;
  name: string;
  current: boolean;
}

export function libraryBrowseCrumbs(
  folderPath: readonly Pick<Folder, "id" | "name">[]
): LibraryBrowseCrumb[] {
  return [
    { name: "Library", current: folderPath.length === 0 },
    ...folderPath.map((folder, index) => ({
      folderId: folder.id,
      name: folder.name,
      current: index === folderPath.length - 1,
    })),
  ];
}

export interface LibraryFolderOption {
  id: string;
  name: string;
  pathLabel: string;
  parentId: string | null;
}

export async function listLibraryFolders(
  kind: DocumentLibraryKind,
  organizationId: string,
  namespaceId: string | undefined,
  parentId: string | undefined,
  signal?: AbortSignal
): Promise<Folder[]> {
  return collectCursorPages(async (pageToken) => {
    if (kind === "organization") {
      const { data } = await v1OrganizationsFoldersGet({
        path: organizationRefPath(organizationId),
        query: {
          ...cursorPageQuery(pageToken),
          ...(parentId ? { parent_id: parentId } : {}),
        },
        signal,
        throwOnError: true,
      });
      return data;
    }

    const { data } = await v1NamespacesFoldersGet({
      path: namespaceRefPath(organizationId, namespaceId ?? ""),
      query: {
        ...cursorPageQuery(pageToken),
        ...(parentId ? { parent_id: parentId } : {}),
      },
      signal,
      throwOnError: true,
    });
    return data;
  });
}

export async function collectLibraryFolderOptions(
  kind: DocumentLibraryKind,
  organizationId: string,
  namespaceId: string | undefined,
  signal?: AbortSignal
): Promise<LibraryFolderOption[]> {
  const options: LibraryFolderOption[] = [];

  async function walk(parentId: string | undefined, prefix: string) {
    const folders = await listLibraryFolders(
      kind,
      organizationId,
      namespaceId,
      parentId,
      signal
    );
    folders.sort((left, right) => left.name.localeCompare(right.name));
    for (const folder of folders) {
      const pathLabel = prefix
        ? folderPathLabel([prefix, folder.name])
        : folder.name;
      options.push({
        id: folder.id,
        name: folder.name,
        pathLabel,
        parentId: folder.parent?.id ?? null,
      });
      await walk(folder.id, pathLabel);
    }
  }

  await walk(undefined, "");
  return options;
}

export const libraryFolderOptionsQueryKey = (
  kind: DocumentLibraryKind,
  organizationId: string,
  namespaceId?: string
) => ["library-folder-options", kind, organizationId, namespaceId] as const;

export function libraryFolderOptionsQuery(
  kind: DocumentLibraryKind,
  organizationId: string,
  namespaceId?: string
) {
  return {
    queryKey: libraryFolderOptionsQueryKey(kind, organizationId, namespaceId),
    queryFn: ({ signal }: { signal: AbortSignal }) =>
      collectLibraryFolderOptions(kind, organizationId, namespaceId, signal),
  };
}

export function descendantFolderIds(
  folders: readonly Pick<LibraryFolderOption, "id" | "parentId">[],
  folderId: string
): Set<string> {
  const childrenByParent = new Map<string | null, string[]>();
  for (const folder of folders) {
    const children = childrenByParent.get(folder.parentId) ?? [];
    children.push(folder.id);
    childrenByParent.set(folder.parentId, children);
  }

  const descendants = new Set<string>();
  const stack = [...(childrenByParent.get(folderId) ?? [])];
  while (stack.length > 0) {
    const current = stack.pop();
    if (!current || descendants.has(current)) {
      continue;
    }
    descendants.add(current);
    stack.push(...(childrenByParent.get(current) ?? []));
  }
  return descendants;
}

export function folderMoveTargets(
  folders: readonly LibraryFolderOption[],
  movingFolderId: string
): LibraryFolderOption[] {
  const excluded = descendantFolderIds(folders, movingFolderId);
  excluded.add(movingFolderId);
  return folders.filter((folder) => !excluded.has(folder.id));
}

export interface LibraryFolderPickerOption {
  value: string;
  title: string;
  searchText: string;
}

export function libraryFolderPickerOptions(
  folders: readonly Pick<LibraryFolderOption, "id" | "name" | "pathLabel">[]
): LibraryFolderPickerOption[] {
  return [
    {
      value: LIBRARY_ROOT_FOLDER_VALUE,
      title: LIBRARY_ROOT_FOLDER_LABEL,
      searchText: "library root unfiled",
    },
    ...folders.map((folder) => ({
      value: folder.id,
      title: folder.pathLabel,
      searchText: `${folder.name} ${folder.pathLabel}`,
    })),
  ];
}

export async function loadFolderAncestors(
  folderId: string,
  signal?: AbortSignal
): Promise<Folder[]> {
  const ancestors: Folder[] = [];
  let currentId: string | undefined = folderId;
  const seen = new Set<string>();

  while (currentId && !seen.has(currentId)) {
    seen.add(currentId);
    const { data } = await v1FolderGet({
      path: { id: currentId },
      signal,
      throwOnError: true,
    });
    const folder: Folder = data;
    ancestors.unshift(folder);
    currentId = folder.parent?.id;
  }

  return ancestors;
}

export const folderPathQueryKey = (folderId: string) =>
  ["folder-path", folderId] as const;

export function folderPathQuery(folderId: string) {
  return {
    queryKey: folderPathQueryKey(folderId),
    queryFn: ({ signal }: { signal: AbortSignal }) =>
      loadFolderAncestors(folderId, signal),
  };
}
