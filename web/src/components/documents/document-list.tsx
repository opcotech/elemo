import {
  ArrowDownAZIcon,
  ChevronDownIcon,
  FileTextIcon,
  MoreHorizontalIcon,
  SearchIcon,
} from "lucide-react";
import { useMemo } from "react";
import type { ReactNode } from "react";

import { AppList } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { SearchableEntitySelect } from "@/components/ui/entity-select";
import type { EntitySelectOption } from "@/components/ui/entity-select";
import { Input } from "@/components/ui/input";
import { InternalLink } from "@/components/ui/internal-link";
import { Item, ItemActions, ItemContent } from "@/components/ui/item";
import { PersonAvatarStack } from "@/components/ui/person-avatar-stack";
import type { PartialDocument } from "@/lib/api/types";
import {
  ALL_DOCUMENT_CREATORS,
  documentAuthorName,
  documentExcerpt,
  documentListSortLabels,
  documentUpdatedAt,
  isDocumentListSort,
} from "@/lib/documents/document-list";
import type {
  DocumentCreatorOption,
  DocumentListSort,
} from "@/lib/documents/document-list";
import { formatDate } from "@/lib/format-date";
import { internalPath } from "@/lib/internal-url";

const EMPTY_CREATORS: DocumentCreatorOption[] = [];

