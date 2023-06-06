'use client';

import { z } from 'zod';
import { Fragment, useEffect, useState } from 'react';
import { Listbox, Transition } from '@headlessui/react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';

import { Button } from '@/components/blocks/Button';
import { Icon } from '@/components/blocks/Icon';
import { Link } from '@/components/blocks/Link';
import { concat, formatErrorMessage, toCapitalCase } from '@/lib/helpers';
import useStore from '@/store';
import { $Todo, Todo, TodoPriority } from '@/lib/api';

const PRIORITY_ORDER: TodoPriority[] = [
  TodoPriority.NORMAL,
  TodoPriority.IMPORTANT,
  TodoPriority.URGENT,
  TodoPriority.CRITICAL
];

const PRIORITY_COLORS: { [key in TodoPriority]: string } = {
  normal: 'text-gray-700',
  important: 'text-blue-600',
  urgent: 'text-yellow-600',
  critical: 'text-red-600'
};

const CREATE_TODO_SCHEMA = z.object({
  title: z
    .string()
    .min($Todo.properties.title.minLength, 'Title is required')
    .max($Todo.properties.title.maxLength, 'Title must be less than 250 characters.'),
  description: z
    .string()
    .min($Todo.properties.description.minLength, 'Description must be at least 10 characters.')
    .max($Todo.properties.description.maxLength, 'Description must be less than 500 characters.')
    .optional()
    .or(z.literal('')),
  completed: z.boolean().default(false),
  priority: z.enum([PRIORITY_ORDER[0], ...PRIORITY_ORDER.slice(1)]).default(TodoPriority.NORMAL),
  due_date: z.string().optional()
});

export interface NewTodoFormProps {
  editing: Todo | undefined;
  onCancel: () => void;
  onHide: () => void;
}

