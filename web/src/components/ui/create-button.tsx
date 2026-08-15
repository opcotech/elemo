import { PlusIcon } from "lucide-react";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { InternalLink } from "@/components/ui/internal-link";
import { internalPath } from "@/lib/internal-url";

export function CreateButton({
  children = "Create",
  onClick,
  href,
}: {
  children?: ReactNode;
  onClick?: () => void;
  href?: string;
}) {
  return (
    <Button
      size="sm"
      onClick={onClick}
      render={href ? <InternalLink to={internalPath(href)} /> : undefined}
    >
      <PlusIcon />
      {children}
    </Button>
  );
}
