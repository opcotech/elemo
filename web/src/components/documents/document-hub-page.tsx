import { FileTextIcon, PlusIcon } from "lucide-react";
import { useMemo } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { PageHeader } from "@/components/ui/page-header";
import { SearchInput } from "@/components/ui/search-input";
import type { AccessibleWorkspace } from "@/lib/api/accessible-namespaces";
import {
  documentLibraryListItems,
  filterDocumentLibraryListItems,
} from "@/lib/documents/library";
import { pluralize } from "@/lib/utils";

export function DocumentHubPage({
  workspace,
  isLoading,
  query,
  onQueryChange,
}: {
  workspace?: AccessibleWorkspace;
  isLoading: boolean;
  query?: string;
  onQueryChange: (query: string | undefined) => void;
}) {
  const libraries = useMemo(
    () =>
      documentLibraryListItems(
        workspace?.organizations ?? [],
        workspace?.namespaces ?? []
      ),
    [workspace]
  );
  const filteredLibraries = useMemo(
    () => filterDocumentLibraryListItems(libraries, query ?? ""),
    [libraries, query]
  );
  const hasQuery = Boolean(query?.trim());

  return (
    <ContentWidth
      width="overview"
      className="space-y-6"
      data-section="document-hub"
    >
      <PageHeader
        title="Documents"
        description="Choose an organization or namespace library to browse folders and create documents."
        actions={
          libraries.length > 0 ? (
            <Button
              size="sm"
              disabled
              title="Choose a library to create a document"
            >
              <PlusIcon />
              Create
            </Button>
          ) : undefined
        }
      />

      {libraries.length > 0 || hasQuery ? (
        <div className="bg-background sticky top-0 z-10 flex flex-wrap items-center gap-2 py-3">
          <SearchInput
            value={query ?? ""}
            onChange={(value) => onQueryChange(value || undefined)}
            placeholder="Search libraries..."
            className="min-w-60"
          />
        </div>
      ) : null}

      {isLoading ? (
        <ListSkeleton />
      ) : filteredLibraries.length > 0 ? (
        <AppList data-section="document-libraries">
          {filteredLibraries.map((library) => (
            <EntityLink
              key={`${library.kind}:${library.id}`}
              type={library.kind}
              href={library.href}
              title={library.name}
              subtitle={`${library.typeLabel} · ${library.documentCount} ${pluralize(
                library.documentCount,
                "document",
                "documents"
              )}`}
            />
          ))}
        </AppList>
      ) : (
        <EmptyState
          icon={<FileTextIcon />}
          title={hasQuery ? "No matching libraries" : "No document libraries"}
          description={
            hasQuery
              ? "Try a different search."
              : "Documents live in organization and namespace libraries. Join or create one first, then come back here to browse folders."
          }
          action={
            hasQuery ? (
              <Button
                variant="outline"
                onClick={() => onQueryChange(undefined)}
              >
                Clear search
              </Button>
            ) : (
              <div className="flex flex-wrap items-center justify-center gap-2">
                <Button
                  variant="outline"
                  render={<InternalLink to="/organizations" />}
                >
                  View organizations
                </Button>
                <Button
                  variant="outline"
                  render={<InternalLink to="/namespaces" />}
                >
                  View namespaces
                </Button>
              </div>
            )
          }
        />
      )}
    </ContentWidth>
  );
}
