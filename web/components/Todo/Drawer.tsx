'use client';

import { memo, useCallback, useState } from 'react';
import Drawer from '@/components/Drawer';
import { ListSkeleton } from '@/components/Skeleton';
import useStore from '@/store';
import TodoItem from './TodoItem';
import { UpdateTodoParams } from '@/store/todoSlice';
import NewTodoForm from './NewTodoForm';

const MemoizedTodoItem = memo(TodoItem);

export interface TodoDrawerState {
  editing?: string;
  deleting: { id: string; timer: NodeJS.Timer | undefined }[];
  loading: string[];
}

export default function TodoDrawer() {
  const drawers = useStore((state) => state.drawers);
  const toggleDrawer = useStore((state) => state.toggleDrawer);

  const todos = useStore((state) => state.todos);
  const fetchingTodos = useStore((state) => state.fetchingTodos);
  const updateTodo = useStore((state) => state.updateTodo);

  const [state, setState] = useState<TodoDrawerState>({
    editing: undefined,
    deleting: [],
    loading: []
  });

  const setLoading = useCallback(
    (id: string, loading: boolean) =>
      setState((state) => ({
        ...state,
        loading: loading ? [...state.loading, id] : state.loading.filter((l) => l !== id)
      })),
    []
  );

  const handleUpdate = useCallback(
    async (id: string, todo: UpdateTodoParams) => {
      setLoading(id, true);
      await updateTodo(id, todo);
      setLoading(id, false);
    },
    [setLoading, updateTodo]
  );

  const handleEdit = useCallback((id: string, editing: boolean) => {
    setState((state) => ({
      ...state,
      editing: editing ? id : undefined
    }));
  }, []);

  const handleDelete = useCallback(
    (id: string, deleting: boolean, timer?: NodeJS.Timer) => {
      setState((state) => ({
        ...state,
        deleting: deleting
          ? [...state.deleting, { id, timer }]
          : state.deleting.filter((d) => {
              if (d.id === id && d.timer) clearTimeout(d.timer);
              return d.id !== id;
            })
      }));
      setLoading(id, deleting);
    },
    [setLoading]
  );

  return (
    <Drawer id="showTodos" title="Todos" show={drawers.showTodos} toggle={() => toggleDrawer('showTodos')}>
      <div className="mb-6">
        <NewTodoForm
          editing={todos.find((t) => t.id === state.editing)}
          onCancel={() => handleEdit(state.editing!, false)}
        />
      </div>

      {todos.length === 0 && fetchingTodos && <ListSkeleton count={5} />}

      <ul className="divide-y divide-gray-200">
        {todos.length === 0 && !fetchingTodos && <li className="text-center text-gray-500">No todos found.</li>}

        {todos.map((item) => (
          <MemoizedTodoItem
            key={item.id}
            {...item}
            editing={state.editing === item.id}
            deleting={state.deleting.filter((d) => d.id === item.id).length > 0}
            loading={state.loading.includes(item.id || '')}
            handleUpdateTodo={handleUpdate}
            handleEditTodo={handleEdit}
            handleDeleteTodo={handleDelete}
          />
        ))}
      </ul>
    </Drawer>
  );
}
