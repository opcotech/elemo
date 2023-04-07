import type { StateCreator } from 'zustand';

export interface MessageSliceState {
  messages: Message[];
  addMessage: (message: Omit<Message, 'id'>) => void;
  removeMessage: (id: number) => void;
}

export const messageSlice: StateCreator<MessageSliceState> = (set) => ({
  messages: [],
  addMessage: (message) => set((state) => ({ messages: [...state.messages, { ...message, id: Date.now() }] })),
  removeMessage: (id: number) => set((state) => ({ messages: state.messages.filter((m) => m.id !== id) }))
});
