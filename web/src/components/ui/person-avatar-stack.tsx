import {
  Avatar,
  AvatarFallback,
  AvatarGroup,
  AvatarGroupCount,
  AvatarImage,
} from "@/components/ui/avatar";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

const MAX_VISIBLE_AVATARS = 3;

export interface PersonAvatarStackPerson {
  readonly id: string;
  readonly name: string;
  readonly picture?: string | null;
}

export function PersonAvatarStack({
  people,
  size = "default",
  emptyLabel = "Unassigned",
  showNames = false,
  namesLabel,
  className,
}: {
  people: readonly PersonAvatarStackPerson[];
  size?: "sm" | "default";
  emptyLabel?: string;
  showNames?: boolean;
  namesLabel?: string;
  className?: string;
}) {
  if (people.length === 0) {
    return (
      <span className={cn("text-muted-foreground text-sm", className)}>
        {emptyLabel}
      </span>
    );
  }

  const visible = people.slice(0, MAX_VISIBLE_AVATARS);
  const overflowPeople = people.slice(MAX_VISIBLE_AVATARS);
  const overflow = overflowPeople.length;
  const avatarClassName = size === "sm" ? "size-5" : "size-6";

  return (
    <div
      className={cn("flex min-w-0 items-center gap-1.5", className)}
      data-slot="person-avatar-stack"
    >
      <TooltipProvider delay={300}>
        <AvatarGroup className="-space-x-1.5">
          {visible.map((person) => (
            <Tooltip key={person.id}>
              <TooltipTrigger
                render={<Avatar className={avatarClassName} size="sm" />}
              >
                {person.picture ? (
                  <AvatarImage src={person.picture} alt={person.name} />
                ) : null}
                <AvatarFallback className="text-[0.625rem]">
                  {person.name.slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </TooltipTrigger>
              <TooltipContent>{person.name}</TooltipContent>
            </Tooltip>
          ))}
          {overflow > 0 && (
            <Tooltip>
              <TooltipTrigger
                render={
                  <AvatarGroupCount
                    className={cn(avatarClassName, "text-[0.625rem]")}
                  />
                }
              >
                +{overflow}
              </TooltipTrigger>
              <TooltipContent>
                {overflowPeople.map((person) => person.name).join(", ")}
              </TooltipContent>
            </Tooltip>
          )}
        </AvatarGroup>
      </TooltipProvider>
      {showNames && (
        <span
          className={cn(
            "min-w-0 truncate leading-none font-medium",
            size === "sm" && "text-sm"
          )}
        >
          {namesLabel ??
            (people.length === 1 ? people[0].name : `${people.length} people`)}
        </span>
      )}
    </div>
  );
}
