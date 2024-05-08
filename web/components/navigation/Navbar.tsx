'use client';

import { Fragment, useEffect, useMemo, useState } from 'react';
import { usePathname } from 'next/navigation';
import { Disclosure, Popover, Transition } from '@headlessui/react';
import { useSession } from 'next-auth/react';
import { concat, getInitials } from '@/lib/helpers';
import { Avatar } from '@/components/blocks/Avatar';
import { Button } from '@/components/blocks/Button';
import { Icon } from '@/components/blocks/Icon';
import { Link } from '@/components/blocks/Link';
import useStore from '@/store';

export interface NavigationItem {
  id: string;
  label: string;
  href: string;
  prefetch: boolean;
}

export interface UserNavigationItem extends NavigationItem {
  onClick?: () => void;
}

export interface NavbarProps {
  navigation: NavigationItem[];
  userNavigation: UserNavigationItem[];
}

export function Navbar({ navigation, userNavigation }: NavbarProps) {
  const [hasTodos, setHasTodos] = useState(false);
  const [hasNotifications, setHasNotifications] = useState(false);

  const [
    todos,
    notifications,
    fetchingTodos,
    fetchTodos,
    fetchingNotifications,
    fetchNotifications,
    toggleTodoDrawer,
    toggleNotificationDrawer
  ] = useStore((state) => [
    state.todos,
    state.notifications,
    state.fetchingTodos,
    state.fetchTodos,
    state.fetchingNotifications,
    state.fetchNotifications,
    () => state.toggleDrawer('todos'),
    () => state.toggleDrawer('notifications')
  ]);

  const { data: session } = useSession();
  const user = session?.user;

  const currentPath = '/' + usePathname()?.split('/')[1];

  function isCurrent(path: string) {
    return path === currentPath;
  }

  const userInitials = useMemo(() => {
    return getInitials(user?.name);
  }, [user?.name]);

  useEffect(() => {
    if (!fetchingTodos && !todos) fetchTodos();
  }, [fetchingTodos, fetchTodos, todos]);

  useEffect(() => {
    if (todos) setHasTodos(todos.some((t) => !t.completed));
  }, [todos]);

  useEffect(() => {
    if (!fetchingNotifications && !notifications) fetchNotifications();
  }, [fetchingNotifications, fetchNotifications, notifications]);

  useEffect(() => {
    if (notifications) setHasNotifications(notifications.some((n) => !n.read));
  }, [notifications]);

  return (
    <Disclosure id="navbar" as="nav" className="bg-gray-50 shadow-sm z-20">
      {({ open }) => (
        <>
          <div className="px-4 sm:px-6 lg:px-8">
            <div className="flex h-16 justify-between">
              <div className="flex">
                <div className="flex flex-shrink-0 items-center">
                  <span className="text-2xl">Elemo</span>
                </div>
                <div className="hidden sm:-my-px sm:ml-6 lg:ml-10 sm:flex sm:space-x-4">
                  {navigation.map((item) => (
                    <Link
                      id={item.id}
                      key={item.id}
                      href={item.href}
                      decorated={false}
                      className={concat(
                        isCurrent(item.href)
                          ? 'border-blue-500 text-blue-500'
                          : 'border-transparent  hover:border-gray-300',
                        'inline-flex items-center px-1 pt-2 border-b-2 text-sm font-medium'
                      )}
                      aria-current={isCurrent(item.href) ? 'page' : undefined}
                      {...(!item.prefetch && { prefetch: item.prefetch })}
                    >
                      {item.label}
                    </Link>
                  ))}
                </div>
              </div>
              <div className="hidden sm:ml-6 sm:flex sm:items-center">
                <Button icon="CheckCircleIcon" size="sm" onClick={toggleTodoDrawer} className="mr-2 relative">
                  {hasTodos && (
                    <span className="absolute top-0 right-0 block h-2 w-2 rounded-full bg-red-400 ring-1 ring-white" />
                  )}
                  <span className="sr-only">View todos</span>
                </Button>
                <Button icon="BellIcon" size="sm" onClick={toggleNotificationDrawer} className="mr-2 relative">
                  {hasNotifications && (
                    <span className="absolute top-0 right-0 block h-2 w-2 rounded-full bg-red-400 ring-1 ring-white" />
                  )}
                  <span className="sr-only">View notifications</span>
                </Button>

                <Popover id="navbar-user-dropdown" className="relative ml-1">
                  <Popover.Button
                    id="btn-avatar"
                    className="flex max-w-xs items-center rounded-full focus:outline-none text-gray-600 hover:text-black focus:ring-gray-600 focus:ring-2 focus:ring-offset-2"
                  >
                    <span className="sr-only">Open user menu</span>
                    <Avatar size="xs" initials={userInitials} src={user?.image ?? ''} />
                  </Popover.Button>
                  <Transition
                    as={Fragment}
                    enter="transition ease-out duration-200"
                    enterFrom="transform opacity-0 scale-95"
                    enterTo="transform opacity-100 scale-100"
                    leave="transition ease-in duration-75"
                    leaveFrom="transform opacity-100 scale-100"
                    leaveTo="transform opacity-0 scale-95"
                  >
                    <Popover.Panel className="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
                      <div className="px-4 py-3">
                        <div className="text-sm font-medium text-gray-800">{user?.name}</div>
                        <div className="text-xs text-gray-500">{user?.email}</div>
                      </div>

                      <div className="grid grid-cols-1">
                        {userNavigation.map((item) => (
                          <div id={item.id} key={item.id} className="hover:bg-gray-100" role="menuitem">
                            <Link
                              href={item.href}
                              decorated={false}
                              className="block px-4 py-2 text-sm text-gray-700"
                              onClick={item.onClick}
                              {...(!item.prefetch && { prefetch: item.prefetch })}
                            >
                              {item.label}
                            </Link>
                          </div>
                        ))}
                      </div>
                    </Popover.Panel>
                  </Transition>
                </Popover>
              </div>
              <div className="-mr-2 flex items-center sm:hidden">
                <Disclosure.Button className="inline-flex items-center justify-center rounded-md bg-gray-50 p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2">
                  <span className="sr-only">Open main menu</span>
                  {open ? (
                    <Icon variant="XMarkIcon" className="block h-6 w-6" aria-hidden="true" />
                  ) : (
                    <Icon variant="Bars3Icon" className="block h-6 w-6" aria-hidden="true" />
                  )}
                </Disclosure.Button>
              </div>
            </div>
          </div>

          <Disclosure.Panel className="sm:hidden">
            <div className="space-y-1 pt-2 pb-3">
              {navigation.map((item) => (
                <Disclosure.Button
                  id={item.id}
                  key={item.id}
                  as={Link}
                  href={item.href}
                  decorated={false}
                  className={concat(
                    isCurrent(item.href)
                      ? 'bg-gray-50 border-gray-500 text-gray-700'
                      : 'border-transparent text-gray-600 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-800',
                    'block pl-3 pr-4 py-2 border-l-4 text-base font-medium'
                  )}
                  aria-current={isCurrent(item.href) ? 'page' : undefined}
                  {...(!item.prefetch && { prefetch: item.prefetch })}
                >
                  {item.label}
                </Disclosure.Button>
              ))}
            </div>
            <div className="border-t border-gray-200 pt-4 pb-3">
              <div className="flex items-center px-4">
                <div className="flex-shrink-0">
                  <Avatar size="sm" initials={userInitials} src={user?.image ?? ''} />
                </div>
                <div className="ml-3">
                  <div className="text-base font-medium text-gray-800">{user?.name}</div>
                  <div className="text-sm text-gray-500">{user?.email}</div>
                </div>
                <div className="flex flex-1 justify-end">
                  <Button
                    icon="CheckCircleIcon"
                    size="sm"
                    onClick={toggleTodoDrawer}
                    className="mr-2 relative flex-shrink-0"
                  >
                    {hasTodos && (
                      <span className="absolute top-0 right-0 block h-2 w-2 rounded-full bg-red-400 ring-1 ring-white" />
                    )}
                    <span className="sr-only">View todos</span>
                  </Button>
                  <Button
                    icon="BellIcon"
                    size="sm"
                    onClick={toggleNotificationDrawer}
                    className="mr-2 relative flex-shrink-0"
                  >
                    {hasNotifications && (
                      <span className="absolute top-0 right-0 block h-2 w-2 rounded-full bg-red-400 ring-1 ring-white" />
                    )}
                    <span className="sr-only">View notifications</span>
                  </Button>
                </div>
              </div>
              <div className="mt-3 space-y-1">
                {userNavigation.map((item) => (
                  <Disclosure.Button
                    id={item.id}
                    key={item.id}
                    as={Link}
                    href={item.href}
                    decorated={false}
                    className="block px-4 py-2 text-base font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-800"
                    onClick={item.onClick}
                    {...(!item.prefetch && { prefetch: item.prefetch })}
                  >
                    {item.label}
                  </Disclosure.Button>
                ))}
              </div>
            </div>
          </Disclosure.Panel>
        </>
      )}
    </Disclosure>
  );
}
