import { useState } from "react";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";
import { commandRegistry } from "@/lib/commands/registry";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  placeholder?: string;
  emptyText?: string;
}

export function CommandPalette({
  open,
  onOpenChange,
  title = "Quick Actions",
  placeholder = "Type a command or select an action...",
  emptyText = "No commands found.",
}: CommandPaletteProps) {
  const [searchQuery, setSearchQuery] = useState("");

  const handleSelect = (commandId: string) => {
    commandRegistry.execute(commandId);
    onOpenChange(false);
  };

  // Get commands based on search query
  const commands = searchQuery.trim()
    ? commandRegistry.searchCommands(searchQuery)
    : commandRegistry.getCommands();

  // Group commands by category
  const groupedCommands = commands.reduce(
    (groups, command) => {
      if (command.hidden) return groups;

      const category = command.category || "general";
      if (!groups[category]) {
        groups[category] = [];
      }
      groups[category].push(command);
      return groups;
    },
    {} as Record<string, typeof commands>
  );

  const categories = Object.keys(groupedCommands);

  return (
    <CommandDialog open={open} onOpenChange={onOpenChange} title={title}>
      <CommandInput
        placeholder={placeholder}
        value={searchQuery}
        onValueChange={setSearchQuery}
      />
      <CommandList>
        <CommandEmpty>{emptyText}</CommandEmpty>
        {categories.map((category, categoryIndex) => (
          <div key={category}>
            {categoryIndex > 0 && <CommandSeparator />}
            <CommandGroup
              heading={
                category.charAt(0).toUpperCase() +
                category.slice(1).replace(/-/g, " ")
              }
            >
              {groupedCommands[category].map((command) => (
                <CommandItem
                  key={command.id}
                  onSelect={() => handleSelect(command.id)}
                  disabled={command.disabled}
                >
                  {command.icon && (
                    <span className="size-4">{command.icon}</span>
                  )}
                  <span className="flex-1">{command.title}</span>
                  {command.shortcut && (
                    <>
                      {command.shortcut.map((key, index) => (
                        <CommandShortcut key={index}>{key}</CommandShortcut>
                      ))}
                    </>
                  )}
                </CommandItem>
              ))}
            </CommandGroup>
          </div>
        ))}
      </CommandList>
    </CommandDialog>
  );
}
