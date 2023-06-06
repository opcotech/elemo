import type { Meta, StoryObj } from '@storybook/react';

import { TodoPriority } from '@/lib/api';
import { TodoListItem } from './TodoListItem';

const dueDate = new Date(Date.now() + 7 * 86400000).toISOString();

const meta: Meta<typeof TodoListItem> = {
  title: 'Todos/TodoListItem',
  component: TodoListItem,
  tags: ['autodocs'],
  render: (args) => (
    <ul>
      <TodoListItem {...args} />
    </ul>
  )
};

export default meta;
type Story = StoryObj<typeof TodoListItem>;

export const Sample: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const NoDescription: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: '',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const Completed: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: true,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const NormalPriority: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const ImportantPriority: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.IMPORTANT,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const UrgentPriority: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.URGENT,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const CriticalPriority: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.CRITICAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const NoDueDate: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: undefined,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const DueDateFarAway: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: new Date(Date.now() + 7 * 86400000).toISOString(),
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const DueDateClose: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: new Date(Date.now() + 2 * 86400000).toISOString(),
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const DueDatePast: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: new Date().toISOString(),
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: false
  }
};

export const Loading: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: true,
    editing: false,
    deleting: false
  }
};

export const Editing: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: true,
    deleting: false
  }
};

export const Deleting: Story = {
  args: {
    id: '1',
    title: 'Sample todo',
    description: 'This is a sample description',
    completed: false,
    priority: TodoPriority.NORMAL,
    created_by: '1',
    owned_by: '1',
    due_date: dueDate,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    loading: false,
    editing: false,
    deleting: true
  }
};
