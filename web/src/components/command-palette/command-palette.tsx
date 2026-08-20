import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { EntityIcon } from "@/components/shared/entity-link";
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
import { Spinner } from "@/components/ui/spinner";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useNavigationContext } from "@/hooks/use-navigation-context";
import { v1SearchGetOptions } from "@/lib/api/query-options";
import type { Command, CommandContext } from "@/lib/commands/registry";
import { internalPath } from "@/lib/internal-url";
import {
  SEARCH_DEBOUNCE_MS,
  SEARCH_PALETTE_PAGE_SIZE,
} from "@/lib/search/params";
import { searchResultEntityType, searchResultHref } from "@/lib/search/result";

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
  const navigate = useNavigate();
  const navigationContext = useNavigationContext();
  const trimmedQuery = searchQuery.trim();
  const debouncedQuery = useDebouncedValue(trimmedQuery, SEARCH_DEBOUNCE_MS);
  const { data, isError, isFetching } = useQuery({
    ...v1SearchGetOptions({
      query: {
        q: debouncedQuery,
        page_size: SEARCH_PALETTE_PAGE_SIZE,
      },
    }),
    enabled: open && debouncedQuery.length > 0,
  });
  const hits = debouncedQuery.length > 0 ? (data?.items ?? []) : [];

  // Use provided context or derive from navigation context
  const context: CommandContext | undefined =
    contextProp ||
    (navigationContext.type === "global"
      ? "global"
      : navigationContext.type === "organization"
        ? "organization"
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

    const query = trimmedQuery.toLowerCase();
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
  const showResults = trimmedQuery.length > 0;
  const searching = showResults && isFetching && hits.length === 0 && !isError;

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title={title}
      shouldFilter={false}
    >
      <CommandInput
        placeholder={placeholder}
        value={searchQuery}
        onValueChange={setSearchQuery}
      />
      <CommandList>
        <CommandEmpty>{emptyText}</CommandEmpty>
        {showResults ? (
          <CommandGroup heading="Results">
            <CommandItem
              value="view-all-results"
              onSelect={() => {
                onOpenChange(false);
                void navigate({
                  to: "/search",
                  search: { q: trimmedQuery, type: "all" },
                });
              }}
            >
              View all results
            </CommandItem>
            {isError ? (
              <CommandItem value="search-error" disabled>
                Search failed. Try again.
              </CommandItem>
            ) : null}
            {searching ? (
              <CommandItem value="search-loading" disabled>
                <Spinner size="xs" />
                Searching…
              </CommandItem>
            ) : null}
            {hits.map((hit) => {
              const href = searchResultHref(hit);
              return (
                <CommandItem
                  key={`${hit.type}:${hit.id}`}
                  value={`result-${hit.type}-${hit.id}`}
                  disabled={!href}
                  onSelect={() => {
                    if (!href) return;
                    onOpenChange(false);
                    void navigate({ to: internalPath(href) as never });
                  }}
                >
                  <EntityIcon type={searchResultEntityType(hit)} />
                  <span className="flex min-w-0 flex-1 flex-col">
                    <span className="truncate">{hit.title}</span>
                    {hit.subtitle ? (
                      <span className="text-muted-foreground truncate text-xs">
                        {hit.subtitle}
                      </span>
                    ) : null}
                  </span>
                </CommandItem>
              );
            })}
          </CommandGroup>
        ) : null}
        {categories.map((category, categoryIndex) => (
          <div key={category}>
            {(showResults || categoryIndex > 0) && <CommandSeparator />}
            <CommandGroup
              heading={
                category.charAt(0).toUpperCase() +
                category.slice(1).replace(/-/g, " ")
              }
            >
              {groupedCommands[category].map((command) => (
                <CommandItem
                  key={command.id}
                  value={command.id}
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
