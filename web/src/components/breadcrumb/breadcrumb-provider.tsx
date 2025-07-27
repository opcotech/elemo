"use client";

import type { ReactNode } from "react";
import { createContext, useCallback, useContext, useState } from "react";

export interface BreadcrumbItem {
  label: string;
  href?: string;
  isNavigatable?: boolean;
}

interface BreadcrumbContextType {
  breadcrumbs: BreadcrumbItem[];
  setBreadcrumbs: (
    breadcrumbs:
      | BreadcrumbItem[]
      | ((prev: BreadcrumbItem[]) => BreadcrumbItem[])
  ) => void;
  clearBreadcrumbs: () => void;
}

const BreadcrumbContext = createContext<BreadcrumbContextType | undefined>(
  undefined
);

interface BreadcrumbProviderProps {
  children: ReactNode;
  initialBreadcrumbs?: BreadcrumbItem[];
}

export function BreadcrumbProvider({
  children,
  initialBreadcrumbs = [],
}: BreadcrumbProviderProps) {
  const [breadcrumbs, setBreadcrumbs] =
    useState<BreadcrumbItem[]>(initialBreadcrumbs);

  const clearBreadcrumbs = useCallback(() => {
    setBreadcrumbs([]);
  }, []);

  const setBreadcrumbsStable = useCallback(
    (
      items: BreadcrumbItem[] | ((prev: BreadcrumbItem[]) => BreadcrumbItem[])
    ) => {
      if (typeof items === "function") {
        setBreadcrumbs(items);
      } else {
        setBreadcrumbs(items);
      }
    },
    []
  );

  const value = {
    breadcrumbs,
    setBreadcrumbs: setBreadcrumbsStable,
    clearBreadcrumbs,
  };

  return (
    <BreadcrumbContext.Provider value={value}>
      {children}
    </BreadcrumbContext.Provider>
  );
}

export function useBreadcrumbs() {
  const context = useContext(BreadcrumbContext);

  if (context === undefined) {
    throw new Error("useBreadcrumbs must be used within a BreadcrumbProvider");
  }

  return context;
}
