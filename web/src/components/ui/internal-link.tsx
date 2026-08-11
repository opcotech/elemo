import { Link } from "@tanstack/react-router";
import type { ComponentProps } from "react";

import type { InternalPath } from "@/lib/internal-url";

type RouterLinkProps = ComponentProps<typeof Link>;

export interface InternalLinkProps extends Omit<
  RouterLinkProps,
  "href" | "to" | "search"
> {
  to: InternalPath;
  // Widened `to` loses route search inference; accept plain search objects.
  search?: Record<string, unknown>;
}

export function InternalLink({ to, search, ...props }: InternalLinkProps) {
  return (
    <Link
      {...props}
      to={to as never}
      {...(search !== undefined ? { search: search as never } : {})}
    />
  );
}
