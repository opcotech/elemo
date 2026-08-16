"use client";

import { ChevronDownIcon, ChevronsUpDownIcon } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useState } from "react";
import type { ComponentProps, ReactNode } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "./avatar";
import { Badge } from "./badge";
import { Button } from "./button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "./command";
import { PersonAvatarStack } from "./person-avatar-stack";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select";

import { cn } from "@/lib/utils";

export interface EntitySelectOption {
  value: string;
  title: string;
  description?: string;
  details?: ReactNode;
  searchText?: string;
  avatarSrc?: string | null;
  avatarFallback?: string;
}

export interface EntitySelectProps extends Pick<
  ComponentProps<typeof SelectTrigger>,
  "id" | "aria-describedby" | "aria-invalid" | "aria-label"
> {
  options: EntitySelectOption[];
  value?: string;
  placeholder?: string;
  disabled?: boolean;
  size?: "sm" | "default";
  triggerClassName?: string;
  contentProps?: ComponentProps<typeof SelectContent>;
  onValueChange?: (value: string) => void;
}

export interface SearchableEntitySelectProps extends Pick<
  ComponentProps<typeof Button>,
  "id" | "aria-describedby" | "aria-invalid" | "aria-label"
> {
  options: EntitySelectOption[];
  value?: string;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  size?: "sm" | "default";
  triggerClassName?: string;
  contentClassName?: string;
  onValueChange?: (value: string) => void;
}

export interface EntityMultiSelectProps extends Pick<
  ComponentProps<typeof Button>,
  "id" | "aria-describedby" | "aria-invalid" | "aria-label"
> {
  options: EntitySelectOption[];
  value?: readonly string[];
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  size?: "sm" | "default";
  triggerClassName?: string;
  contentClassName?: string;
  onValueChange?: (value: string[]) => void;
  onOpenChange?: (open: boolean) => void;
}

const MAX_VISIBLE_CHIPS = 2;

function optionHasAvatar(option: EntitySelectOption): boolean {
  return (
    typeof option.avatarSrc === "string" || option.avatarFallback !== undefined
  );
}

export function optionCommandValue(option: EntitySelectOption): string {
  return [option.title, option.description, option.searchText, option.value]
    .filter((part): part is string => Boolean(part))
    .join(" ");
}

function entitySelectTriggerClassName(
  size: "sm" | "default",
  triggerClassName?: string
) {
  return cn(
    "border-border bg-card hover:bg-card dark:bg-input dark:hover:bg-input/80 w-full justify-between font-normal shadow-none",
    size === "sm" ? "h-8 px-2" : "h-9 px-3",
    triggerClassName
  );
}

function OptionContent({
  option,
  size = "default",
  className,
  showDetails = false,
}: {
  option: EntitySelectOption;
  size?: "sm" | "default";
  className?: string;
  showDetails?: boolean;
}) {
  const showAvatar = optionHasAvatar(option);
  const fallbackText =
    option.avatarFallback ?? option.title.slice(0, 2).toUpperCase();
  const details = showDetails ? option.details : undefined;
  const description =
    details == null && option.description && size !== "sm"
      ? option.description
      : undefined;

  return (
    <div
      className={cn(
        "flex items-center text-left",
        size === "sm" ? "gap-1.5" : "gap-2",
        className
      )}
    >
      {showAvatar ? (
        <Avatar className={size === "sm" ? "size-5" : "size-8"}>
          {option.avatarSrc ? (
            <AvatarImage src={option.avatarSrc} alt={option.title} />
          ) : null}
          <AvatarFallback>{fallbackText}</AvatarFallback>
        </Avatar>
      ) : null}
      <div className="flex min-w-0 flex-col">
        <span
          className={cn(
            "truncate leading-none font-medium",
            size === "sm" && "text-sm"
          )}
        >
          {option.title}
        </span>
        {details ? (
          <div className="text-muted-foreground mt-1">{details}</div>
        ) : description ? (
          <span className="text-muted-foreground text-xs">{description}</span>
        ) : null}
      </div>
    </div>
  );
}

