import type { Meta, StoryObj } from '@storybook/react';

import { TodoForm } from './TodoForm';
import { TodoPriority } from '@/lib/api';

const meta: Meta<typeof TodoForm> = {
  title: 'Todos/TodoForm',
  component: TodoForm,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof TodoForm>;

export const Sample: Story = {
  args: {
    editing: undefined,
    onHide: () => {},
    onCancel: () => {}
  }
};

export const Editing: Story = {
  args: {
    editing: {
      id: '1',
      title: 'Todo item',
      description: 'This item is important for me',
      completed: false,
      priority: TodoPriority.IMPORTANT,
      created_by: '1',
      owned_by: '1',
      due_date: new Date().toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    },
    onHide: () => {},
    onCancel: () => {}
  }
};
