import { getErrorMessage, Todo, TodoPriority, TodosService } from '@/lib/api';
import { StateCreator } from 'zustand/esm';
import { MessageSliceState } from '@/store/messageSlice';
import { getSession } from 'next-auth/react';

const PRIORITY_MAP = {
  null: 4,
  [TodoPriority.NORMAL]: 4,
  [TodoPriority.IMPORTANT]: 3,
  [TodoPriority.URGENT]: 2,
  [TodoPriority.CRITICAL]: 1
};

export type FetchTodosParams = {
  offset?: number;
  limit?: number;
  completed?: boolean;
};

export type CreateTodoParams = {
  title: string;
  description?: string | null;
  priority: TodoPriority;
  owned_by: string;
  due_date?: string | null;
};

export type UpdateTodoParams = {
  title?: string;
  description?: string | null;
  priority?: TodoPriority;
  completed?: boolean;
  owned_by?: string;
  due_date?: string | null;
};

export interface TodoSliceState {
  todos: Todo[] | undefined;
  fetchingTodos: boolean;
  fetchTodos: (params?: FetchTodosParams) => Promise<void>;
  createTodo: (todo: CreateTodoParams) => Promise<void>;
  updateTodo: (id: string, todo: UpdateTodoParams) => Promise<void>;
  deleteTodo: (id: string) => Promise<void>;
}

export function sortTodos(items: Todo[]): Todo[] {
  return Object.assign([] as Todo[], items).sort((a, b) => {
    if (a.completed && !b.completed) {
      return 1;
    }

    if (!a.completed && b.completed) {
      return -1;
    }

    if (a.priority === b.priority && a.due_date && b.due_date) {
      return a.due_date > b.due_date ? 1 : -1;
    }

    return PRIORITY_MAP[a.priority] - PRIORITY_MAP[b.priority];
  });
}

export const createTodoSlice: StateCreator<TodoSliceState & Partial<MessageSliceState>> = (set, get) => ({
  todos: undefined,
  fetchingTodos: false,
  fetchTodos: async ({ offset = 0, limit = 100, completed }: FetchTodosParams = {}) => {
    let todos: Todo[] = [];

    try {
      set({ fetchingTodos: true });
      todos = await TodosService.v1TodosGet(offset, limit, completed);
    } catch (e) {
      return get().addMessage?.({ type: 'error', title: 'Failed to fetch todos', message: getErrorMessage(e) });
    } finally {
      set({ todos: sortTodos(todos), fetchingTodos: false });
    }
  },
  createTodo: async (todo: CreateTodoParams) => {
    let data: { id: string };

    const session = await getSession();
    todo = { ...todo, owned_by: session!.user!.id };

    try {
      data = await TodosService.v1TodosCreate(todo);
      get().addMessage?.({ type: 'success', title: 'Todo Created', message: `Todo "${data.id}" created successfully` });
    } catch (e) {
      return get().addMessage?.({ type: 'error', title: 'Failed to create todo', message: getErrorMessage(e) });
    }

    // NOTE: This is an ugly hack, the API should return the created todo
    // instead of just the ID.
    set((state) => ({
      todos: sortTodos([
        {
          ...todo,
          id: data.id,
          description: todo.description || '',
          due_date: todo.due_date || null,
          completed: false,
          created_by: todo.owned_by,
          created_at: new Date().toISOString(),
          updated_at: null
        },
        ...(state.todos || [])
      ])
    }));
  },
  updateTodo: async (id: string, todo: UpdateTodoParams) => {
    let updated: Todo;

    try {
      updated = await TodosService.v1TodoUpdate(id, todo);
      get().addMessage?.({ type: 'success', title: 'Todo updated', message: `Todo "${id}" updated successfully.` });
    } catch (e) {
      return get().addMessage?.({ type: 'error', title: 'Failed to update todo', message: getErrorMessage(e) });
    }

    set((state) => ({ todos: sortTodos((state.todos || []).map((todo) => (todo.id === id ? updated : todo))) }));
  },
  deleteTodo: async (id: string) => {
    try {
      await TodosService.v1TodoDelete(id);
      get().addMessage?.({ type: 'success', title: 'Todo deleted', message: `Todo "${id}" deleted successfully.` });
    } catch (e) {
      return get().addMessage?.({ type: 'error', title: 'Failed to delete todo', message: getErrorMessage(e) });
    }

    set((state) => ({ todos: sortTodos((state.todos || []).filter((todo) => todo.id !== id)) }));
  }
});
