import type { ReactNode } from "react";

import { ListContainer } from "@/components/ui/list-container";
import { SearchInput } from "@/components/ui/search-input";

export function SettingsResourceTable({
  title,
  description,
  dataSection,
  isLoading,
  error,
  actionButton,
  search,
  empty,
  skeleton,
  children,
}: {
  title: string;
  description: string;
  dataSection?: string;
  isLoading: boolean;
  error: unknown;
  actionButton?: ReactNode;
  search: {
    value: string;
    onChange: (value: string) => void;
    placeholder: string;
    itemCount: number;
  };
  empty: {
    icon: ReactNode;
    title: string;
    description: string;
    action?: ReactNode;
    searchTitle: string;
    searchDescription: string;
    hasItems: boolean;
    hasFilteredItems: boolean;
  };
  skeleton: ReactNode;
  children: ReactNode;
}) {
  const searchActive = search.value.trim() !== "";
  const emptyState = !empty.hasItems
    ? {
        icon: empty.icon,
        title: empty.title,
        description: empty.description,
        action: empty.action,
      }
    : !empty.hasFilteredItems && searchActive
      ? {
          icon: empty.icon,
          title: empty.searchTitle,
          description: empty.searchDescription,
        }
      : undefined;

  return (
    <ListContainer
      data-section={dataSection}
      title={title}
      description={description}
      isLoading={isLoading}
      error={error}
      emptyState={emptyState}
      actionButton={actionButton}
      searchInput={
        search.itemCount > 0 || searchActive ? (
          <SearchInput
            value={search.value}
            onChange={search.onChange}
            placeholder={search.placeholder}
            disabled={isLoading}
          />
        ) : undefined
      }
    >
      {isLoading ? skeleton : children}
    </ListContainer>
  );
}
