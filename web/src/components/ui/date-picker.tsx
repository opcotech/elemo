"use client";

import { format } from "date-fns";
import { CalendarIcon } from "lucide-react";
import { Suspense, lazy, useState } from "react";
import type { ComponentProps } from "react";
import type { Matcher } from "react-day-picker";

import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  RemovableInputGroup,
  RemovableInputGroupRemove,
} from "@/components/ui/removable-input-group";
import { cn } from "@/lib/utils";

const Calendar = lazy(() =>
  import("@/components/ui/calendar").then((module) => ({
    default: module.Calendar,
  }))
);

interface DatePickerProps extends Pick<
  ComponentProps<typeof Button>,
  "id" | "aria-describedby" | "aria-invalid" | "aria-label"
> {
  date?: Date | null;
  onDateChange?: (date: Date | null) => void;
  disabled?: boolean;
  placeholder?: string;
  className?: string;
  disabledDays?: Matcher | Matcher[];
  clearable?: boolean;
  clearAriaLabel?: string;
}

export function DatePicker({
  date,
  onDateChange,
  disabled = false,
  placeholder = "Pick a date",
  className,
  disabledDays = [],
  clearable = false,
  clearAriaLabel = "Clear date",
  id,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
  "aria-label": ariaLabel,
}: DatePickerProps) {
  const [open, setOpen] = useState(false);

  const picker = (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            id={id}
            aria-describedby={ariaDescribedBy}
            aria-invalid={ariaInvalid}
            aria-label={ariaLabel}
            variant="outline"
            className={cn(
              "border-border bg-card hover:bg-card dark:bg-input dark:hover:bg-input/80 h-9 w-full justify-start rounded-md border shadow-none",
              !className && "font-normal",
              !date && "text-muted-foreground",
              clearable &&
                "group-has-[[data-slot=input-group-remove]:hover]/input-group:text-destructive h-full min-w-0 flex-1 border-0 bg-transparent px-0 shadow-none hover:bg-transparent dark:bg-transparent dark:hover:bg-transparent",
              !clearable && className
            )}
            disabled={disabled}
          />
        }
      >
        <CalendarIcon className="size-4" />
        {date ? format(date, "PPP") : placeholder}
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        {open ? (
          <Suspense
            fallback={
              <div className="text-muted-foreground p-4 text-sm">Loading…</div>
            }
          >
            <Calendar
              mode="single"
              selected={date || undefined}
              onSelect={(selectedDate) => {
                onDateChange?.(selectedDate || null);
                setOpen(false);
              }}
              disabled={disabledDays}
            />
          </Suspense>
        ) : null}
      </PopoverContent>
    </Popover>
  );

  if (!clearable) {
    return picker;
  }

  return (
    <RemovableInputGroup size="default" className={cn("h-9", className)}>
      {picker}
      {date ? (
        <RemovableInputGroupRemove
          addonClassName="pr-0.5"
          disabled={disabled}
          aria-label={clearAriaLabel}
          title={clearAriaLabel}
          onClick={() => {
            onDateChange?.(null);
          }}
        />
      ) : null}
    </RemovableInputGroup>
  );
}
