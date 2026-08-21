import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import {
  ChevronDownIcon,
  FileTextIcon,
  FilesIcon,
  FolderIcon,
  MoreHorizontalIcon,
} from "lucide-react";
import { useMemo, useState } from "react";

import { DocumentDeleteDialog } from "./document-delete-dialog";
import { DocumentList, DocumentListToolbar } from "./document-list";
import { DocumentMoveDialog } from "./document-move-dialog";
import { DocumentRenameDialog } from "./document-rename-dialog";
import { FolderCreateDialog } from "./folder-create-dialog";
import { FolderDeleteDialog } from "./folder-delete-dialog";
import { FolderMoveDialog } from "./folder-move-dialog";
import { FolderRenameDialog } from "./folder-rename-dialog";

import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { EntityHeader } from "@/components/shared/entity-header";
import { AppList } from "@/components/shared/entity-link";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { CreateButton } from "@/components/ui/create-button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import {
  Item,
  ItemActions,
  ItemContent,
  ItemMedia,
  ItemTitle,
} from "@/components/ui/item";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1NamespacesDocumentsGetOptions,
  v1NamespacesFoldersGetOptions,
  v1OrganizationsDocumentsGetOptions,
  v1OrganizationsFoldersGetOptions,
} from "@/lib/api/query-options";
import type { Folder, PartialDocument } from "@/lib/api/types";
import {
  ALL_DOCUMENT_CREATORS,
  documentCreators,
  visibleDocuments,
} from "@/lib/documents/document-list";
import type { DocumentListSort } from "@/lib/documents/document-list";
import {
  documentLibraryHref,
  documentLibrarySearchParams,
  folderPathQuery,
  libraryBrowseCrumbs,
  resolveDocumentLibrarySearch,
} from "@/lib/documents/library";
import type {
  DocumentLibraryKind,
  DocumentLibrarySearch,
} from "@/lib/documents/library";
import { pluralize } from "@/lib/utils";

