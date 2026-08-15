import { Badge } from "@/components/ui/badge";
import type { Label } from "@/lib/api/types";
import { cn } from "@/lib/utils";
import { resolveWorkLabels } from "@/lib/work/resolve-work-labels";

export function WorkLabelBadges({
  labelIds,
  labels,
  limit = 3,
  truncate = true,
  className,
}: {
  labelIds: readonly string[];
  labels?: readonly Pick<Label, "id" | "name">[] | null;
  limit?: number;
  truncate?: boolean;
  className?: string;
}) {
  const resolved = resolveWorkLabels(labelIds, labels);

  if (resolved.length === 0) {
    return null;
  }

  const visible = resolved.slice(0, limit);
  const overflow = resolved.length - visible.length;

  return (
    <div
      className={cn(
        "flex max-w-full min-w-0 flex-wrap items-center gap-x-2 gap-y-1.5",
        className
      )}
    >
      {visible.map((label) => (
        <Badge
          key={label.id}
          variant="secondary"
          className={cn("px-1.5", truncate && "max-w-28 truncate")}
        >
          {label.name}
        </Badge>
      ))}
      {overflow > 0 && (
        <span className="text-muted-foreground text-xs">+{overflow}</span>
      )}
    </div>
  );
}
