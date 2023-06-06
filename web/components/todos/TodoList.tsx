import { memo } from 'react';
import { Todo } from '@/lib/api';
import { TodoListItem } from './TodoListItem';
import { UpdateTodoParams } from '@/store/todoSlice';

const MemoizedTodoListItem = memo(TodoListItem);

export interface TodoListProps {
  todos: Todo[];
  editing?: string;
  deleting: { id: string; timer: NodeJS.Timer | undefined }[];
  loading: string[];
  handleUpdate: (id: string, todo: UpdateTodoParams) => Promise<void>;
  handleEdit: (id: string, editing: boolean) => void;
  handleDelete: (id: string, deleting: boolean, timer?: NodeJS.Timer) => void;
}

export function TodoList({ todos, editing, deleting, loading, handleEdit, handleUpdate, handleDelete }: TodoListProps) {
  return (
    <ul>
      {todos.length === 0 && <li className="text-center text-gray-500">No todos found.</li>}

      {todos.map((todo) => (
        <MemoizedTodoListItem
          key={todo.id}
          {...todo}
          editing={editing === todo.id}
          deleting={deleting.filter((d) => d.id === todo.id).length > 0}
          loading={loading.includes(todo.id)}
          handleEditTodo={handleEdit}
          handleUpdateTodo={handleUpdate}
          handleDeleteTodo={handleDelete}
        />
      ))}
    </ul>
  );
}