export function DocumentLibraryPage({
  kind,
  libraryId,
  libraryName,
  search,
  mayWrite,
  hasReadAccess = true,
  documentCount,
  limitedAccessTitle,
  limitedAccessDescription,
}: {
  kind: DocumentLibraryKind;
  libraryId: string;
  libraryName: string;
  search: DocumentLibrarySearch;
  mayWrite: boolean;
  hasReadAccess?: boolean;
  documentCount?: number;
  limitedAccessTitle: string;
  limitedAccessDescription: string;
}) {
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<DocumentListSort>("updated-desc");
  const [creatorId, setCreatorId] = useState(ALL_DOCUMENT_CREATORS);
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [renamingFolder, setRenamingFolder] = useState<Folder | null>(null);
  const [movingFolder, setMovingFolder] = useState<Folder | null>(null);
  const [deletingFolder, setDeletingFolder] = useState<Folder | null>(null);
  const [renamingDocument, setRenamingDocument] =
    useState<PartialDocument | null>(null);
  const [movingDocument, setMovingDocument] = useState<PartialDocument | null>(
    null
  );
  const [deletingDocument, setDeletingDocument] =
    useState<PartialDocument | null>(null);
  const navigate = useNavigate();
  const { folderId, showAll } = resolveDocumentLibrarySearch(search);
  const documentsHref = documentLibraryHref(kind, libraryId);
  const BrowseIcon = showAll ? FilesIcon : FolderIcon;
  const browseLabel = showAll ? "All" : "Library";
  const documentsQuery = showAll
    ? { all: true as const }
    : folderId
      ? { folder_id: folderId }
      : {};
  const documentsNav = useCursorPageNav({
    resetKey: `${kind}:${libraryId}:${folderId ?? ""}:${showAll}:${query}`,
  });
  const listOptions =
    kind === "organization"
      ? v1OrganizationsDocumentsGetOptions({
          path: { id: libraryId },
          query: {
            ...cursorPageQuery(documentsNav.pageToken),
            ...documentsQuery,
          },
        })
      : v1NamespacesDocumentsGetOptions({
          path: { id: libraryId },
          query: {
            ...cursorPageQuery(documentsNav.pageToken),
            ...documentsQuery,
          },
        });
  const folderListOptions =
    kind === "organization"
      ? v1OrganizationsFoldersGetOptions({
          path: { id: libraryId },
          query: {
            ...cursorPageQuery(),
            ...(folderId ? { parent_id: folderId } : {}),
          },
        })
      : v1NamespacesFoldersGetOptions({
          path: { id: libraryId },
          query: {
            ...cursorPageQuery(),
            ...(folderId ? { parent_id: folderId } : {}),
          },
        });

  const { data: documentsPage, isLoading: isDocumentsLoading } = useQuery({
    ...listOptions,
    enabled: hasReadAccess,
  });
  const { data: foldersPage, isLoading: isFoldersLoading } = useQuery({
    ...folderListOptions,
    enabled: hasReadAccess && !showAll,
  });
  const { data: folderPath } = useQuery({
    ...folderPathQuery(folderId ?? ""),
    enabled: hasReadAccess && Boolean(folderId),
  });

  const documents = useMemo(
    () => visibleDocuments(documentsPage?.items ?? [], query, sort, creatorId),
    [creatorId, documentsPage?.items, query, sort]
  );
  const creators = useMemo(
    () => documentCreators(documentsPage?.items ?? []),
    [documentsPage?.items]
  );
  const folders = useMemo(() => {
    const items = [...(foldersPage?.items ?? [])];
    const normalized = query.trim().toLowerCase();
    const filtered = normalized
      ? items.filter((folder) => folder.name.toLowerCase().includes(normalized))
      : items;
    return filtered.sort((left, right) => left.name.localeCompare(right.name));
  }, [foldersPage?.items, query]);

  const countLabel =
    documentCount != null
      ? `${documentCount} ${pluralize(documentCount, "document", "documents")} that live here.`
      : "Documents that live here.";
  const hasFolders = !showAll && folders.length > 0;
  const hasDocuments = documents.length > 0;
  const isLibraryLoading = isDocumentsLoading || isFoldersLoading;
  const emptyTitle = query
    ? "No document matches"
    : showAll
      ? "No documents yet"
      : hasFolders
        ? "No documents in this folder"
        : "No documents yet";
  const emptyDescription = query
    ? "Try a different title, summary, or folder name."
    : showAll
      ? "Documents added to this library will appear here."
      : "Create a document or folder at this location.";

  return (
    <ContentWidth
      width="overview"
      className="space-y-6"
      data-section="documents"
    >
      <EntityHeader
        type="document"
        eyebrow={libraryName}
        title="Documents"
        description={countLabel}
        showIcon={false}
        actions={
          hasReadAccess && mayWrite ? (
            <div className="flex items-center gap-2">
              {!showAll ? (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCreateFolderOpen(true)}
                >
                  <FolderIcon />
                  New folder
                </Button>
              ) : null}
              <CreateButton onClick={() => openQuickCreate("document")} />
            </div>
          ) : undefined
        }
      />
      {!hasReadAccess ? (
        <EmptyState
          icon={<FileTextIcon />}
          title={limitedAccessTitle}
          description={limitedAccessDescription}
        />
      ) : (
        <>
          <LibraryBrowseHeader
            libraryName={libraryName}
            documentsHref={documentsHref}
            folderPath={folderId ? (folderPath ?? []) : []}
            showAll={showAll}
          />
          <DocumentListToolbar
            query={query}
            onQueryChange={setQuery}
            sort={sort}
            onSortChange={setSort}
            creators={creators}
            creatorId={creatorId}
            onCreatorChange={setCreatorId}
            trailing={
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={
                    <Button variant="outline" aria-label="Browse documents" />
                  }
                >
                  <BrowseIcon aria-hidden />
                  {browseLabel}
                  <ChevronDownIcon />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start">
                  <DropdownMenuRadioGroup
                    value={showAll ? "all" : "library"}
                    onValueChange={(value) => {
                      void navigate({
                        to: documentsHref as never,
                        search: documentLibrarySearchParams(
                          value === "all" ? { showAll: true } : { folderId }
                        ) as never,
                      });
                    }}
                  >
                    <DropdownMenuRadioItem value="library">
                      Library
                    </DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value="all">
                      All
                    </DropdownMenuRadioItem>
                  </DropdownMenuRadioGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            }
          />
          {isLibraryLoading ? (
            <ListSkeleton />
          ) : hasFolders || hasDocuments ? (
            <>
              {hasFolders ? (
                <AppList data-section="library-folders">
                  {folders.map((folder) => (
                    <Item
                      key={folder.id}
                      role="listitem"
                      size="sm"
                      className="hover:bg-muted/40 p-0 [a]:hover:bg-transparent"
                    >
                      <InternalLink
                        to={documentsHref}
                        search={documentLibrarySearchParams({
                          folderId: folder.id,
                        })}
                        className="focus-visible:ring-ring flex min-w-0 flex-1 items-center gap-3 px-4 py-4 outline-none focus-visible:ring-2 focus-visible:ring-inset"
                      >
                        <ItemMedia
                          variant="icon"
                          className="bg-muted text-muted-foreground size-10 rounded-lg"
                        >
                          <FolderIcon className="size-5" />
                        </ItemMedia>
                        <ItemContent>
                          <ItemTitle>{folder.name}</ItemTitle>
                        </ItemContent>
                      </InternalLink>
                      {mayWrite ? (
                        <ItemActions className="pr-3">
                          <FolderRowMenu
                            folder={folder}
                            onRename={setRenamingFolder}
                            onMove={setMovingFolder}
                            onDelete={setDeletingFolder}
                          />
                        </ItemActions>
                      ) : null}
                    </Item>
                  ))}
                </AppList>
              ) : null}
              {hasDocuments ? (
                <DocumentList
                  documents={documents}
                  onRename={
                    mayWrite
                      ? (document) => {
                          setRenamingDocument(document);
                        }
                      : undefined
                  }
                  onMove={
                    mayWrite
                      ? (document) => {
                          setMovingDocument(document);
                        }
                      : undefined
                  }
                  onDelete={
                    mayWrite
                      ? (document) => {
                          setDeletingDocument(document);
                        }
                      : undefined
                  }
                />
              ) : null}
              <CursorPaginator
                {...cursorPaginatorProps(documentsPage, documentsNav)}
              />
            </>
          ) : (
            <EmptyState
              icon={<FileTextIcon />}
              title={emptyTitle}
              description={emptyDescription}
              action={
                query ? (
                  <Button variant="outline" onClick={() => setQuery("")}>
                    Clear search
                  </Button>
                ) : mayWrite ? (
                  <CreateButton onClick={() => openQuickCreate("document")} />
                ) : undefined
              }
            />
          )}
        </>
      )}
      {mayWrite ? (
        <>
          <FolderCreateDialog
            kind={kind}
            libraryId={libraryId}
            parentId={folderId}
            open={createFolderOpen}
            onOpenChange={setCreateFolderOpen}
          />
          <FolderRenameDialog
            folder={renamingFolder}
            open={renamingFolder != null}
            onOpenChange={(open) => {
              if (!open) {
                setRenamingFolder(null);
              }
            }}
          />
          <FolderMoveDialog
            folder={movingFolder}
            kind={kind}
            libraryId={libraryId}
            open={movingFolder != null}
            onOpenChange={(open) => {
              if (!open) {
                setMovingFolder(null);
              }
            }}
          />
          <FolderDeleteDialog
            folder={deletingFolder}
            open={deletingFolder != null}
            onOpenChange={(open) => {
              if (!open) {
                setDeletingFolder(null);
              }
            }}
          />
        </>
      ) : null}
      {renamingDocument ? (
        <DocumentRenameDialog
          document={renamingDocument}
          open
          onOpenChange={(open) => {
            if (!open) {
              setRenamingDocument(null);
            }
          }}
        />
      ) : null}
      {movingDocument ? (
        <DocumentMoveDialog
          documentId={movingDocument.id}
          documentTitle={movingDocument.title}
          kind={kind}
          libraryId={libraryId}
          currentFolderId={folderId}
          open
          onOpenChange={(open) => {
            if (!open) {
              setMovingDocument(null);
            }
          }}
        />
      ) : null}
      {deletingDocument ? (
        <DocumentDeleteDialog
          document={deletingDocument}
          open
          onOpenChange={(open) => {
            if (!open) {
              setDeletingDocument(null);
            }
          }}
          navigateOnSuccess={false}
        />
      ) : null}
    </ContentWidth>
  );
}

