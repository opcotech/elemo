import {
  ChevronUpIcon,
  ChevronsUpIcon,
  EqualIcon,
  MinusIcon,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import type { TodoPriority } from "@/lib/api/types";

export const todoPriorities = [
  "normal",
  "important",
  "urgent",
  "critical",
] as const satisfies readonly TodoPriority[];

export const todoPriorityLabels: Record<TodoPriority, string> = {
  normal: "Normal",
  important: "Important",
  urgent: "Urgent",
  critical: "Critical",
};

export const todoPriorityIcons: Record<TodoPriority, LucideIcon> = {
  normal: MinusIcon,
  important: EqualIcon,
  urgent: ChevronUpIcon,
  critical: ChevronsUpIcon,
};

export const todoPriorityToneClassName: Record<TodoPriority, string> = {
  normal: "text-muted-foreground",
  important: "text-primary",
  urgent: "text-warning",
  critical: "text-destructive",
};
