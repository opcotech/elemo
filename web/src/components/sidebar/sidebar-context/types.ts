import type { useNavigationContext } from "@/hooks/use-navigation-context";

export interface NavigationItemConfig {
  label: string;
  pathSuffix?: string;
  icon: React.ElementType;
}

export interface NavigationItem extends NavigationItemConfig {
  href: string;
}

export interface NavigationConfig {
  label: string;
  items: NavigationItem[];
}

export type NavigationContext = ReturnType<typeof useNavigationContext>;

export type NavigationConfigBuilder = (
  context: NavigationContext,
  namespace?: { name: string } | null
) => NavigationConfig | null;
