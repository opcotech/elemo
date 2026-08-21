import { Button } from "@/components/ui/button";
import type { CursorPage } from "@/lib/api/cursor-pages";
import { nextCursorPageToken } from "@/lib/api/cursor-pages";

export function cursorPaginatorProps<T>(
  page: CursorPage<T> | null | undefined,
  nav: {
    canGoPrevious: boolean;
    goPrevious: () => void;
    goNext: (nextPageToken: string) => void;
  }
) {
  const nextPageToken = nextCursorPageToken(page);
  return {
    canGoPrevious: nav.canGoPrevious,
    canGoNext: Boolean(nextPageToken),
    onPrevious: nav.goPrevious,
    onNext: () => {
      if (nextPageToken) {
        nav.goNext(nextPageToken);
      }
    },
  };
}

export function CursorPaginator({
  canGoPrevious,
  canGoNext,
  onPrevious,
  onNext,
}: {
  canGoPrevious: boolean;
  canGoNext: boolean;
  onPrevious: () => void;
  onNext: () => void;
}) {
  if (!canGoPrevious && !canGoNext) {
    return null;
  }

  return (
    <div className="flex items-center justify-center gap-2">
      <Button
        type="button"
        variant="outline"
        disabled={!canGoPrevious}
        onClick={onPrevious}
      >
        Previous
      </Button>
      <Button
        type="button"
        variant="outline"
        disabled={!canGoNext}
        onClick={onNext}
      >
        Next
      </Button>
    </div>
  );
}
