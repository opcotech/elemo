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

export interface EntitySelectProps {
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
}: EntitySelectProps) {
  const selectedOption = value
    ? options.find((option) => option.value === value)
    : undefined;

  return (
    <Select value={value} onValueChange={onValueChange} disabled={disabled}>
      <SelectTrigger className={cn("w-full", triggerClassName)}>
        {selectedOption ? (
          <SelectValue asChild>
            <OptionContent
              option={selectedOption}
              className="w-full justify-start"
            />
          </SelectValue>
        ) : (
          <SelectValue
            placeholder={placeholder ?? "Select an option"}
            className="text-muted-foreground"
          />
        )}
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
