import { SearchIcon } from "lucide-react";

import { SidebarInputButton } from "@/components/ui/sidebar";

interface CommandTriggerProps {
  onOpen: () => void;
  placeholder?: string;
  className?: string;
}

export function CommandTrigger({
  onOpen,
  placeholder = "Search or jump to...",
  className,
}: CommandTriggerProps) {
  return (
    <SidebarInputButton
      onClick={onOpen}
      aria-label="Open command palette"
      className={className}
    >
      <SearchIcon className="h-4 w-4 opacity-50" />
      <span className="flex-1 text-left font-normal text-sm">
        {placeholder}
      </span>
      <kbd className="pointer-events-none ml-auto select-none space-x-0.5 rounded border bg-muted px-1.5 py-0.5 font-medium text-muted-foreground text-xs opacity-100">
        <span>⌘</span>
        <span>K</span>
      </kbd>
    </SidebarInputButton>
  );
}
