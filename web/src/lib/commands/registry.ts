import type { ReactNode } from "react";

export type CommandContext =
  "global" | "organization" | "namespace" | "project";

export interface Command {
  id: string;
  title: string;
  description?: string;
  icon?: ReactNode;
  shortcut?: string[];
  keywords?: string[];
  category?: string;
  disabled?: boolean;
  hidden?: boolean;
  context?: CommandContext | CommandContext[];
  action: () => void;
}
