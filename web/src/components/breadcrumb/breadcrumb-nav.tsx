import { useMatches, useRouterState } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";
import { Fragment, useMemo } from "react";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { InternalLink } from "@/components/ui/internal-link";
import { resolveRouteBreadcrumb } from "@/lib/breadcrumb";
import { internalPath } from "@/lib/internal-url";

interface BreadcrumbNavProps {
  className?: string;
}

type NavCrumb = {
  label: string;
  href?: string;
  isNavigatable?: boolean;
};

function crumbFromRouteBreadcrumb(
  breadcrumb: ReturnType<typeof resolveRouteBreadcrumb>,
  matchPathname: string
): NavCrumb {
  if (typeof breadcrumb === "string") {
    return {
      label: breadcrumb,
      href: matchPathname,
      isNavigatable: true,
    };
  }

  if (breadcrumb.href === null) {
    return {
      label: breadcrumb.label,
      isNavigatable: false,
    };
  }

  return {
    label: breadcrumb.label,
    href: breadcrumb.href ?? matchPathname,
    isNavigatable: true,
  };
}

export function BreadcrumbNav({ className }: BreadcrumbNavProps) {
  const currentPath = useRouterState({
    select: (state) => state.location.pathname,
  });
  const matchCrumbs = useMatches({
    select: (matches) =>
      matches.flatMap((match) => {
        const breadcrumb = match.staticData.breadcrumb;
        if (!breadcrumb) {
          return [];
        }

        return [
          crumbFromRouteBreadcrumb(
            resolveRouteBreadcrumb(breadcrumb, match.loaderData),
            match.pathname
          ),
        ];
      }),
  });

  const breadcrumbs = useMemo(() => {
    const normalizedCrumbs = matchCrumbs.filter(
      (item, index, items) =>
        index ===
        items.findIndex(
          (candidate) =>
            candidate.label === item.label && candidate.href === item.href
        )
    );

    if (currentPath === "/") {
      return normalizedCrumbs.length > 0
        ? normalizedCrumbs
        : [{ label: "Home" }];
    }

    if (normalizedCrumbs[0]?.label !== "Home") {
      return [
        {
          label: "Home",
          href: "/",
          isNavigatable: true,
        },
        ...normalizedCrumbs,
      ];
    }

    return normalizedCrumbs;
  }, [currentPath, matchCrumbs]);

  if (breadcrumbs.length === 0) {
    return null;
  }

  return (
    <Breadcrumb className={className}>
      <BreadcrumbList className="flex-nowrap overflow-hidden">
        {breadcrumbs.map((item, index) => {
          const isLast = index === breadcrumbs.length - 1;
          const showLink = !isLast && item.isNavigatable !== false && item.href;

          return (
            <Fragment key={`${item.label}-${index}`}>
              <BreadcrumbItem className="min-w-0">
                {showLink && item.href ? (
                  <BreadcrumbLink
                    render={<InternalLink to={internalPath(item.href)} />}
                    className="hover:text-primary max-w-40 truncate transition-colors duration-150"
                  >
                    {item.label}
                  </BreadcrumbLink>
                ) : (
                  <BreadcrumbPage className="max-w-56 truncate">
                    {item.label}
                  </BreadcrumbPage>
                )}
              </BreadcrumbItem>
              {!isLast && (
                <BreadcrumbSeparator>
                  <ChevronRight />
                </BreadcrumbSeparator>
              )}
            </Fragment>
          );
        })}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
