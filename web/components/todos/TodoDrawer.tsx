'use client';

import { useCallback, useEffect, useState } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import useStore from '@/store';
import { Drawer } from '@/components/blocks/Drawer';
import { Button } from '@/components/blocks/Button';
import { UpdateTodoParams } from '@/store/todoSlice';
import { ListSkeleton } from '@/components/blocks/Skeleton/ListSkeleton';
import { TodoForm } from './TodoForm';
import { TodoList } from './TodoList';

type TodoListState = {
  editing?: string;
  deleting: { id: string; timer: NodeJS.Timer | undefined }[];
  loading: string[];
};

export function TodoDrawer() {
  const [showNewTodoForm, setShowNewTodoForm] = useState(false);

  const [todos, show, toggleDrawer, fetchingTodos, fetchTodos, updateTodo] = useStore((state) => [
    state.todos,
    state.showing.todos,
    () => state.toggleDrawer('todos'),
    state.fetchingTodos,
    state.fetchTodos,
    state.updateTodo
  ]);

  const [state, setState] = useState<TodoListState>({
    editing: undefined,
    deleting: [],
    loading: []
  });

  useEffect(() => {
    if (show && !fetchingTodos && todos === undefined) fetchTodos();
  }, [show, fetchingTodos, fetchTodos, todos]);

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
    setShowNewTodoForm(true);
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

  const handleNewTodoFormOpen = useCallback(() => {
    setShowNewTodoForm(true);
  }, []);

  const handleNewTodoFormClose = useCallback(() => {
    setShowNewTodoForm(false);
    setState((state) => ({
      ...state,
      editing: undefined
    }));
  }, []);

  return (
    <Drawer id="todos" title="Todo list" show={show} toggle={toggleDrawer}>
      <div className="mt-4 mb-8">
        <AnimatePresence>
          {showNewTodoForm ? (
            <AnimatePresence>
              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>
                <TodoForm
                  editing={todos?.find((t) => t.id === state.editing)}
                  onCancel={() => handleEdit(state.editing!, false)}
                  onHide={handleNewTodoFormClose}
                />
              </motion.div>
            </AnimatePresence>
          ) : (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1, transition: { delay: 0.025 } }}
              exit={{ opacity: 0 }}
              className={
                'flex px-2 py-2 cursor-pointer rounded-md text-gray-500 hover:text-gray-600 bg-gray-50 hover:bg-gray-100'
              }
              onClick={handleNewTodoFormOpen}
            >
              <Button icon={'PlusCircleIcon'} size={'sm'} />
              <p className={'ml-2'}>Add new todo</p>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <div className="px-3">
        {fetchingTodos && todos === undefined ? (
          <ListSkeleton count={5} fullWidth />
        ) : (
          <TodoList
            todos={todos || []}
            editing={state.editing}
            deleting={state.deleting}
            loading={state.loading}
            handleEdit={handleEdit}
            handleUpdate={handleUpdate}
            handleDelete={handleDelete}
          />
        )}
      </div>
    </Drawer>
  );
}
