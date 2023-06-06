import { Organization, OrganizationStatus } from '@/lib/api';
import { Avatar } from '@/components/blocks/Avatar';
import { getInitials, toCapitalCase } from '@/lib/helpers';
import { Link } from '@/components/blocks/Link';
import { Menu, Transition } from '@headlessui/react';
import { Icon } from '@/components/blocks/Icon';
import { Fragment } from 'react';
import { Badge } from '@/components/blocks/Badge';

export interface OrganizationListItemProps {
  organization: Organization;
  canView?: boolean;
  canEdit?: boolean;
  canDelete?: boolean;
}

export function OrganizationListItem({ organization, canView, canEdit, canDelete }: OrganizationListItemProps) {
  return (
    <li className="flex justify-between gap-x-6 py-5">
      <div className="flex gap-x-4">
        <Avatar
          size="md"
          src={organization.logo || ''}
          initials={getInitials(organization.name)}
          grayscale={organization.status === OrganizationStatus.DELETED}
        />
        <div className="min-w-0 flex-auto">
          <p className="text-sm font-medium leading-6">
            {canView || canEdit || canDelete ? (
              <Link href={`/settings/organizations/${organization.id}`}>{organization.name}</Link>
            ) : (
              organization.name
            )}
          </p>
          <p className="mt-1 flex text-xs leading-5 text-gray-500">
            <Link href={`mailto:${organization.email}`}>{organization.email}</Link>
          </p>
        </div>
      </div>
      <div className="flex items-center gap-x-6">
        <div className="hidden sm:flex sm:flex-col sm:items-end">
          <p className="text-sm leading-6">
            <Badge
              title={toCapitalCase(organization.status)}
              variant={organization.status === OrganizationStatus.ACTIVE ? 'success' : 'danger'}
            />
          </p>
          <p className="mt-1 text-xs leading-5 text-gray-500">
            {organization.members.length}&nbsp;{organization.members.length === 1 ? 'member' : 'members'}
          </p>
        </div>
        {canView && ((canEdit && organization.status !== OrganizationStatus.DELETED) || canDelete) && (
          <Menu as="div" className="relative flex-none">
            <Menu.Button className="-m-2.5 block p-2.5 text-gray-500 hover:text-gray-900">
              <span className="sr-only">Open options</span>
              <Icon size={'sm'} variant="EllipsisVerticalIcon" aria-hidden="true" />
            </Menu.Button>
            <Transition
              as={Fragment}
              enter="transition ease-out duration-100"
              enterFrom="transform opacity-0 scale-95"
              enterTo="transform opacity-100 scale-100"
              leave="transition ease-in duration-75"
              leaveFrom="transform opacity-100 scale-100"
              leaveTo="transform opacity-0 scale-95"
            >
              <Menu.Items className="absolute right-0 z-10 mt-2 w-32 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none">
                {canEdit && organization.status !== OrganizationStatus.DELETED && (
                  <Menu.Item>
                    <Link
                      href={`/settings/organizations/${organization.id}/edit`}
                      decorated={false}
                      className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                      role={'menuitem'}
                    >
                      Edit
                    </Link>
                  </Menu.Item>
                )}
                {canDelete && (
                  <Menu.Item>
                    {organization.status !== OrganizationStatus.DELETED ? (
                      <Link
                        href={`/settings/organizations/${organization.id}/edit#delete`}
                        decorated={false}
                        className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                        role={'menuitem'}
                      >
                        Delete
                      </Link>
                    ) : (
                      <Link
                        href={`/settings/organizations/${organization.id}/edit#restore`}
                        decorated={false}
                        className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                        role={'menuitem'}
                      >
                        Restore
                      </Link>
                    )}
                  </Menu.Item>
                )}
              </Menu.Items>
            </Transition>
          </Menu>
        )}
      </div>
    </li>
  );
}
