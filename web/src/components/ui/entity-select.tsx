import type { ComponentProps } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "./avatar";
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
  triggerClassName?: string;
  contentProps?: ComponentProps<typeof SelectContent>;
  onValueChange?: (value: string) => void;
}

const OptionContent = ({
  option,
  className,
}: {
  option: EntitySelectOption;
  className?: string;
}) => {
  const showAvatar =
    typeof option.avatarSrc === "string" || option.avatarFallback !== undefined;
  const fallbackText =
    option.avatarFallback ?? option.title.slice(0, 2).toUpperCase();

  return (
    <div className={cn("flex items-center gap-2 text-left", className)}>
      {showAvatar && (
        <Avatar className="size-8">
          {option.avatarSrc && (
            <AvatarImage src={option.avatarSrc} alt={option.title} />
          )}
          <AvatarFallback>{fallbackText}</AvatarFallback>
        </Avatar>
      )}
      <div className="flex flex-col">
        <span className="leading-none font-medium">{option.title}</span>
        {option.description && (
          <span className="text-muted-foreground text-xs">
            {option.description}
          </span>
        )}
      </div>
    </div>
  );
};

export function EntitySelect({
  options,
  value,
  placeholder,
  disabled,
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
        aria-describedby={ariaDescribedBy}
        aria-invalid={ariaInvalid}
        aria-label={ariaLabel}
        className={cn("w-full", triggerClassName)}
      >
        <SelectValue placeholder={placeholder ?? "Select an option"}>
          {selectedOption ? (
            <OptionContent
              option={selectedOption}
              className="w-full justify-start"
            />
          ) : null}
        </SelectValue>
      </SelectTrigger>
      <SelectContent {...contentProps}>
        {options.map((option) => (
          <SelectItem key={option.value} value={option.value} className="py-2">
            <OptionContent option={option} />
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
