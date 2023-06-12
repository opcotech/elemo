import type { StateCreator } from 'zustand';

type Message = {
  id: number;
  title: string;
  message?: string;
  type: 'info' | 'success' | 'warning' | 'error';
  dismissAfter?: number;
};

export interface MessageSliceState {
  messages: Message[];
  addMessage: (message: Omit<Message, 'id'>) => void;
  removeMessage: (id: number) => void;
}

export const createMessageSlice: StateCreator<MessageSliceState> = (set) => ({
  messages: [],
  addMessage: (message) => set((state) => ({ messages: [...state.messages, { ...message, id: Date.now() }] })),
  removeMessage: (id: number) => set((state) => ({ messages: state.messages.filter((m) => m.id !== id) }))
});
