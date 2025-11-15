import { FileText } from "lucide-react";
import { useMemo, useState } from "react";

import { ListContainer } from "@/components/ui/list-container";
import { SearchInput } from "@/components/ui/search-input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { NamespaceDocument } from "@/lib/api";
import { formatDate } from "@/lib/utils";

function NamespaceDocumentsListSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Excerpt</TableHead>
          <TableHead>Created At</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-5 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-48" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-24" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

interface NamespaceDocumentsListProps {
  documents: NamespaceDocument[];
  isLoading: boolean;
  error: unknown;
}

export function NamespaceDocumentsList({
  documents,
  isLoading,
  error,
}: NamespaceDocumentsListProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const filteredDocuments = useMemo(() => {
    if (!searchTerm.trim()) return documents;
    const term = searchTerm.toLowerCase();
    return documents.filter(
      (document) =>
        document.name.toLowerCase().includes(term) ||
        (document.excerpt && document.excerpt.toLowerCase().includes(term))
    );
  }, [documents, searchTerm]);

  const emptyState =
    documents.length === 0
      ? {
          icon: <FileText />,
          title: "No documents found",
          description:
            "Documents will appear here when they are added to this namespace.",
        }
      : filteredDocuments.length === 0 && searchTerm.trim()
        ? {
            icon: <FileText />,
            title: "No documents found",
            description:
              "No documents match your search criteria. Try adjusting your search.",
          }
        : undefined;

  const shouldShowSearch = documents.length > 0 || searchTerm.trim() !== "";

  return (
    <ListContainer
      title="Documents"
      description="Documents in this namespace."
      isLoading={isLoading}
      error={error}
      emptyState={emptyState}
      searchInput={
        shouldShowSearch ? (
          <SearchInput
            value={searchTerm}
            onChange={setSearchTerm}
            placeholder="Search documents..."
            disabled={isLoading}
          />
        ) : undefined
      }
    >
      {isLoading ? (
        <NamespaceDocumentsListSkeleton />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Excerpt</TableHead>
              <TableHead>Created At</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredDocuments.map((document) => (
              <TableRow key={document.id}>
                <TableCell className="font-medium">{document.name}</TableCell>
                <TableCell>
                  <span className="text-muted-foreground text-sm">
                    {document.excerpt || "—"}
                  </span>
                </TableCell>
                <TableCell>
                  <span className="text-muted-foreground text-sm">
                    {document.created_at
                      ? formatDate(document.created_at)
                      : "—"}
                  </span>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </ListContainer>
  );
}
