import type { StateCreator } from 'zustand';
import client, { ContentType, TodoPriority, getErrorMessage, V1TodosGetParams, V1TodoUpdateData } from '@/lib/api';
import type { Todo } from '@/lib/api';
import type { MessageSliceState } from './messageSlice';

const PRIORITY_MAP = {
  null: 4,
  [TodoPriority.Normal]: 4,
  [TodoPriority.Important]: 3,
  [TodoPriority.Urgent]: 2,
  [TodoPriority.Critical]: 1
};

export type CreateTodoInput = {
  title: string;
  description?: string | undefined;
  priority: TodoPriority;
  owned_by: string;
  due_date?: string | undefined;
};

export type UpdateTodoInput = {
  title?: string;
  description?: string;
  priority?: TodoPriority;
  completed?: boolean;
  owned_by?: string;
  due_date?: string | null;
};

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

export interface TodoSliceState {
  todos: Todo[];
  fetchingTodos: boolean;
  fetchedTodos: boolean;
  fetchTodos: (params?: V1TodosGetParams) => Promise<void>;
  createTodo: (todo: CreateTodoInput) => Promise<void>;
  updateTodo: (id: string, todo: UpdateTodoInput) => Promise<void>;
  deleteTodo: (id: string) => Promise<void>;
}

export const createTodoSlice: StateCreator<TodoSliceState & Partial<MessageSliceState>> = (set, get) => ({
  todos: [],
  fetchingTodos: false,
  fetchedTodos: false,
  fetchTodos: async (params: V1TodosGetParams = {}): Promise<void> => {
    set({ fetchingTodos: true });
    try {
      const res = await client.v1.v1TodosGet(params);
      const todos: Todo[] = await res.json();
      set({ todos: sortTodos(todos) });
      set({ fetchingTodos: false, fetchedTodos: true });
    } catch (e) {
      set({ fetchingTodos: false });
      get().addMessage?.({
        type: 'error',
        title: 'Failed to fetch todos',
        message: getErrorMessage(e)
      });
    }
  },
  createTodo: async (todo: CreateTodoInput): Promise<void> => {
    try {
      const res = await client.v1.v1TodosCreate(
        {
          title: todo.title,
          description: todo.description === '' ? undefined : todo.description,
          priority: todo.priority,
          owned_by: todo.owned_by,
          due_date: todo.due_date || ''
        },
        { type: ContentType.Json }
      );

      const data: { id: string } = await res.json();

      // NOTE: This is an ugly hack, the API should return the created todo
      // instead of just the ID.
      set((state) => ({
        todos: sortTodos([
          {
            id: data.id,
            title: todo.title,
            description: todo.description || '',
            priority: todo.priority,
            owned_by: todo.owned_by,
            due_date: todo.due_date || null,
            completed: false,
            created_by: todo.owned_by,
            created_at: new Date().toISOString(),
            updated_at: null
          },
          ...state.todos
        ])
      }));

      get().addMessage?.({
        type: 'success',
        title: 'Todo Created',
        message: `Todo "${data.id}" created successfully`
      });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to create todo',
        message: getErrorMessage(e)
      });
    }
  },
  updateTodo: async (id: string, todo: UpdateTodoInput): Promise<void> => {
    try {
      const res = await client.v1.v1TodoUpdate(
        id,
        {
          title: todo.title,
          description: todo.description,
          priority: todo.priority,
          completed: todo.completed,
          owned_by: todo.owned_by,
          due_date: todo.due_date === null ? '' : todo.due_date
        },
        { type: ContentType.Json }
      );
      const updated: V1TodoUpdateData = await res.json();
      set((state) => ({ todos: sortTodos(state.todos.map((todo) => (todo.id === id ? updated : todo))) }));
      get().addMessage?.({
        type: 'success',
        title: 'Todo updated',
        message: `Todo "${id}" updated successfully.`
      });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to update todo',
        message: getErrorMessage(e)
      });
    }
  },
  deleteTodo: async (id: string): Promise<void> => {
    try {
      await client.v1.v1TodoDelete(id);
      set((state) => ({ todos: state.todos.filter((todo) => todo.id !== id) }));
      get().addMessage?.({
        type: 'success',
        title: 'Todo deleted',
        message: `Todo "${id}" deleted successfully.`
      });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to delete todo',
        message: getErrorMessage(e)
      });
    }
  }
});
