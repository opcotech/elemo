/**
 * Elemo Theme Configuration
 *
 * Single elegant blue theme with dark/light mode support
 * Designed for modern project management interfaces
 */

export const themeConfig = {
  // Base design tokens
  borderRadius: "0.5rem",

  // Color definitions using OKLCH for better color consistency
  colors: {
    light: {
      // Core surfaces
      background: "oklch(99.5% 0.002 247)",
      foreground: "oklch(15% 0.004 247)",
      card: "oklch(100% 0 0)",
      cardForeground: "oklch(15% 0.004 247)",
      popover: "oklch(100% 0 0)",
      popoverForeground: "oklch(15% 0.004 247)",

      // Primary blue - sophisticated and confident
      primary: "oklch(55% 0.15 247)",
      primaryForeground: "oklch(98% 0.01 247)",

      // Secondary grays
      secondary: "oklch(96% 0.005 247)",
      secondaryForeground: "oklch(25% 0.01 247)",
      muted: "oklch(97% 0.005 247)",
      mutedForeground: "oklch(45% 0.01 247)",
      accent: "oklch(94% 0.02 247)",
      accentForeground: "oklch(30% 0.05 247)",

      // Semantic colors
      destructive: "oklch(65% 0.2 15)",
      destructiveForeground: "oklch(98% 0.01 15)",
      success: "oklch(70% 0.15 150)",
      successForeground: "oklch(98% 0.01 150)",
      warning: "oklch(75% 0.15 85)",
      warningForeground: "oklch(15% 0.05 85)",
      info: "oklch(70% 0.12 220)",
      infoForeground: "oklch(98% 0.01 220)",

      // UI elements
      border: "oklch(90% 0.005 247)",
      input: "oklch(94% 0.005 247)",
      ring: "oklch(55% 0.15 247)",

      // Chart colors for data visualization
      chart1: "oklch(55% 0.15 247)",
      chart2: "oklch(65% 0.15 170)",
      chart3: "oklch(70% 0.15 85)",
      chart4: "oklch(60% 0.15 300)",
      chart5: "oklch(65% 0.15 15)",

      // Sidebar navigation
      sidebar: "oklch(98% 0.005 247)",
      sidebarForeground: "oklch(20% 0.01 247)",
      sidebarPrimary: "oklch(55% 0.15 247)",
      sidebarPrimaryForeground: "oklch(98% 0.01 247)",
      sidebarAccent: "oklch(94% 0.02 247)",
      sidebarAccentForeground: "oklch(30% 0.05 247)",
      sidebarBorder: "oklch(92% 0.005 247)",
      sidebarRing: "oklch(55% 0.15 247)",
    },

    dark: {
      // Core surfaces - deep, sophisticated darks
      background: "oklch(8% 0.01 247)",
      foreground: "oklch(95% 0.01 247)",
      card: "oklch(12% 0.01 247)",
      cardForeground: "oklch(95% 0.01 247)",
      popover: "oklch(12% 0.01 247)",
      popoverForeground: "oklch(95% 0.01 247)",

      // Primary blue - brighter for dark backgrounds
      primary: "oklch(70% 0.15 247)",
      primaryForeground: "oklch(8% 0.01 247)",

      // Secondary grays for dark mode
      secondary: "oklch(18% 0.01 247)",
      secondaryForeground: "oklch(85% 0.01 247)",
      muted: "oklch(16% 0.01 247)",
      mutedForeground: "oklch(65% 0.01 247)",
      accent: "oklch(20% 0.02 247)",
      accentForeground: "oklch(80% 0.05 247)",

      // Semantic colors for dark mode
      destructive: "oklch(70% 0.18 15)",
      destructiveForeground: "oklch(95% 0.01 15)",
      success: "oklch(75% 0.15 150)",
      successForeground: "oklch(8% 0.01 150)",
      warning: "oklch(80% 0.15 85)",
      warningForeground: "oklch(8% 0.01 85)",
      info: "oklch(75% 0.12 220)",
      infoForeground: "oklch(8% 0.01 220)",

      // UI elements for dark mode
      border: "oklch(25% 0.01 247)",
      input: "oklch(20% 0.01 247)",
      ring: "oklch(70% 0.15 247)",

      // Chart colors optimized for dark backgrounds
      chart1: "oklch(70% 0.15 247)",
      chart2: "oklch(75% 0.15 170)",
      chart3: "oklch(80% 0.15 85)",
      chart4: "oklch(70% 0.15 300)",
      chart5: "oklch(75% 0.15 15)",

      // Sidebar for dark mode
      sidebar: "oklch(10% 0.01 247)",
      sidebarForeground: "oklch(90% 0.01 247)",
      sidebarPrimary: "oklch(70% 0.15 247)",
      sidebarPrimaryForeground: "oklch(8% 0.01 247)",
      sidebarAccent: "oklch(18% 0.02 247)",
      sidebarAccentForeground: "oklch(85% 0.05 247)",
      sidebarBorder: "oklch(22% 0.01 247)",
      sidebarRing: "oklch(70% 0.15 247)",
    },
  },

  // Project management specific colors
  projectStatus: {
    todo: "oklch(45% 0.01 247)",
    inProgress: "oklch(55% 0.15 247)",
    review: "oklch(75% 0.15 85)",
    done: "oklch(70% 0.15 150)",
    blocked: "oklch(65% 0.2 15)",
  },

  priority: {
    low: "oklch(45% 0.01 247)",
    medium: "oklch(75% 0.15 85)",
    high: "oklch(65% 0.2 15)",
    critical: "oklch(65% 0.2 15)",
  },
} as const;

export type ThemeConfig = typeof themeConfig;
