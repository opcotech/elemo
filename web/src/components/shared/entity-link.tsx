import type { BoxIcon } from "lucide-react";
import {
  BlocksIcon,
  BookOpenIcon,
  Building2Icon,
  FileTextIcon,
  FolderKanbanIcon,
  Link2Icon,
  ListTodoIcon,
  UserIcon,
} from "lucide-react";
import type { ComponentProps, ReactNode } from "react";

import { InternalLink } from "@/components/ui/internal-link";
import {
  Item,
  ItemContent,
  ItemGroup,
  ItemMedia,
  ItemTitle,
} from "@/components/ui/item";
import type { AppEntityType } from "@/lib/entity-types";
import { internalPath } from "@/lib/internal-url";
import { cn } from "@/lib/utils";

export type { AppEntityType } from "@/lib/entity-types";
export { entityHref } from "@/lib/entity-types";

const entityIcons: Record<AppEntityType, typeof BoxIcon> = {
  organization: Building2Icon,
  namespace: BlocksIcon,
  project: FolderKanbanIcon,
  "work-item": ListTodoIcon,
  document: FileTextIcon,
  person: UserIcon,
  relation: Link2Icon,
  "saved-view": BookOpenIcon,
};

export function EntityIcon({
  type,
  className,
}: {
  type: AppEntityType;
  className?: string;
}) {
  const Icon = entityIcons[type];
  return <Icon aria-hidden className={cn("size-4 shrink-0", className)} />;
}

export function EntityLink({
  href,
  type,
  title,
  subtitle,
  imageUrl,
  className,
}: {
  href: string;
  type: AppEntityType;
  title: ReactNode;
  subtitle?: ReactNode;
  imageUrl?: string | null;
  className?: string;
}) {
  return (
    <Item
      role="listitem"
      size="sm"
      className={cn("group/entity min-w-0 p-0", className)}
    >
      <InternalLink
        to={internalPath(href)}
        className="flex w-full min-w-0 items-center gap-2.5 px-3 py-2.5"
      >
        {imageUrl ? (
          <ItemMedia variant="image">
            <img src={imageUrl} alt="" />
          </ItemMedia>
        ) : (
          <ItemMedia
            variant="icon"
            className="bg-muted text-muted-foreground size-8 rounded-lg"
          >
            <EntityIcon type={type} />
          </ItemMedia>
        )}
        <ItemContent className="min-w-0">
          <ItemTitle className="group-hover/entity:text-primary block max-w-full truncate">
            {title}
          </ItemTitle>
          {subtitle && (
            <span className="text-muted-foreground block truncate text-xs">
              {subtitle}
            </span>
          )}
        </ItemContent>
      </InternalLink>
    </Item>
  );
}

export function AppList({
  children,
  className,
  ...props
}: {
  children: ReactNode;
  className?: string;
} & ComponentProps<"div">) {
  return (
    <ItemGroup variant="outline" role="list" className={className} {...props}>
      {children}
    </ItemGroup>
  );
}
