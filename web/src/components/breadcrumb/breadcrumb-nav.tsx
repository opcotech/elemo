import { ChevronRight } from "lucide-react";

import { useBreadcrumbs } from "./breadcrumb-provider";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

interface BreadcrumbNavProps {
  className?: string;
}

export function BreadcrumbNav({ className }: BreadcrumbNavProps) {
  const { breadcrumbs } = useBreadcrumbs();

  if (breadcrumbs.length === 0) {
    return null;
  }

  return (
    <Breadcrumb className={className}>
      <BreadcrumbList>
        {breadcrumbs.map((item, index) => (
          <div key={index} className="flex items-center">
            <BreadcrumbItem>
              {item.isNavigatable !== false && item.href ? (
                <BreadcrumbLink
                  href={item.href}
                  className="hover:text-primary transition-colors duration-200"
                >
                  {item.label}
                </BreadcrumbLink>
              ) : (
                <BreadcrumbPage>{item.label}</BreadcrumbPage>
              )}
            </BreadcrumbItem>
            {index < breadcrumbs.length - 1 && (
              <BreadcrumbSeparator>
                <ChevronRight className="h-4 w-4" />
              </BreadcrumbSeparator>
            )}
          </div>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
