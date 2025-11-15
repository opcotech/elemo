import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import {
  NamespaceNavigation,
  ProjectNavigation,
  configBuilders,
} from "./sidebar-context";

import { useNavigationContext } from "@/hooks/use-navigation-context";
import { v1NamespaceGetOptions } from "@/lib/api";

export function ContextualNavigationSection() {
  const context = useNavigationContext();

  // Fetch namespace data if in namespace context
  const { data: namespace } = useQuery({
    ...v1NamespaceGetOptions({
      path: { id: context.namespaceId! },
    }),
    enabled: context.type === "namespace" && !!context.namespaceId,
  });

  // Don't render if not in a contextual view
  if (context.type === "global") {
    return null;
  }

  // Calculate navigation configuration based on context
  const navigationConfig = useMemo(() => {
    if (context.type === "global") return null;
    const builder = configBuilders[context.type];
    return builder ? builder(context, namespace) : null;
  }, [context, namespace]);

  if (!navigationConfig) {
    return null;
  }

  // For namespace context, show projects and documents
  if (context.type === "namespace" && namespace) {
    return (
      <NamespaceNavigation
        namespace={namespace}
        context={context}
        navigationItems={navigationConfig.items}
      />
    );
  }

  // For project context, show standard navigation
  if (context.type === "project") {
    return <ProjectNavigation navigationConfig={navigationConfig} />;
  }

  return null;
}
