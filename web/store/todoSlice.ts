import type { StateCreator } from 'zustand';
import client, { ContentType, CreateTodoData, GetTodosData, getErrorMessage } from '@/lib/api';
import type { Todo, GetTodosParams, UpdateTodoData } from '@/lib/api';
import type { MessageSliceState } from './messageSlice';

type OmittedTodoFields = 'id' | 'created_at' | 'updated_at';
export type CreateTodoParams = Omit<Todo, OmittedTodoFields | 'completed'>;
export type UpdateTodoParams = Omit<Todo, OmittedTodoFields | 'created_by' | 'owned_by'>;

export interface TodoSliceState {
  todos: Todo[];
  fetchingTodos: boolean;
  fetchedTodos: boolean;
  fetchTodos: (params?: GetTodosParams) => Promise<void>;
  createTodo: (todo: CreateTodoParams) => Promise<void>;
  updateTodo: (id: string, todo: UpdateTodoData) => Promise<void>;
  deleteTodo: (id: string) => Promise<void>;
}

export const createTodoSlice: StateCreator<TodoSliceState & Partial<MessageSliceState>> = (set, get) => ({
  todos: [],
  fetchingTodos: false,
  fetchedTodos: false,
  fetchTodos: async (params?: GetTodosParams): Promise<void> => {
    set({ fetchingTodos: true });
    const res = await client.v1.getTodos(params || {});
    const todos: GetTodosData = await res.json();

    if (!res.ok) {
      set({ fetchingTodos: false });
      return get().addMessage?.({
        type: 'error',
        title: 'Failed to fetch todos',
        message: res.error.message
      });
    }

    set({ todos });
    set({ fetchingTodos: false, fetchedTodos: true });
  },
  createTodo: async (todo: CreateTodoParams): Promise<void> => {
    try {
      const res = await client.v1.createTodo({ ...todo, completed: false }, { type: ContentType.Json });
      const data: CreateTodoData = await res.json();
      set((state) => ({ todos: [...state.todos, { ...todo, id: data.todo_id }] }));
      get().addMessage?.({
        type: 'success',
        title: 'Todo Created',
        message: `Todo "${data.todo_id}" created successfully`
      });
    } catch (e) {
      get().addMessage?.({
        type: 'error',
        title: 'Failed to create todo',
        message: getErrorMessage(e)
      });
    }
  },
  updateTodo: async (id: string, todo: UpdateTodoParams): Promise<void> => {
    try {
      const res = await client.v1.updateTodo(id, todo, { type: ContentType.Json });
      const updated: UpdateTodoData = await res.json();
      set((state) => ({ todos: state.todos.map((todo) => (todo.id === id ? updated : todo)) }));
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
      await client.v1.deleteTodo(id, { type: ContentType.Json });
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
