'use client';

import { AnimatePresence, motion } from 'framer-motion';
import { memo, useCallback, useEffect, useRef, useState } from 'react';
import Drawer from '@/components/Drawer';
import { ListSkeleton } from '@/components/Skeleton';
import useStore from '@/store';
import TodoItem from './TodoItem';
import { UpdateTodoInput } from '@/store/todoSlice';
import NewTodoForm from './NewTodoForm';
import { IconButton } from '@/components/Button';

const MemoizedTodoItem = memo(TodoItem);

export interface TodoDrawerState {
  editing?: string;
  deleting: { id: string; timer: NodeJS.Timer | undefined }[];
  loading: string[];
}

export default function TodoDrawer() {
  const todosRendered = useRef(false);

  const [showNewTodoForm, setShowNewTodoForm] = useState(false);

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

  useEffect(() => {
    todosRendered.current = true;
  }, []);

  const setLoading = useCallback(
    (id: string, loading: boolean) =>
      setState((state) => ({
        ...state,
        loading: loading ? [...state.loading, id] : state.loading.filter((l) => l !== id)
      })),
    []
  );

  const handleUpdate = useCallback(
    async (id: string, todo: UpdateTodoInput) => {
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

  const handleDrawerClose = useCallback(() => {
    handleNewTodoFormClose();
    toggleDrawer('showTodos');
  }, [toggleDrawer, handleNewTodoFormClose]);

  return (
    <Drawer id="showTodos" title="Todos" show={drawers.showTodos} toggle={handleDrawerClose}>
      <div className="mb-6">
        <AnimatePresence>
          {showNewTodoForm ? (
            <AnimatePresence>
              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>
                <NewTodoForm
                  editing={todos.find((t) => t.id === state.editing)}
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
              <IconButton icon={'PlusCircleIcon'} />
              <p className={'ml-2'}>Add new todo</p>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {todos.length === 0 && fetchingTodos && <ListSkeleton count={5} />}

      <ul className="divide-y divide-gray-200">
        {todos.length === 0 && !fetchingTodos && <li className="text-center text-gray-500">No todos found.</li>}

        {todos.map((item, i) => (
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
