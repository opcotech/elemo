import { Link2Icon, PlusIcon } from "lucide-react";
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

export function AddButton({
  onClick,
  disabled,
  size = "xs",
}: {
  onClick?: () => void;
  disabled?: boolean;
  size?: "xs" | "sm";
}) {
  return (
    <Button
      type="button"
      variant="ghost"
      size={size}
      disabled={disabled}
      onClick={onClick}
    >
      <PlusIcon />
      Add
    </Button>
  );
}

export function LinkButton({
  onClick,
  disabled,
  size = "xs",
}: {
  onClick?: () => void;
  disabled?: boolean;
  size?: "xs" | "sm";
}) {
  return (
    <Button
      type="button"
      variant="ghost"
      size={size}
      disabled={disabled}
      onClick={onClick}
    >
      <Link2Icon />
      Link
    </Button>
  );
}