export function TodoForm(props: NewTodoFormProps) {
  const [loading, setLoading] = useState(false);

  const createTodo = useStore((state) => state.createTodo);
  const updateTodo = useStore((state) => state.updateTodo);

  const isEditing = props.editing?.id !== undefined;
  const todoId = props.editing?.id || undefined;

  const [priority, setPriority] = useState<TodoPriority>(props.editing?.priority || PRIORITY_ORDER[0]);

  const {
    register,
    handleSubmit,
    reset,
    clearErrors,
    setValue,
    getValues,
    setFocus,
    formState: { errors }
  } = useForm<Todo>({
    resolver: zodResolver(CREATE_TODO_SCHEMA)
  });

  function resetFormState() {
    setPriority(PRIORITY_ORDER[0]);
    reset();
    clearErrors();
    props = { editing: undefined, onCancel: props.onCancel, onHide: props.onHide };
  }

  async function onSubmit(todo: Todo) {
    setLoading(true);

    // Fix the due_date field format
    if (todo.due_date) {
      todo = { ...todo, due_date: new Date(todo.due_date).toISOString() };
    } else {
      todo = { ...todo, due_date: null };
    }

    if (!todoId) {
      await createTodo({
        ...todo,
        description: todo.description?.trim() || undefined,
        due_date: todo.due_date || undefined
      });
    } else {
      await updateTodo(todoId, {
        ...todo,
        description: todo.description?.trim() || undefined
      });
      handleCancel();
    }

    resetFormState();
    setLoading(false);
  }

  function handleCancel() {
    props.onCancel();
    resetFormState();
  }

  function handleHide() {
    props.onHide();
    resetFormState();
  }

  function handlePriorityChange(priority: TodoPriority) {
    setPriority(priority);
    setValue('priority', priority);
  }

  // Set default values for form fields even when editing
  useEffect(() => {
    setPriority(props.editing?.priority || PRIORITY_ORDER[0]);
  }, [props.editing?.priority, props.editing?.due_date]);

  useEffect(() => {
    setFocus('title');
    setValue('title', props.editing?.title || getValues('title'));
    setValue('description', props.editing?.description || getValues('description'));
    setValue('priority', priority || getValues('priority'));
    setValue('completed', getValues('completed'));
    setValue('due_date', props.editing?.due_date?.split('T')[0] || getValues('due_date'));
  }, [
    props.editing?.title,
    props.editing?.description,
    priority,
    setValue,
    getValues,
    setFocus,
    props.editing?.due_date
  ]);

  return (
    <form id="form-add-todo-item" action="web/components/todo#" className="relative" onSubmit={handleSubmit(onSubmit)}>
      <div
        className={
          'overflow-hidden rounded-lg border border-gray-300 shadow-sm focus-within:border-gray-500 focus-within:ring-1 focus-within:ring-gray-500'
        }
      >
        <label htmlFor="title" className="sr-only">
          Title
        </label>
        <input
          id="title"
          type="text"
          className="block w-full border-0 pt-2.5 text-lg font-medium placeholder-gray-500 focus:ring-0"
          placeholder="Title"
          autoComplete="off"
          aria-invalid={errors.title ? 'true' : 'false'}
          aria-describedby={errors.title ? 'title-error' : undefined}
          required={true}
          {...register('title')}
        />

        <label htmlFor="description" className="sr-only">
          Description
        </label>
        <textarea
          rows={3}
          id="description"
          className="block w-full resize-none border-0 py-0 placeholder-gray-500 focus:ring-0 sm:text-sm"
          placeholder="Today I'll complete..."
          autoComplete="off"
          aria-invalid={errors.description ? 'true' : 'false'}
          aria-describedby={errors.description ? 'description-error' : undefined}
          {...register('description')}
        />

        {Object.entries(errors).filter(([, value]) => value.message).length > 0 && (
          <div className="space-y-2 mt-4 px-3">
            {Object.entries(errors).map(([key, value]) => (
              <p id={`${key}-error`} key={key} className="text-sm text-red-600">
                {formatErrorMessage(key, value.message)}
              </p>
            ))}
          </div>
        )}

        {/* Spacer element to match the height of the toolbar */}
        <div aria-hidden="true">
          <div className="py-2">
            <div className="h-4" />
          </div>
          <div className="h-px" />
          <div className="py-2">
            <div className="py-px">
              <div className="h-4" />
            </div>
          </div>
        </div>
      </div>

      <div className="absolute inset-x-px bottom-0">
        <div className="flex items-center justify-between space-x-3 border-t border-gray-200 px-2 py-2 sm:px-3">
          <div className="flex items-center">
            <Button variant="secondary" loading={loading} type="submit">
              {isEditing ? 'Update' : 'Add'}
            </Button>
            {isEditing && (
              <Link className={'ml-3 text-sm'} onClick={handleCancel}>
                Cancel
              </Link>
            )}
            {!isEditing && (
              <Link className={'ml-3 text-sm'} onClick={handleHide}>
                Hide
              </Link>
            )}
          </div>
          <div className="flex items-center space-x-2">
            <Listbox as="div" value={priority} onChange={handlePriorityChange} className="flex-shrink-0">
              {({ open }) => (
                <>
                  <Listbox.Label className="sr-only"> Add a label </Listbox.Label>
                  <div className="relative">
                    <Listbox.Button
                      id="btn-todo-priority"
                      className="relative inline-flex items-center whitespace-nowrap rounded-full bg-gray-50 py-2 px-4 text-sm text-gray-500 hover:bg-gray-100"
                    >
                      <Icon
                        size={'xs'}
                        variant="FlagIcon"
                        className={concat(PRIORITY_COLORS[priority], 'h-4 w-4 flex-shrink-0 sm:-ml-1')}
                        aria-hidden="true"
                      />
                      <span className="hidden truncate sm:ml-2 sm:block text-gray-900">{toCapitalCase(priority)}</span>
                    </Listbox.Button>

                    <Transition
                      show={open}
                      as={Fragment}
                      enter="transition ease-out duration-200"
                      enterFrom="transform opacity-0 scale-95"
                      enterTo="transform opacity-100 scale-100"
                      leave="transition ease-in duration-75"
                      leaveFrom="transform opacity-100 scale-100"
                      leaveTo="transform opacity-0 scale-95"
                    >
                      <Listbox.Options
                        id="menu-todo-priority"
                        className="absolute -left-8 right-0 z-10 mt-1 max-h-56 w-52 overflow-auto rounded-lg bg-white py-3 text-base shadow ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm sm:left-auto"
                      >
                        {PRIORITY_ORDER.map((label) => (
                          <Listbox.Option
                            key={label}
                            className={({ active }) =>
                              concat(
                                active ? 'bg-gray-100' : 'bg-white',
                                'relative cursor-default select-none py-2 px-3'
                              )
                            }
                            value={label}
                            data-value={label}
                          >
                            <div className="flex items-center">
                              <span className="block truncate">{toCapitalCase(label)}</span>
                            </div>
                          </Listbox.Option>
                        ))}
                      </Listbox.Options>
                    </Transition>
                  </div>
                </>
              )}
            </Listbox>

            <input
              id="due_date"
              type="date"
              className="relative inverse-datepicker inline-flex items-center whitespace-nowrap rounded-full border-none bg-gray-50 py-2 px-2 text-sm text-gray-900 hover:bg-gray-100 sm:px-3 focus:ring-0"
              autoComplete="off"
              aria-invalid={errors.due_date ? 'true' : 'false'}
              {...register('due_date')}
              min={new Date().toISOString().split('T')[0]}
            />
          </div>
        </div>
      </div>
    </form>
  );
}