export function DocumentListToolbar({
  query,
  onQueryChange,
  sort,
  onSortChange,
  creators = EMPTY_CREATORS,
  creatorId = ALL_DOCUMENT_CREATORS,
  onCreatorChange,
  trailing,
}: {
  query: string;
  onQueryChange: (value: string) => void;
  sort: DocumentListSort;
  onSortChange: (value: DocumentListSort) => void;
  creators?: readonly DocumentCreatorOption[];
  creatorId?: string;
  onCreatorChange?: (creatorId: string) => void;
  trailing?: ReactNode;
}) {
  const creatorOptions = useMemo(
    (): EntitySelectOption[] => [
      { value: ALL_DOCUMENT_CREATORS, title: "Anyone" },
      ...creators.map((creator) => ({
        value: creator.id,
        title: creator.name,
        avatarSrc: creator.picture,
        avatarFallback: creator.initials,
      })),
    ],
    [creators]
  );

  return (
    <div className="bg-background sticky top-0 z-10 flex flex-wrap items-center gap-2 py-3">
      <div className="relative min-w-60 flex-1">
        <SearchIcon className="text-muted-foreground absolute top-2.5 left-3 size-4" />
        <Input
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          placeholder="Search documents..."
          className="pl-9"
        />
      </div>
      {trailing}
      {onCreatorChange ? (
        <SearchableEntitySelect
          options={creatorOptions}
          value={creatorId}
          onValueChange={onCreatorChange}
          placeholder="Anyone"
          searchPlaceholder="Search people…"
          emptyMessage="No people found."
          appearance="button"
          aria-label="Filter by creator"
          triggerClassName="max-w-52"
        />
      ) : null}
      <DropdownMenu>
        <DropdownMenuTrigger
          render={<Button variant="outline" aria-label="Sort documents" />}
        >
          <ArrowDownAZIcon aria-hidden />
          Sort: {documentListSortLabels[sort]}
          <ChevronDownIcon />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuRadioGroup
            value={sort}
            onValueChange={(value) => {
              if (isDocumentListSort(value)) {
                onSortChange(value);
              }
            }}
          >
            <DropdownMenuRadioItem value="updated-desc">
              Updated
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="created-desc">
              Created
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="created-asc">
              Oldest
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="title-asc">
              Title
            </DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export function DocumentList({
  documents,
  onRename,
  onMove,
  onUnlink,
  onDelete,
}: {
  documents: readonly PartialDocument[];
  onRename?: (document: PartialDocument) => void;
  onMove?: (document: PartialDocument) => void;
  onUnlink?: (document: PartialDocument) => void;
  onDelete?: (document: PartialDocument) => void;
}) {
  return (
    <AppList>
      {documents.map((document) => {
        const hasActions = Boolean(onRename || onMove || onUnlink || onDelete);
        const excerpt = documentExcerpt(document);
        const authorName = documentAuthorName(document);
        const updatedAt = documentUpdatedAt(document);
        const updatedLabel = documentDateLabel(updatedAt);
        return (
          <Item
            key={document.id}
            role="listitem"
            size="sm"
            className="hover:bg-muted/40 p-0 [a]:hover:bg-transparent"
          >
            <InternalLink
              to={internalPath(`/documents/${document.id}`)}
              className="focus-visible:ring-ring grid min-w-0 flex-1 items-center gap-3 px-4 py-3 outline-none focus-visible:ring-2 focus-visible:ring-inset sm:grid-cols-[minmax(0,1fr)_minmax(9rem,12rem)_7.5rem]"
            >
              <div className="flex min-w-0 gap-3">
                <span className="bg-muted text-muted-foreground flex size-10 shrink-0 items-center justify-center rounded-lg">
                  <FileTextIcon className="size-5" />
                </span>
                <ItemContent className="min-w-0">
                  <h2 className="truncate font-medium">{document.title}</h2>
                  {excerpt ? (
                    <p className="text-muted-foreground mt-0.5 line-clamp-1 text-sm">
                      {excerpt}
                    </p>
                  ) : null}
                  <p className="text-muted-foreground mt-1 truncate text-xs sm:hidden">
                    {authorName}
                    {updatedLabel ? ` · ${updatedLabel}` : ""}
                  </p>
                </ItemContent>
              </div>
              <div className="hidden min-w-0 sm:flex sm:items-center">
                <PersonAvatarStack
                  people={[
                    {
                      id: document.created_by.id,
                      name: authorName,
                      picture: document.created_by.picture,
                    },
                  ]}
                  size="sm"
                />
                <span className="text-muted-foreground ml-2 truncate text-sm">
                  {authorName}
                </span>
              </div>
              <time
                className="text-muted-foreground hidden text-sm sm:block"
                dateTime={updatedAt ?? undefined}
              >
                {updatedLabel ?? "—"}
              </time>
            </InternalLink>
            {hasActions ? (
              <ItemActions className="pr-3">
                <DocumentRowMenu
                  document={document}
                  onRename={onRename}
                  onMove={onMove}
                  onUnlink={onUnlink}
                  onDelete={onDelete}
                />
              </ItemActions>
            ) : null}
          </Item>
        );
      })}
    </AppList>
  );
}

function DocumentRowMenu({
  document,
  onRename,
  onMove,
  onUnlink,
  onDelete,
}: {
  document: PartialDocument;
  onRename?: (document: PartialDocument) => void;
  onMove?: (document: PartialDocument) => void;
  onUnlink?: (document: PartialDocument) => void;
  onDelete?: (document: PartialDocument) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label={`Document actions for ${document.title}`}
          />
        }
      >
        <MoreHorizontalIcon />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {onRename ? (
          <DropdownMenuItem onClick={() => onRename(document)}>
            Rename
          </DropdownMenuItem>
        ) : null}
        {onMove ? (
          <DropdownMenuItem onClick={() => onMove(document)}>
            Move
          </DropdownMenuItem>
        ) : null}
        {onUnlink ? (
          <DropdownMenuItem onClick={() => onUnlink(document)}>
            Unlink
          </DropdownMenuItem>
        ) : null}
        {onDelete ? (
          <>
            {onRename || onMove || onUnlink ? <DropdownMenuSeparator /> : null}
            <DropdownMenuItem
              variant="destructive"
              onClick={() => onDelete(document)}
            >
              Delete
            </DropdownMenuItem>
          </>
        ) : null}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function documentDateLabel(value: string | null | undefined): string | null {
  if (!value) {
    return null;
  }
  const label = formatDate(value);
  return label === "N/A" ? null : label;
}
