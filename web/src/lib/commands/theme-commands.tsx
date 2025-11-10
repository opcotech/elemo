import { Laptop, Moon, Palette, Sun } from "lucide-react";

import { commandRegistry } from "./registry";

export function registerThemeCommands(
  onToggleTheme: () => void,
  onSetLightTheme: () => void,
  onSetDarkTheme: () => void,
  onSetSystemTheme: () => void
): void {
  const themeCommands = [
    {
      id: "toggle-theme",
      title: "Toggle Theme",
      description: "Switch between light and dark themes",
      icon: <Palette className="size-4" />,
      shortcut: [],
      keywords: ["theme", "dark", "light", "toggle", "switch", "mode"],
      category: "appearance",
      action: onToggleTheme,
    },
    {
      id: "light-theme",
      title: "Light Theme",
      description: "Switch to light theme",
      icon: <Sun className="size-4" />,
      shortcut: [],
      keywords: ["theme", "light", "bright", "day"],
      category: "appearance",
      action: onSetLightTheme,
    },
    {
      id: "dark-theme",
      title: "Dark Theme",
      description: "Switch to dark theme",
      icon: <Moon className="size-4" />,
      shortcut: [],
      keywords: ["theme", "dark", "night", "dim"],
      category: "appearance",
      action: onSetDarkTheme,
    },
    {
      id: "system-theme",
      title: "System Theme",
      description: "Use system theme preference",
      icon: <Laptop className="size-4" />,
      shortcut: [],
      keywords: ["theme", "system", "auto", "preference"],
      category: "appearance",
      action: onSetSystemTheme,
    },
  ];

  // Register all theme commands
  themeCommands.forEach((command) => {
    commandRegistry.register(command);
  });
}

export function unregisterThemeCommands(): void {
  commandRegistry.unregister("toggle-theme");
  commandRegistry.unregister("light-theme");
  commandRegistry.unregister("dark-theme");
  commandRegistry.unregister("system-theme");
}
