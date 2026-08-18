import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import { ItemActions } from "@/components/ui/item";
import type { PartialDocument } from "@/lib/api/types";

export function DocumentSummaryList({
  documents,
  onUnlink,
}: {
  documents: readonly PartialDocument[];
  onUnlink?: (document: PartialDocument) => void;
}) {
  return (
    <AppList>
      {documents.map((document) => (
        <EntityLink
          key={document.id}
          type="document"
          href={`/documents/${document.id}`}
          title={document.title}
          subtitle={document.excerpt || "No summary"}
          actions={
            onUnlink ? (
              <ItemActions className="pr-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="xs"
                  aria-label={`Unlink ${document.title}`}
                  onClick={() => {
                    onUnlink(document);
                  }}
                >
                  Unlink
                </Button>
              </ItemActions>
            ) : undefined
          }
        />
      ))}
    </AppList>
  );
}
