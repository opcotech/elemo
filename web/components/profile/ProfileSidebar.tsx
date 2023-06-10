import { User } from '@/lib/api';
import { Avatar } from '@/components/blocks/Avatar';
import { getInitials } from '@/lib/helpers';
import { Link } from '@/components/blocks/Link';
import { Icon } from '@/components/blocks/Icon';
import { Divider } from '@/components/blocks/Divider';
import { LANGUAGES } from '@/lib/constants';

export interface ProfileSidebarProps {
  user: User;
  isCurrentUser: boolean;
}

export function ProfileSidebar({ user, isCurrentUser }: ProfileSidebarProps) {
  const displayName = user.first_name && user.last_name ? `${user.first_name} ${user.last_name}` : user.username;

  return (
    <aside className="bg-gray-50 px-4 py-6 sm:px-6 lg:pl-8 xl:w-80 xl:shrink-0 xl:pl-6 rounded-md border border-gray-100">
      <section className="mb-4">
        <div className={'sm:flex'}>
          <div className="mb-4 flex-shrink-0 sm:mb-0 sm:mr-4">
            <Avatar size={'lg'} src={user.picture || ''} initials={getInitials(displayName)} />
          </div>
          <div>
            <h4 className="text-lg font-medium">{displayName}</h4>
            {user.title && <p className={'text-sm mt-0.5'}>{user.title}</p>}
            {isCurrentUser && (
              <p className={'text-sm mt-1'}>
                <Link href={'/settings'} className={'flex space-x-1.5'}>
                  <Icon size={'xs'} variant={'PencilSquareIcon'} className={'mt-0.5'} />
                  <span>Edit my profile</span>
                </Link>
              </p>
            )}
          </div>
        </div>
      </section>
      <Divider className={'pb-4'} />
      <section className={'my-4'}>
        <dl className={'space-y-3'}>
          <div className="px-4 sm:px-0">
            <dt className="text-sm font-medium leading-6 text-gray-900">Username</dt>
            <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">@{user.username}</dd>
          </div>
          {user.first_name && user.last_name && (
            <div className="px-4 sm:px-0">
              <dt className="text-sm font-medium leading-6 text-gray-900">Full name</dt>
              <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">
                {user.first_name}&nbsp;{user.last_name}
              </dd>
            </div>
          )}
          <div className="px-4 sm:px-0">
            <dt className="text-sm font-medium leading-6 text-gray-900">Email address</dt>
            <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">
              <Link href={'mailto:' + user.email}>{user.email}</Link>
            </dd>
          </div>
          {user.phone && (
            <div className="px-4 sm:px-0">
              <dt className="text-sm font-medium leading-6 text-gray-900">Phone number</dt>
              <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">
                <Link href={'tel:' + user.phone}>{user.phone}</Link>
              </dd>
            </div>
          )}
        </dl>
      </section>
      <Divider className={'pb-4'} />
      <section className={'my-4'}>
        <dl className={'space-y-3'}>
          <div className="px-4 sm:px-0">
            <dt className="text-sm font-medium leading-6 text-gray-900">Address</dt>
            <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">
              <p>{user.address ? user.address : 'No address set.'}</p>
            </dd>
          </div>
        </dl>
      </section>
      {user.languages.length > 0 && (
        <>
          <Divider className={'pb-4'} />
          <section className={'my-4'}>
            <dl className={'space-y-3'}>
              <div className="px-4 sm:px-0">
                <dt className="text-sm font-medium leading-6 text-gray-900">Languages</dt>
                <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">
                  <ul className={'list-disc list-inside'}>
                    {user.languages?.map((language) => (
                      <li key={language}>
                        <span className={'-ml-1'}>{LANGUAGES.find((l) => l.code === language)?.name}</span>
                      </li>
                    ))}
                  </ul>
                </dd>
              </div>
            </dl>
          </section>
        </>
      )}
      {user.bio && (
        <>
          <Divider className={'pb-4'} />
          <section className={'my-4'}>
            <dl className={'space-y-3'}>
              <div className="px-4 sm:px-0">
                <dt className="text-sm font-medium leading-6 text-gray-900">Bio</dt>
                <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">
                  <p>{user.bio}</p>
                </dd>
              </div>
            </dl>
          </section>
        </>
      )}
      <Divider className={'pb-4'} />
      <section className={'mt-4'}>
        <dl className={'space-y-3'}>
          <div className="px-4 sm:px-0">
            <dt className="text-sm font-medium leading-6 text-gray-900">Links</dt>
            <dd className="mt-0.5 text-sm leading-6 text-gray-700 sm:mt-1">
              <ul className={'list-disc list-inside'}>
                {user.links?.map((link) => (
                  <li key={link}>
                    <Link href={link} className={'-ml-1'}>
                      {link}
                    </Link>
                  </li>
                ))}
              </ul>
            </dd>
          </div>
        </dl>
      </section>
    </aside>
  );
}
