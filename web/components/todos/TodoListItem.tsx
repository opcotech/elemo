import { useCallback } from 'react';
import { Todo, TodoPriority } from '@/lib/api';
import { concat } from '@/lib/helpers';
import { Icon } from '@/components/blocks/Icon';
import { Button } from '@/components/blocks/Button';
import useStore from '@/store';

const PRIORITY_CLASSES: Record<TodoPriority, string> = {
  normal: '',
  important: 'text-blue-600',
  urgent: 'text-yellow-600',
  critical: 'text-red-600'
};

export interface TodoListItemProps extends Todo {
  loading: boolean;
  editing: boolean;
  deleting: boolean;
  handleUpdateTodo: (id: string, todo: any) => void;
  handleEditTodo: (id: string, editing: boolean) => void;
  handleDeleteTodo: (id: string, deleting: boolean, timer?: NodeJS.Timer) => void;
}

export function TodoListItem({
  id,
  title,
  description,
  priority,
  completed,
  due_date,
  loading,
  editing,
  deleting,
  handleEditTodo,
  handleUpdateTodo,
  handleDeleteTodo
}: TodoListItemProps) {
  const dueDateClass = () => {
    const today = new Date();
    const dueDate = new Date(due_date!);

    if (dueDate < today) {
      return 'text-red-600';
    } else if (dueDate.getDate() - today.getDate() <= 3) {
      return 'text-yellow-600';
    }

    return '';
  };

  const deleteTodo = useStore((state) => state.deleteTodo);

  const priorityClass = PRIORITY_CLASSES[priority || 'normal'];
  const interactive = !editing && !deleting && !loading;

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
            className="rounded-full disabled:bg-gray-100 disabled:border-gray-300  disabled:hover:text-gray-100 disabled:hover:border-gray-300"
          />
        </div>

        <div className="min-w-0 flex-1">
          <p className="text-gray-900">
            {priorityClass && (
              <Icon
                size="sm"
                variant="FlagIcon"
                className={`inline font-bold pr-1 priority-${priority} ${priorityClass}`}
              />
            )}
            <span className={completed ? 'line-through' : ''}>{title}</span>
          </p>
          {description && (
            <p className={concat('text-sm text-gray-500', completed ? 'line-through' : '')}>{description}</p>
          )}
          {due_date && (
            <p className={concat('text-xs pt-2 text-gray-500', completed ? 'line-through' : '', dueDateClass())}>
              {new Date(due_date).toLocaleDateString()}
            </p>
          )}
        </div>
        <div className="space-x-2">
          {!completed && interactive && (
            <Button size="xs" icon="PencilSquareIcon" onClick={handleEdit} disabled={!interactive}>
              <span className="sr-only">Edit item</span>
            </Button>
          )}

          {deleting && (
            <Button size="xs" icon="ArrowUturnLeftIcon" onClick={handleDeleteCancel}>
              <span className="sr-only">Cancel deletion</span>
            </Button>
          )}
          {!deleting && !editing && (
            <Button
              size="xs"
              icon="TrashIcon"
              onClick={handleDelete}
              disabled={!interactive}
              className="text-red-600 hover:text-red-700"
            >
              <span className="sr-only">Delete item</span>
            </Button>
          )}
        </div>
      </div>
    </li>
  );
}
