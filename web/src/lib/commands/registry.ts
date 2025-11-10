import type { ReactNode } from "react";

export type CommandContext =
  | "global"
  | "namespace"
  | "project"
  | "document"
  | "issue";

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

class CommandRegistry {
  private commands: Map<string, Command> = new Map();
  private categories: Set<string> = new Set();

  register(command: Command): void {
    this.commands.set(command.id, command);
    if (command.category) {
      this.categories.add(command.category);
    }
  }

  unregister(commandId: string): void {
    this.commands.delete(commandId);
  }

  getCommand(commandId: string): Command | undefined {
    return this.commands.get(commandId);
  }

  getCommands(context?: CommandContext): Command[] {
    const allCommands = Array.from(this.commands.values());
    if (!context) {
      return allCommands;
    }
    // Filter commands that are available in the given context
    return allCommands.filter((cmd) => {
      if (!cmd.context) {
        return true;
      }
      if (Array.isArray(cmd.context)) {
        return cmd.context.includes(context);
      }
      return cmd.context === context;
    });
  }

  getCommandsByCategory(category: string, context?: CommandContext): Command[] {
    return this.getCommands(context).filter((cmd) => cmd.category === category);
  }

  getCategories(): string[] {
    return Array.from(this.categories);
  }

  searchCommands(query: string, context?: CommandContext): Command[] {
    const normalizedQuery = query.toLowerCase().trim();
    const commands = this.getCommands(context);
    if (!normalizedQuery) return commands;

    return commands.filter((command) => {
      // Search in title
      if (command.title.toLowerCase().includes(normalizedQuery)) {
        return true;
      }

      // Search in description
      if (command.description?.toLowerCase().includes(normalizedQuery)) {
        return true;
      }

      // Search in keywords
      if (
        command.keywords?.some((keyword) =>
          keyword.toLowerCase().includes(normalizedQuery)
        )
      ) {
        return true;
      }

      return false;
    });
  }

  execute(commandId: string): void {
    const command = this.getCommand(commandId);
    if (command && !command.disabled) {
      command.action();
    }
  }

  clear(): void {
    this.commands.clear();
    this.categories.clear();
  }
}

// Export singleton instance
export const commandRegistry = new CommandRegistry();
