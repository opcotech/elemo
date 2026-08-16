"use client";

import { ChevronDownIcon, XIcon } from "lucide-react";
import type { ComponentProps, ReactNode } from "react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
} from "@/components/ui/input-group";
import { cn } from "@/lib/utils";

const removableInputGroupDangerClassName = cn(
  "has-[[data-slot=input-group-remove]:hover]:border-destructive/40",
  "has-[[data-slot=input-group-remove]:hover]:bg-destructive/10",
  "has-[[data-slot=input-group-remove]:hover]:text-destructive",
  "has-[[data-slot=input-group-remove]:hover]:ring-1",
  "has-[[data-slot=input-group-remove]:hover]:ring-inset",
  "has-[[data-slot=input-group-remove]:hover]:ring-destructive/20",
  "has-[[data-slot=input-group-remove]:hover]:hover:border-destructive/40",
  "has-[[data-slot=input-group-remove]:hover]:hover:bg-destructive/10",
  "has-[[data-slot=input-group-remove]:hover]:hover:text-destructive",
  "dark:has-[[data-slot=input-group-remove]:hover]:hover:border-destructive/40",
  "dark:has-[[data-slot=input-group-remove]:hover]:hover:bg-destructive/10",
  "has-[[data-slot=input-group-remove]:hover]:**:data-[slot=dropdown-menu-trigger]:text-destructive",
  "has-[[data-slot=input-group-remove]:hover]:**:data-[slot=input-group-content]:text-destructive",
  "has-[[data-slot=input-group-remove]:hover]:**:data-[slot=input-group-content]:*:text-destructive",
  "has-[[data-slot=input-group-remove]:hover]:[&_svg]:text-destructive"
);

function RemovableInputGroup({
  className,
  size = "sm",
  ...props
}: ComponentProps<typeof InputGroup> & { size?: "sm" | "default" }) {
  return (
    <InputGroup
      data-size={size}
      className={cn(
        size === "sm" && "h-8",
        removableInputGroupDangerClassName,
        className
      )}
      {...props}
    />
  );
}

function RemovableInputGroupDropdown({
  label,
  children,
  disabled,
  align = "inline-start",
  contentAlign = "start",
  "aria-label": ariaLabel,
  className,
  buttonClassName,
}: {
  label: ReactNode;
  children: ReactNode;
  disabled?: boolean;
  align?: ComponentProps<typeof InputGroupAddon>["align"];
  contentAlign?: ComponentProps<typeof DropdownMenuContent>["align"];
  "aria-label"?: string;
  className?: string;
  buttonClassName?: string;
}) {
  return (
    <InputGroupAddon align={align} className={className}>
      <DropdownMenu disabled={disabled}>
        <DropdownMenuTrigger
          render={
            <InputGroupButton
              variant="ghost"
              aria-label={ariaLabel}
              className={cn("pr-1.5! text-xs", buttonClassName)}
            />
          }
        >
          {label}
          <ChevronDownIcon className="size-3" />
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align={contentAlign}
          className="[--radius:0.95rem]"
        >
          {children}
        </DropdownMenuContent>
      </DropdownMenu>
    </InputGroupAddon>
  );
}

function RemovableInputGroupContent({
  className,
  ...props
}: ComponentProps<"div">) {
  return (
    <div
      data-slot="input-group-content"
      className={cn("flex min-w-0 flex-1 items-center px-2 text-xs", className)}
      {...props}
    />
  );
}

function RemovableInputGroupRemove({
  addonClassName,
  className,
  children,
  size = "icon-xs",
  variant = "ghost",
  ...props
}: ComponentProps<typeof InputGroupButton> & { addonClassName?: string }) {
  return (
    <InputGroupAddon
      align="inline-end"
      data-slot="input-group-remove"
      className={cn("cursor-default py-0 pr-2", addonClassName)}
    >
      <InputGroupButton
        size={size}
        variant={variant}
        className={cn(
          "hover:text-destructive hover:bg-transparent hover:ring-0",
          className
        )}
        {...props}
      >
        {children ?? <XIcon />}
      </InputGroupButton>
    </InputGroupAddon>
  );
}

export {
  RemovableInputGroup,
  RemovableInputGroupContent,
  RemovableInputGroupDropdown,
  RemovableInputGroupRemove,
  removableInputGroupDangerClassName,
};
