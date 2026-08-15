import { Item, ItemContent, ItemGroup, ItemMedia } from "@/components/ui/item";
import { Skeleton } from "@/components/ui/skeleton";

export function ListSkeleton({
  rows = 6,
  className,
}: {
  rows?: number;
  className?: string;
}) {
  return (
    <ItemGroup
      variant="outline"
      role="status"
      aria-busy="true"
      className={className}
    >
      <span className="sr-only">Loading</span>
      {Array.from({ length: rows }, (_, index) => (
        <Item key={index} size="sm" className="min-w-0 p-0">
          <div className="flex w-full min-w-0 items-center gap-2.5 px-3 py-2.5">
            <ItemMedia variant="icon" className="size-8 rounded-lg p-0">
              <Skeleton className="size-8 rounded-lg" />
            </ItemMedia>
            <ItemContent className="min-w-0">
              <Skeleton className="h-3 w-24" />
              <Skeleton className="h-4 w-40" />
            </ItemContent>
          </div>
        </Item>
      ))}
    </ItemGroup>
  );
}
