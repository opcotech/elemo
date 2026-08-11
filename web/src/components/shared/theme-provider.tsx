"use client";

import * as React from "react";
import { createContext, useContext, useEffect, useState } from "react";

import type { Theme } from "@/lib/theme";

type ThemeProviderProps = {
  children: React.ReactNode;
  initialTheme?: Theme;
};

type ThemeProviderState = {
  theme: Theme;
  setTheme: (theme: Theme) => void;
};

const initialState: ThemeProviderState = {
  theme: "system",
  setTheme: () => null,
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

export function ThemeProvider({
  children,
  initialTheme = "system",
  ...props
}: ThemeProviderProps) {
  const [theme, setThemeState] = useState<Theme>(initialTheme);

  useEffect(() => {
    const root = window.document.documentElement;
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const applyTheme = () => {
      root.classList.remove("light", "dark");
      root.classList.add(
        theme === "system" ? (media.matches ? "dark" : "light") : theme
      );
    };

    applyTheme();
    media.addEventListener("change", applyTheme);
    return () => media.removeEventListener("change", applyTheme);
  }, [theme]);

  const value = {
    theme,
    setTheme: (nextTheme: Theme) => {
      setThemeState(nextTheme);
      void import("@/lib/theme").then(({ setThemeFn }) =>
        setThemeFn({ data: nextTheme })
      );
    },
  };

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext);

  if (context === undefined)
    throw new Error("useTheme must be used within a ThemeProvider");

  return context;
};
