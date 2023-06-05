import type { Meta, StoryObj } from '@storybook/react';

import { TodoPriority } from '@/lib/api';
import { TodoList } from './TodoList';

const meta: Meta<typeof TodoList> = {
  title: 'Todos/TodoList',
  component: TodoList,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof TodoList>;

export const Sample: Story = {
  args: {
    todos: [
      {
        id: '1',
        title: 'Sample todo',
        description: 'This is a sample description',
        completed: false,
        priority: TodoPriority.NORMAL,
        due_date: new Date().toISOString(),
        created_by: '1',
        owned_by: '1',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: '2',
        title: 'Completed todo',
        description: 'This is a completed todo description',
        completed: true,
        priority: TodoPriority.NORMAL,
        due_date: new Date().toISOString(),
        created_by: '1',
        owned_by: '1',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: '3',
        title: 'Critical prio todo',
        description: 'This is a critical todo description',
        completed: false,
        priority: TodoPriority.CRITICAL,
        due_date: new Date().toISOString(),
        created_by: '1',
        owned_by: '1',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: '4',
        title: 'Urgent prio todo',
        description: 'This is a urgent todo description',
        completed: false,
        priority: TodoPriority.URGENT,
        due_date: new Date().toISOString(),
        created_by: '1',
        owned_by: '1',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: '5',
        title: 'Important prio todo',
        description: 'This is a important todo description',
        completed: false,
        priority: TodoPriority.IMPORTANT,
        due_date: new Date().toISOString(),
        created_by: '1',
        owned_by: '1',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    ]
  }
};

export const WithNoItems: Story = {
  args: {
    todos: []
  }
};