function FolderRowMenu({
  folder,
  onRename,
  onMove,
  onDelete,
}: {
  folder: Folder;
  onRename: (folder: Folder) => void;
  onMove: (folder: Folder) => void;
  onDelete: (folder: Folder) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label={`Folder actions for ${folder.name}`}
          />
        }
      >
        <MoreHorizontalIcon />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => onRename(folder)}>
          Rename
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => onMove(folder)}>Move</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          variant="destructive"
          onClick={() => onDelete(folder)}
        >
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function LibraryBrowseHeader({
  libraryName,
  documentsHref,
  folderPath,
  showAll,
}: {
  libraryName: string;
  documentsHref: ReturnType<typeof documentLibraryHref>;
  folderPath: Folder[];
  showAll: boolean;
}) {
  if (showAll) {
    return (
      <p className="text-muted-foreground text-sm">
        Every document in this library, including those in folders.
      </p>
    );
  }

  const crumbs = libraryBrowseCrumbs(folderPath);

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {crumbs.map((crumb, index) => (
          <span key={crumb.folderId ?? "library"} className="contents">
            {index > 0 ? <BreadcrumbSeparator /> : null}
            <BreadcrumbItem>
              {crumb.current ? (
                <BreadcrumbPage>{crumb.name}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink
                  render={
                    <InternalLink
                      to={documentsHref}
                      search={documentLibrarySearchParams({
                        folderId: crumb.folderId,
                      })}
                    />
                  }
                >
                  {crumb.name}
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
          </span>
        ))}
        {folderPath.length === 0 ? (
          <span className="sr-only">{libraryName} library root</span>
        ) : null}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
