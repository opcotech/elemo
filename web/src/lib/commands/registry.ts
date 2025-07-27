import type { ReactNode } from "react";

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

  getCommands(): Command[] {
    return Array.from(this.commands.values());
  }

  getCommandsByCategory(category: string): Command[] {
    return this.getCommands().filter((cmd) => cmd.category === category);
  }

  getCategories(): string[] {
    return Array.from(this.categories);
  }

  searchCommands(query: string): Command[] {
    const normalizedQuery = query.toLowerCase().trim();
    if (!normalizedQuery) return this.getCommands();

    return this.getCommands().filter((command) => {
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
