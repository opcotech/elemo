import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { internalPath } from "./internal-url";
import { uiActions, uiStore, useUiSelector } from "./ui-store";

import type { Todo } from "@/lib/api/types";

const todo: Todo = {
  id: "todo-1",
  title: "Review modernization",
  description: "Check state ownership",
  priority: "normal",
  completed: false,
  owned_by: "user-1",
  created_by: "user-1",
  due_date: null,
  created_at: "2026-08-10T00:00:00.000Z",
  updated_at: null,
};

function createStorage() {
  const values = new Map<string, string>();
  return {
    clear: () => values.clear(),
    getItem: (key: string) => values.get(key) ?? null,
    key: (index: number) => [...values.keys()][index] ?? null,
    get length() {
      return values.size;
    },
    removeItem: (key: string) => values.delete(key),
    setItem: (key: string, value: string) => values.set(key, value),
  } satisfies Storage;
}

describe("uiStore actions", () => {
  let localStorage: Storage;

  beforeEach(() => {
    localStorage = createStorage();
    vi.stubGlobal("window", { localStorage });
    uiStore.setState(() => ({
      commandPaletteOpen: false,
      todoSheetOpen: false,
      addTodoOpen: false,
      editingTodo: null,
      recentEntities: [],
    }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("owns the todo workflow atomically", () => {
    uiActions.openEditTodo(todo);
    expect(uiStore.state).toMatchObject({
      todoSheetOpen: true,
      addTodoOpen: false,
      editingTodo: todo,
    });

    uiActions.openAddTodo();
    expect(uiStore.state).toMatchObject({
      todoSheetOpen: true,
      addTodoOpen: true,
      editingTodo: null,
    });

    uiActions.closeTodoSheet();
    expect(uiStore.state).toMatchObject({
      todoSheetOpen: false,
      addTodoOpen: false,
      editingTodo: null,
    });
  });

  it("updates command palette state without touching todo state", () => {
    uiActions.openTodoSheet();
    uiActions.setCommandPaletteOpen(true);

    expect(uiStore.state.commandPaletteOpen).toBe(true);
    expect(uiStore.state.todoSheetOpen).toBe(true);
  });

  it("deduplicates, bounds, and persists recent navigation context", () => {
    for (let index = 0; index < 10; index += 1) {
      uiActions.rememberRecentEntity({
        id: `namespace-${index}`,
        type: "namespace",
        label: `Namespace ${index}`,
        href: internalPath(`/namespaces/namespace-${index}`),
        namespaceId: `namespace-${index}`,
      });
    }
    uiActions.rememberRecentEntity({
      id: "namespace-5",
      type: "namespace",
      label: "Namespace 5 revisited",
      href: "/namespaces/namespace-5",
      namespaceId: "namespace-5",
    });

    expect(uiStore.state.recentEntities).toHaveLength(10);
    expect(uiStore.state.recentEntities[0]).toMatchObject({
      id: "namespace-5",
      label: "Namespace 5 revisited",
    });
    expect(
      uiStore.state.recentEntities.filter(
        (entity) => entity.id === "namespace-5"
      )
    ).toHaveLength(1);
    const persisted = JSON.parse(
      localStorage.getItem("elemo_navigation_context") ?? "{}"
    );
    expect(persisted.recentEntities).toHaveLength(10);
  });

  it("keeps only the seven most recent work items", () => {
    for (let index = 0; index < 9; index += 1) {
      uiActions.rememberRecentEntity({
        id: `work-${index}`,
        type: "work-item",
        label: `Work ${index}`,
        href: internalPath(`/work/work-${index}`),
        namespaceId: "namespace-product",
      });
    }

    expect(
      uiStore.state.recentEntities
        .filter((entity) => entity.type === "work-item")
        .map((entity) => entity.id)
    ).toEqual([
      "work-8",
      "work-7",
      "work-6",
      "work-5",
      "work-4",
      "work-3",
      "work-2",
    ]);
  });

  it("keeps only the five most recent documents", () => {
    for (let index = 0; index < 7; index += 1) {
      uiActions.rememberRecentEntity({
        id: `document-${index}`,
        type: "document",
        label: `Document ${index}`,
        href: internalPath(`/documents/document-${index}`),
        namespaceId: "namespace-product",
      });
    }

    expect(
      uiStore.state.recentEntities
        .filter((entity) => entity.type === "document")
        .map((entity) => entity.id)
    ).toEqual([
      "document-6",
      "document-5",
      "document-4",
      "document-3",
      "document-2",
    ]);
  });

  it("keeps only the three most recent projects", () => {
    for (let index = 0; index < 5; index += 1) {
      uiActions.rememberRecentEntity({
        id: `project-${index}`,
        type: "project",
        label: `Project ${index}`,
        href: internalPath(`/namespaces/ns/projects/project-${index}`),
        namespaceId: "ns",
      });
    }

    expect(
      uiStore.state.recentEntities
        .filter((entity) => entity.type === "project")
        .map((entity) => entity.id)
    ).toEqual(["project-4", "project-3", "project-2"]);
  });

  it("respects the global recent entity bound across types", () => {
    for (let index = 0; index < 25; index += 1) {
      uiActions.rememberRecentEntity({
        id: `namespace-${index}`,
        type: "namespace",
        label: `Namespace ${index}`,
        href: internalPath(`/namespaces/namespace-${index}`),
        namespaceId: `namespace-${index}`,
      });
    }

    expect(uiStore.state.recentEntities).toHaveLength(20);
    expect(uiStore.state.recentEntities[0]?.id).toBe("namespace-24");
    expect(uiStore.state.recentEntities.at(-1)?.id).toBe("namespace-5");
  });

  it("forgets a recent entity and persists the result", () => {
    uiActions.rememberRecentEntity({
      id: "project-1",
      type: "project",
      label: "Alpha",
      href: internalPath("/namespaces/ns/projects/project-1"),
      namespaceId: "ns",
    });
    uiActions.rememberRecentEntity({
      id: "work-1",
      type: "work-item",
      label: "Work One",
      href: internalPath("/work/work-1"),
      namespaceId: "ns",
    });
    uiActions.rememberRecentEntity({
      id: "document-1",
      type: "document",
      label: "Doc One",
      href: internalPath("/documents/document-1"),
      namespaceId: "ns",
    });

    uiActions.forgetRecentEntity({ id: "work-1", type: "work-item" });

    expect(uiStore.state.recentEntities.map((entity) => entity.id)).toEqual([
      "document-1",
      "project-1",
    ]);
    const persisted = JSON.parse(
      localStorage.getItem("elemo_navigation_context") ?? "{}"
    );
    expect(
      persisted.recentEntities.map((entity: { id: string }) => entity.id)
    ).toEqual(["document-1", "project-1"]);
  });

  it("preserves persisted recents when remembering before hydrate", () => {
    localStorage.setItem(
      "elemo_navigation_context",
      JSON.stringify({
        recentEntities: [
          {
            id: "lmo-101",
            type: "work-item",
            label: "LMO-101 Work",
            href: "/work/lmo-101",
            visitedAt: 10,
          },
        ],
      })
    );

    uiActions.rememberRecentEntity({
      id: "document-1",
      type: "document",
      label: "Doc",
      href: internalPath("/documents/document-1"),
      namespaceId: "ns",
    });

    expect(
      uiStore.state.recentEntities.map(
        (entity) => `${entity.type}:${entity.id}`
      )
    ).toEqual(["document:document-1", "work-item:lmo-101"]);
  });

  it("hydrates valid navigation fields and ignores unsafe entities", () => {
    localStorage.setItem(
      "elemo_navigation_context",
      JSON.stringify({
        inspectorWidth: 900,
        recentEntities: [
          {
            id: "document-1",
            type: "document",
            label: "Incident review",
            href: "/documents/document-1",
            visitedAt: 10,
          },
          {
            id: "unsafe",
            type: "document",
            label: "Unsafe",
            href: "https://example.com/phishing",
            visitedAt: 20,
          },
          {
            id: "legacy-relation",
            type: "relation",
            label: "Legacy relation",
            href: "/relations/work-item/lmo-1",
            visitedAt: 30,
          },
          { id: "invalid" },
        ],
      })
    );

    uiActions.hydrateNavigation();

    expect(uiStore.state.recentEntities.map((entity) => entity.id)).toEqual([
      "document-1",
    ]);
    expect(uiStore.state).not.toHaveProperty("inspectorWidth");
  });

  it("rejects unsafe recent entity URLs before persistence", () => {
    uiActions.rememberRecentEntity({
      id: "unsafe",
      type: "document",
      label: "Unsafe",
      href: "https://example.com/phishing" as never,
    });

    expect(uiStore.state.recentEntities).toEqual([]);
    expect(localStorage.getItem("elemo_navigation_context")).toBeNull();
  });

  it("projects only the requested selector value", () => {
    const PaletteState = () =>
      createElement(
        "span",
        null,
        useUiSelector((state) => state.commandPaletteOpen) ? "open" : "closed"
      );

    expect(renderToStaticMarkup(createElement(PaletteState))).toContain(
      "closed"
    );

    uiActions.openTodoSheet();
    expect(renderToStaticMarkup(createElement(PaletteState))).toContain(
      "closed"
    );

    uiActions.setCommandPaletteOpen(true);
    expect(renderToStaticMarkup(createElement(PaletteState))).toContain("open");
  });
});
