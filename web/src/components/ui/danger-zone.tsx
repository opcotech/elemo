import * as React from "react";

import { cn } from "@/lib/utils";

function DangerZone({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="danger-zone"
      className={cn(
        "group/danger-zone border-destructive bg-destructive/5 text-card-foreground flex flex-col gap-(--danger-zone-spacing) overflow-hidden rounded-xl border py-(--danger-zone-spacing) text-sm [--danger-zone-spacing:--spacing(5)]",
        className
      )}
      {...props}
    />
  );
}

function DangerZoneHeader({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="danger-zone-header"
      className={cn(
        "grid auto-rows-min items-start gap-1.5 px-(--danger-zone-spacing) has-data-[slot=danger-zone-description]:grid-rows-[auto_auto]",
        className
      )}
      {...props}
    />
  );
}

function DangerZoneTitle({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="danger-zone-title"
      className={cn(
        "text-destructive text-base leading-snug font-medium",
        className
      )}
      {...props}
    />
  );
}

function DangerZoneDescription({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="danger-zone-description"
      className={cn("text-muted-foreground text-sm", className)}
      {...props}
    />
  );
}

function DangerZoneContent({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="danger-zone-content"
      className={cn("px-(--danger-zone-spacing)", className)}
      {...props}
    />
  );
}

function DangerZoneActions({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="danger-zone-actions"
      className={cn(
        "flex items-center justify-end gap-2 px-(--danger-zone-spacing)",
        className
      )}
      {...props}
    />
  );
}

export {
  DangerZone,
  DangerZoneHeader,
  DangerZoneTitle,
  DangerZoneDescription,
  DangerZoneContent,
  DangerZoneActions,
};
