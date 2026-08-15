import {
  BugIcon,
  CircleDotIcon,
  CrownIcon,
  SquareCheckIcon,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { PropertyRibbon } from "./property-ribbon";

import type { IssueKind } from "@/lib/api/types";
import { cn } from "@/lib/utils";

const kindIcons: Record<IssueKind, LucideIcon> = {
  epic: CrownIcon,
  story: CircleDotIcon,
  task: SquareCheckIcon,
  bug: BugIcon,
};

const kindToneClassName: Record<IssueKind, string> = {
  epic: "text-purple-500",
  story: "text-success",
  task: "text-primary",
  bug: "text-destructive",
};

const kindLabels: Record<IssueKind, string> = {
  epic: "Epic",
  story: "Story",
  task: "Task",
  bug: "Bug",
};

export const issueKinds: readonly IssueKind[] = [
  "epic",
  "story",
  "task",
  "bug",
];

export function KindRibbon({
  kind,
  className,
  labelClassName,
  showLabel = true,
}: {
  kind: IssueKind;
  className?: string;
  labelClassName?: string;
  showLabel?: boolean;
}) {
  return (
    <PropertyRibbon
      icon={kindIcons[kind]}
      label={kindLabels[kind]}
      className={cn(kindToneClassName[kind], className)}
      labelClassName={labelClassName}
      showLabel={showLabel}
      data-slot="kind-ribbon"
      data-kind={kind}
    />
  );
}

export {
  kindLabels as issueKindLabels,
  kindToneClassName as issueKindToneClassName,
};