function SelectedEntitiesSummary({
  options,
  size = "default",
  placeholder,
}: {
  options: EntitySelectOption[];
  size?: "sm" | "default";
  placeholder: string;
}) {
  if (options.length === 0) {
    return (
      <span className="text-muted-foreground truncate font-normal">
        {placeholder}
      </span>
    );
  }

  if (!options.some(optionHasAvatar)) {
    const visible = options.slice(0, MAX_VISIBLE_CHIPS);
    const overflow = options.length - visible.length;

    return (
      <div className="flex min-w-0 flex-1 items-center gap-1">
        {visible.map((option) => (
          <Badge
            key={option.value}
            variant="secondary"
            className="max-w-24 truncate px-1.5"
          >
            {option.title}
          </Badge>
        ))}
        {overflow > 0 ? (
          <span className="text-muted-foreground text-xs">+{overflow}</span>
        ) : null}
      </div>
    );
  }

  return (
    <PersonAvatarStack
      people={options.map((option) => ({
        id: option.value,
        name: option.title,
        picture: option.avatarSrc,
      }))}
      size={size}
      showNames
      emptyLabel={placeholder}
      namesLabel={
        options.length === 1 ? undefined : `${options.length} selected`
      }
      className="min-w-0 flex-1"
    />
  );
}

