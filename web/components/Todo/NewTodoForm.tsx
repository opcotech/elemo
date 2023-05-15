'use client';

import { unknown, z } from 'zod';
import { Listbox, Transition } from '@headlessui/react';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ChangeEvent } from 'react';
import { Fragment, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';

import Button from '@/components/Button';
import Icon from '@/components/Icon';
import Link from '@/components/Link';
import { concat, formatErrorMessage, toCapitalCase } from '@/helpers';
import useStore from '@/store';
import { useSession } from 'next-auth/react';
import { Todo, TodoPriority } from '@/lib/api';

const PRIORITY_ORDER: TodoPriority[] = [
  TodoPriority.Normal,
  TodoPriority.Important,
  TodoPriority.Urgent,
  TodoPriority.Critical
];

const PRIORITY_COLORS: { [key in TodoPriority]: string } = {
  normal: 'text-gray-700',
  important: 'text-blue-600',
  urgent: 'text-yellow-600',
  critical: 'text-red-600'
};

const CREATE_TODO_SCHEMA = z.object({
  title: z.string().min(3, 'Title is required').max(250, 'Title must be less than 250 characters.'),
  description: z
    .string()
    .min(10, 'Description must be at least 10 characters.')
    .max(500, 'Description must be less than 500 characters.')
    .optional()
    .or(z.literal('')),
  completed: z.boolean().default(false),
  priority: z.enum([PRIORITY_ORDER[0], ...PRIORITY_ORDER.slice(1)]).default(TodoPriority.Normal),
  due_date: z.string().optional()
});

export interface NewTodoFormProps {
  editing: Todo | undefined;
  onCancel: () => void;
}

export default function NewTodoForm(props: NewTodoFormProps) {
  const { data: session } = useSession();

  const [loading, setLoading] = useState(false);

  const createTodo = useStore((state) => state.createTodo);
  const updateTodo = useStore((state) => state.updateTodo);

  const isEditing = props.editing?.id !== undefined;
  const todoId = props.editing?.id || undefined;

  const [priority, setPriority] = useState<TodoPriority>(props.editing?.priority || PRIORITY_ORDER[0]);
  const [date, setDate] = useState<Date | undefined>(
    props.editing?.due_date ? new Date(props.editing?.due_date) : undefined
  );

  const {
    register,
    handleSubmit,
    reset,
    clearErrors,
    setValue,
    getValues,
    formState: { errors }
  } = useForm<Todo>({
    resolver: zodResolver(CREATE_TODO_SCHEMA)
  });

  function resetFormState() {
    setPriority(PRIORITY_ORDER[0]);
    setDate(new Date());
    reset();
    clearErrors();
    props = { editing: undefined, onCancel: props.onCancel };
  }

  async function onSubmit(todo: Todo) {
    setLoading(true);

    if (!todoId) {
      await createTodo({ ...todo, owned_by: session!.user!.id });
    } else {
      await updateTodo(todoId, todo);
      handleCancel();
    }

    resetFormState();
    setLoading(false);
  }

  function handleCancel() {
    props.onCancel();
    resetFormState();
  }

  function handlePriorityChange(priority: TodoPriority) {
    setPriority(priority);
    setValue('priority', priority);
  }

  function handleDateChange(e: ChangeEvent<HTMLInputElement>) {
    if (!e.target.value) {
      setDate(undefined);
      setValue('due_date', undefined);
      return;
    }

    const date = new Date(e.target.value);
    setDate(date);
    setValue('due_date', date.toISOString());
  }

  // Set default values for form fields even when editing
  useEffect(() => {
    setPriority(props.editing?.priority || PRIORITY_ORDER[0]);
    setDate(props.editing?.due_date ? new Date(props.editing?.due_date) : new Date());
  }, [props.editing?.priority, props.editing?.due_date]);

  useEffect(() => {
    setValue('title', props.editing?.title || getValues('title'));
    setValue('description', props.editing?.description || getValues('description'));
    setValue('priority', priority || getValues('priority'));
    setValue('completed', false || getValues('completed'));
    setValue('due_date', date?.toISOString() || getValues('due_date'));
  }, [props.editing?.title, props.editing?.description, priority, date, setValue, getValues]);

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
              name="due_date"
              className="relative inverse-datepicker inline-flex items-center whitespace-nowrap rounded-full border-none bg-gray-50 py-2 px-2 text-sm text-gray-900 hover:bg-gray-100 sm:px-3 focus:ring-0"
              autoComplete="off"
              onChange={handleDateChange}
              aria-invalid={errors.due_date ? 'true' : 'false'}
              value={date?.toISOString().split('T')[0]}
              min={new Date().toISOString().split('T')[0]}
            />
          </div>
        </div>
      </div>
    </form>
  );
}
