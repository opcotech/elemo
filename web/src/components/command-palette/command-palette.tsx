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
import { useNavigationContext } from "@/hooks/use-navigation-context";
import type { Command, CommandContext } from "@/lib/commands/registry";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  placeholder?: string;
  emptyText?: string;
  context?: CommandContext;
  commands: Command[];
}

export function CommandPalette({
  open,
  onOpenChange,
  title = "Quick Actions",
  placeholder = "Type a command or select an action...",
  emptyText = "No commands found.",
  context: contextProp,
  commands: availableCommands,
}: CommandPaletteProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const navigationContext = useNavigationContext();

  // Use provided context or derive from navigation context
  const context: CommandContext | undefined =
    contextProp ||
    (navigationContext.type === "global"
      ? "global"
      : navigationContext.type === "namespace"
        ? "namespace"
        : navigationContext.type === "project"
          ? "project"
          : undefined);

  const handleSelect = (commandId: string) => {
    availableCommands.find((command) => command.id === commandId)?.action();
    onOpenChange(false);
  };

  const commands = availableCommands.filter((command) => {
    const isInContext =
      !context ||
      !command.context ||
      (Array.isArray(command.context)
        ? command.context.includes(context)
        : command.context === context);
    if (!isInContext) return false;

    const query = searchQuery.toLowerCase().trim();
    if (!query) return true;
    return (
      command.title.toLowerCase().includes(query) ||
      command.description?.toLowerCase().includes(query) ||
      command.keywords?.some((keyword) => keyword.toLowerCase().includes(query))
    );
  });

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
