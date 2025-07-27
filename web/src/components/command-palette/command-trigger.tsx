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
      <span className="flex-1 text-left text-sm font-normal">
        {placeholder}
      </span>
      <kbd className="text-muted-foreground pointer-events-none ml-auto space-x-0.5 rounded border bg-neutral-100 px-1.5 py-0.5 text-xs font-medium opacity-100 select-none">
        <span>⌘</span>
        <span>K</span>
      </kbd>
    </SidebarInputButton>
  );
}
