import { OrganizationCardSkeleton } from "./organization-card-skeleton";

export function OrganizationListSkeleton({ count = 6 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, i) => (
        <OrganizationCardSkeleton key={i} />
      ))}
    </div>
  );
}
