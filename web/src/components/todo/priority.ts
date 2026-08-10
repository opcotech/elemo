import type { StatusIndicatorTone } from "@/components/shared/status-indicator";
import type { TodoPriority } from "@/lib/api/types";

export const todoPriorityTone: Record<TodoPriority, StatusIndicatorTone> = {
  critical: "danger",
  urgent: "warning",
  important: "primary",
  normal: "neutral",
};
