'use client';

import { useCallback } from 'react';
import IconButton from '@/components/Button/IconButton';
import Icon from '@/components/Icon';
import { concat } from '@/helpers';
import { TodoPriority, Todo } from '@/lib/api';
import { UpdateTodoParams } from '@/store/todoSlice';
import useStore from '@/store';

const PRIORITY_CLASSES: Record<TodoPriority, string> = {
  normal: '',
  important: 'text-blue-600',
  urgent: 'text-yellow-600',
  critical: 'text-red-600'
};

export interface TodoItemProps extends Todo {
  loading: boolean;
  editing: boolean;
  deleting: boolean;
  handleUpdateTodo: (id: string, todo: UpdateTodoParams) => void;
  handleEditTodo: (id: string, editing: boolean) => void;
  handleDeleteTodo: (id: string, deleting: boolean, timer?: NodeJS.Timer) => void;
}

export default function TodoItem({
  id,
  title,
  description,
  priority,
  completed,
  loading,
  editing,
  deleting,
  handleEditTodo,
  handleUpdateTodo,
  handleDeleteTodo
}: TodoItemProps) {
  const priorityClass = PRIORITY_CLASSES[priority || 'normal'];
  const interactive = !editing && !deleting && !loading;
  const deleteTodo = useStore((state) => state.deleteTodo);

  const handleComplete = useCallback(() => {
    handleUpdateTodo(id!, { completed: !completed });
  }, [id, completed, handleUpdateTodo]);

  const handleEdit = useCallback(() => {
    handleEditTodo(id!, true);
  }, [id, handleEditTodo]);

  const handleDelete = useCallback(() => {
    const handler = setTimeout(async () => {
      await deleteTodo(id!);
      clearInterval(handler);
      handleDeleteTodo(id!, false);
    }, 5000);

    handleDeleteTodo(id!, true, handler);
  }, [id, deleteTodo, handleDeleteTodo]);

  async function handleDeleteCancel() {
    handleDeleteTodo(id!, false);
  }

  return (
    <li className={'py-4'}>
      <div className="flex items-start space-x-3">
        <div className="flex-shrink-0 items-start">
          <input
            type="checkbox"
            disabled={!interactive}
            checked={completed}
            onChange={handleComplete}
            className="rounded-full"
          />
        </div>

        <div className="min-w-0 flex-1">
          <p className="text-gray-900">
            {priorityClass && (
              <Icon
                variant="FlagIcon"
                className={`h-5 w-5 inline font-bold pr-1 priority-${priority} ${priorityClass}`}
              />
            )}
            <span className={completed ? 'line-through' : ''}>{title}</span>
          </p>
          {description && (
            <p className={concat('text-sm text-gray-500', completed ? 'line-through' : '')}>{description}</p>
          )}
        </div>
        <div className="space-x-2">
          {!completed && interactive && (
            <IconButton
              size={4}
              icon={'PencilSquareIcon'}
              onClick={handleEdit}
              disabled={!interactive}
              className="text-gray-600 hover:text-black disabled:opacity-70"
            >
              <span className="sr-only">Edit item</span>
            </IconButton>
          )}

          {deleting && (
            <IconButton
              size={4}
              icon={'ArrowUturnLeftIcon'}
              onClick={handleDeleteCancel}
              className="text-gray-600 hover:text-black disabled:opacity-70"
            >
              <span className="sr-only">Cancel deletion</span>
            </IconButton>
          )}
          {!deleting && !editing && (
            <IconButton
              size={4}
              icon={'TrashIcon'}
              onClick={handleDelete}
              disabled={!interactive}
              className="text-red-600 hover:text-red-700 disabled:opacity-70"
            >
              <span className="sr-only">Delete item</span>
            </IconButton>
          )}
        </div>
      </div>
    </li>
  );
}
