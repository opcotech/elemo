import type { ReactNode } from "react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function DetailCard({
  title,
  description,
  actions,
  children,
  className,
  "data-section": dataSection,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
  "data-section"?: string;
}) {
  return (
    <Card data-section={dataSection} className={className}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle>{title}</CardTitle>
            {description && <CardDescription>{description}</CardDescription>}
          </div>
          {actions ? (
            <div className="flex items-center gap-2">{actions}</div>
          ) : null}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">{children}</div>
      </CardContent>
    </Card>
  );
}

export function DetailPageSkeleton({ fields = 5 }: { fields?: number }) {
  return (
    <div className="space-y-6">
      <div className="mb-6">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="mt-2 h-5 w-96" />
      </div>
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="mt-2 h-4 w-80" />
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {Array.from({ length: fields }).map((_, index) => (
              <div key={index}>
                <Skeleton className="h-4 w-20" />
                <Skeleton className="mt-1 h-5 w-32" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
