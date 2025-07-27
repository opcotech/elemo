import { useCallback } from "react";

import type { BreadcrumbItem } from "@/components/breadcrumb";
import { useBreadcrumbs } from "@/components/breadcrumb";

export function useBreadcrumbUtils() {
  const { setBreadcrumbs, clearBreadcrumbs } = useBreadcrumbs();

  const setBreadcrumbsFromPath = useCallback(
    (path: string, basePath?: string) => {
      const segments = path.split("/").filter(Boolean);
      const breadcrumbs: BreadcrumbItem[] = [];

      // Add base path if provided
      if (basePath) {
        breadcrumbs.push({
          label: basePath,
          href: "/",
          isNavigatable: true,
        });
      }

      // Build breadcrumbs from path segments
      let currentPath = "";
      segments.forEach((segment, index) => {
        currentPath += `/${segment}`;
        const isLast = index === segments.length - 1;

        breadcrumbs.push({
          label:
            segment.charAt(0).toUpperCase() +
            segment.slice(1).replace(/-/g, " "),
          href: isLast ? undefined : currentPath,
          isNavigatable: !isLast,
        });
      });

      setBreadcrumbs(breadcrumbs);
    },
    [setBreadcrumbs]
  );

  const setBreadcrumbsFromItems = useCallback(
    (items: BreadcrumbItem[]) => {
      setBreadcrumbs(items);
    },
    [setBreadcrumbs]
  );

  const addBreadcrumb = useCallback(
    (item: BreadcrumbItem) => {
      setBreadcrumbs((prev) => [...prev, item]);
    },
    [setBreadcrumbs]
  );

  const removeLastBreadcrumb = useCallback(() => {
    setBreadcrumbs((prev) => prev.slice(0, -1));
  }, [setBreadcrumbs]);

  return {
    setBreadcrumbsFromPath,
    setBreadcrumbsFromItems,
    addBreadcrumb,
    removeLastBreadcrumb,
    clearBreadcrumbs,
  };
}
