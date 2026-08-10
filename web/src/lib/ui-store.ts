import { Store, useSelector } from "@tanstack/react-store";

import type { Todo } from "@/lib/api/types";
import { isSafeInternalPath } from "@/lib/internal-url";
import type { InternalPath } from "@/lib/internal-url";

const NAVIGATION_STORAGE_KEY = "elemo_navigation_context";
const MAX_RECENT_ENTITIES = 20;

export type RecentEntityType = "namespace" | "project" | "work" | "document";

const MAX_RECENT_BY_TYPE: Partial<Record<RecentEntityType, number>> = {
  work: 7,
  document: 5,
  project: 3,
};

export interface RecentEntity {
  id: string;
  type: RecentEntityType;
  label: string;
  href: InternalPath;
  namespaceId?: string;
  visitedAt: number;
}

interface PersistedNavigationState {
  recentEntities: RecentEntity[];
}

export interface UiState {
  commandPaletteOpen: boolean;
  todoSheetOpen: boolean;
  addTodoOpen: boolean;
  editingTodo: Todo | null;
  recentEntities: RecentEntity[];
}

const initialState: UiState = {
  commandPaletteOpen: false,
  todoSheetOpen: false,
  addTodoOpen: false,
  editingTodo: null,
  recentEntities: [],
};

export const uiStore = new Store<UiState>(initialState);

function updateUiState(update: (state: UiState) => UiState) {
  uiStore.setState(update);
}

function isRecentEntity(value: unknown): value is RecentEntity {
  if (!value || typeof value !== "object") {
    return false;
  }

  const entity = value as Partial<RecentEntity>;
  return (
    typeof entity.id === "string" &&
    typeof entity.type === "string" &&
    ["namespace", "project", "work", "document"].includes(entity.type) &&
    typeof entity.label === "string" &&
    isSafeInternalPath(entity.href) &&
    typeof entity.visitedAt === "number"
  );
}

function persistNavigation(state: UiState) {
  if (typeof window === "undefined") {
    return;
  }

  const persistedState: PersistedNavigationState = {
    recentEntities: state.recentEntities,
  };

  try {
    window.localStorage.setItem(
      NAVIGATION_STORAGE_KEY,
      JSON.stringify(persistedState)
    );
  } catch (error) {
    console.warn("Failed to persist navigation context", error);
  }
}

function readPersistedRecentEntities(): RecentEntity[] {
  if (typeof window === "undefined") {
    return [];
  }

  try {
    const stored = JSON.parse(
      window.localStorage.getItem(NAVIGATION_STORAGE_KEY) || "{}"
    ) as Partial<PersistedNavigationState>;
    return Array.isArray(stored.recentEntities)
      ? stored.recentEntities.filter(isRecentEntity)
      : [];
  } catch {
    return [];
  }
}

function mergeRecentEntities(...groups: RecentEntity[][]): RecentEntity[] {
  const merged = new Map<string, RecentEntity>();

  for (const group of groups) {
    for (const entity of group) {
      const key = `${entity.type}:${entity.id}`;
      const previous = merged.get(key);
      if (!previous || entity.visitedAt >= previous.visitedAt) {
        merged.set(key, entity);
      }
    }
  }

  return [...merged.values()].sort(
    (left, right) => right.visitedAt - left.visitedAt
  );
}

function boundRecentEntities(entities: RecentEntity[]): RecentEntity[] {
  const typeCounts: Partial<Record<RecentEntityType, number>> = {};
  const bounded: RecentEntity[] = [];

  for (const entity of entities) {
    const maxForType = MAX_RECENT_BY_TYPE[entity.type];
    if (maxForType !== undefined) {
      const currentCount = typeCounts[entity.type] ?? 0;
      if (currentCount >= maxForType) {
        continue;
      }
      typeCounts[entity.type] = currentCount + 1;
    }

    bounded.push(entity);
    if (bounded.length >= MAX_RECENT_ENTITIES) {
      break;
    }
  }

  return bounded;
}

export const uiActions = {
  setCommandPaletteOpen(open: boolean) {
    updateUiState((state) => ({ ...state, commandPaletteOpen: open }));
  },
  openTodoSheet() {
    updateUiState((state) => ({ ...state, todoSheetOpen: true }));
  },
  closeTodoSheet() {
    updateUiState((state) => ({
      ...state,
      todoSheetOpen: false,
      addTodoOpen: false,
      editingTodo: null,
    }));
  },
  openAddTodo() {
    updateUiState((state) => ({
      ...state,
      todoSheetOpen: true,
      addTodoOpen: true,
      editingTodo: null,
    }));
  },
  closeAddTodo() {
    updateUiState((state) => ({ ...state, addTodoOpen: false }));
  },
  openEditTodo(todo: Todo) {
    updateUiState((state) => ({
      ...state,
      todoSheetOpen: true,
      addTodoOpen: false,
      editingTodo: todo,
    }));
  },
  closeEditTodo() {
    updateUiState((state) => ({ ...state, editingTodo: null }));
  },
  hydrateNavigation() {
    if (typeof window === "undefined") {
      return;
    }

    try {
      const persisted = readPersistedRecentEntities();
      updateUiState((state) => ({
        ...state,
        recentEntities: boundRecentEntities(
          mergeRecentEntities(persisted, state.recentEntities)
        ),
      }));
    } catch {
      window.localStorage.removeItem(NAVIGATION_STORAGE_KEY);
    }
  },
  rememberRecentEntity(entity: Omit<RecentEntity, "visitedAt">) {
    if (!isSafeInternalPath(entity.href)) {
      return;
    }

    updateUiState((state) => {
      const recentEntity: RecentEntity = {
        ...entity,
        visitedAt: Date.now(),
      };
      const existing = mergeRecentEntities(
        readPersistedRecentEntities(),
        state.recentEntities
      ).filter((item) => !(item.id === entity.id && item.type === entity.type));
      const recentEntities = boundRecentEntities([recentEntity, ...existing]);
      const nextState = {
        ...state,
        recentEntities,
      };
      persistNavigation(nextState);
      return nextState;
    });
  },
  forgetRecentEntity(entity: Pick<RecentEntity, "id" | "type">) {
    updateUiState((state) => {
      const recentEntities = state.recentEntities.filter(
        (item) => !(item.id === entity.id && item.type === entity.type)
      );
      if (recentEntities.length === state.recentEntities.length) {
        return state;
      }

      const nextState = {
        ...state,
        recentEntities,
      };
      persistNavigation(nextState);
      return nextState;
    });
  },
};

export function useUiSelector<T>(selector: (state: UiState) => T): T {
  return useSelector(uiStore, selector);
}