function EntityCommandPopover({
  open,
  onOpenChange,
  disabled,
  size = "default",
  triggerClassName,
  contentClassName,
  searchPlaceholder,
  emptyMessage,
  triggerIcon: TriggerIcon,
  trigger,
  children,
  id,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
  "aria-label": ariaLabel,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  disabled?: boolean;
  size?: "sm" | "default";
  triggerClassName?: string;
  contentClassName?: string;
  searchPlaceholder?: string;
  emptyMessage: string;
  triggerIcon: LucideIcon;
  trigger: ReactNode;
  children: ReactNode;
  id?: string;
  "aria-describedby"?: string;
  "aria-invalid"?: ComponentProps<typeof Button>["aria-invalid"];
  "aria-label"?: string;
}) {
  return (
    <Popover open={open} onOpenChange={onOpenChange} modal={false}>
      <PopoverTrigger
        render={
          <Button
            id={id}
            type="button"
            variant="outline"
            size={size}
            disabled={disabled}
            aria-describedby={ariaDescribedBy}
            aria-invalid={ariaInvalid}
            aria-label={ariaLabel}
            aria-expanded={open}
            className={entitySelectTriggerClassName(size, triggerClassName)}
          />
        }
      >
        {trigger}
        <TriggerIcon className="text-muted-foreground pointer-events-none size-4 shrink-0" />
      </PopoverTrigger>
      <PopoverContent
        align="start"
        className={cn("gap-0 p-0", contentClassName)}
      >
        <Command>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            <CommandEmpty>{emptyMessage}</CommandEmpty>
            <CommandGroup>{children}</CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

export function EntitySelect({
  options,
  value,
  placeholder,
  disabled,
  size = "default",
  triggerClassName,
  contentProps,
  onValueChange,
  id,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
  "aria-label": ariaLabel,
}: EntitySelectProps) {
  const selectedOption = value
    ? options.find((option) => option.value === value)
    : undefined;

  return (
    <Select
      value={value}
      onValueChange={(next) => {
        if (next != null) {
          onValueChange?.(next);
        }
      }}
      disabled={disabled}
      items={options.map((option) => ({
        value: option.value,
        label: option.title,
      }))}
    >
      <SelectTrigger
        id={id}
        size={size}
        aria-describedby={ariaDescribedBy}
        aria-invalid={ariaInvalid}
        aria-label={ariaLabel}
        className={cn("w-full", triggerClassName)}
      >
        <SelectValue placeholder={placeholder ?? "Select an option"}>
          {selectedOption ? (
            <OptionContent
              option={selectedOption}
              size={size}
              className="w-full justify-start"
            />
          ) : null}
        </SelectValue>
      </SelectTrigger>
      <SelectContent {...contentProps}>
        {options.map((option) => (
          <SelectItem
            key={option.value}
            value={option.value}
            className={size === "sm" ? "py-1.5" : "py-2"}
          >
            <OptionContent option={option} size={size} />
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

export function SearchableEntitySelect({
  options,
  value,
  placeholder = "Select an option",
  searchPlaceholder = "Search…",
  emptyMessage = "No results found.",
  disabled,
  size = "default",
  triggerClassName,
  contentClassName,
  onValueChange,
  id,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
  "aria-label": ariaLabel,
}: SearchableEntitySelectProps) {
  const [open, setOpen] = useState(false);
  const selectedOption = value
    ? options.find((option) => option.value === value)
    : undefined;

  return (
    <EntityCommandPopover
      open={open}
      onOpenChange={setOpen}
      disabled={disabled}
      size={size}
      triggerClassName={triggerClassName}
      contentClassName={cn("w-(--anchor-width) min-w-72", contentClassName)}
      searchPlaceholder={searchPlaceholder}
      emptyMessage={emptyMessage}
      triggerIcon={ChevronDownIcon}
      trigger={
        selectedOption ? (
          <OptionContent
            option={selectedOption}
            size={size}
            className="min-w-0 flex-1 justify-start"
          />
        ) : (
          <span className="text-muted-foreground truncate font-normal">
            {placeholder}
          </span>
        )
      }
      id={id}
      aria-describedby={ariaDescribedBy}
      aria-invalid={ariaInvalid}
      aria-label={ariaLabel}
    >
      {options.map((option) => (
        <CommandItem
          key={option.value}
          value={optionCommandValue(option)}
          data-checked={option.value === value || undefined}
          onSelect={() => {
            onValueChange?.(option.value);
            setOpen(false);
          }}
          className={size === "sm" ? "py-1.5" : "py-2"}
        >
          <OptionContent option={option} size={size} showDetails />
        </CommandItem>
      ))}
    </EntityCommandPopover>
  );
}

/** Compact multi-entity picker (Popover + Command) for assignee/reviewer rows. */
export function EntityMultiSelect({
  options,
  value = [],
  placeholder = "Select…",
  searchPlaceholder = "Search…",
  emptyMessage = "No results found.",
  disabled,
  size = "default",
  triggerClassName,
  contentClassName,
  onValueChange,
  onOpenChange,
  id,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
  "aria-label": ariaLabel,
}: EntityMultiSelectProps) {
  const [open, setOpen] = useState(false);
  const selectedSet = new Set(value);
  const selectedOptions = value
    .map((optionId) => options.find((option) => option.value === optionId))
    .filter((option): option is EntitySelectOption => option != null);

  const toggleValue = (nextValue: string) => {
    const current = [...value];
    const next = selectedSet.has(nextValue)
      ? current.filter((optionId) => optionId !== nextValue)
      : [...current, nextValue];
    onValueChange?.(next);
  };

  return (
    <EntityCommandPopover
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        onOpenChange?.(next);
      }}
      disabled={disabled}
      size={size}
      triggerClassName={triggerClassName}
      contentClassName={cn("w-72", contentClassName)}
      searchPlaceholder={searchPlaceholder}
      emptyMessage={emptyMessage}
      triggerIcon={ChevronsUpDownIcon}
      trigger={
        <SelectedEntitiesSummary
          options={selectedOptions}
          size={size}
          placeholder={placeholder}
        />
      }
      id={id}
      aria-describedby={ariaDescribedBy}
      aria-invalid={ariaInvalid}
      aria-label={ariaLabel}
    >
      {options.map((option) => {
        const isSelected = selectedSet.has(option.value);
        return (
          <CommandItem
            key={option.value}
            value={optionCommandValue(option)}
            data-checked={isSelected || undefined}
            onSelect={() => {
              toggleValue(option.value);
            }}
            className={size === "sm" ? "py-1.5" : "py-2"}
          >
            <OptionContent option={option} size={size} showDetails />
          </CommandItem>
        );
      })}
    </EntityCommandPopover>
  );
}
