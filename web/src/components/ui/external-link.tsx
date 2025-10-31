import { ExternalLinkIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export function ExternalLink({
  href,
  children,
  className,
}: {
  href: string;
  children?: React.ReactNode | string;
  className?: string;
}) {
  return (
    <div className="flex items-center gap-1">
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className={cn("text-primary hover:underline", className)}
      >
        {children ?? href}
      </a>
      <ExternalLinkIcon className="text-primary size-4" />
    </div>
  );
}
