import { getServerSession } from 'next-auth';
import { UsersService } from '@/lib/api';
import { authOptions } from '@/lib/auth';
import { ProfileSidebar } from '@/components/profile/ProfileSidebar';

export const dynamic = 'force-dynamic';
export async function generateMetadata({ params: { userId } }: ProfilePageProps) {
  const user = await UsersService.v1UserGet(userId || 'me');

  return {
    title: `${user.first_name} ${user.last_name} | Profile | Elemo`
  };
}

export interface ProfilePageProps {
  params: {
    userId?: string;
  };
}

export async function getData({ params: { userId } }: ProfilePageProps) {
  const user = await UsersService.v1UserGet(userId || 'me');
  return { user };
}

export default async function Profile({ params: { userId } }: ProfilePageProps) {
  const { user } = await getData({ params: { userId } });
  const session = await getServerSession(authOptions);
  const isCurrentUser = session?.user?.id === user.id;

  return (
    <div className="grow lg:flex lg:space-x-4">
      <ProfileSidebar user={user} isCurrentUser={isCurrentUser} />
      <main className="px-4 py-6 sm:px-6 lg:pl-8 xl:flex-1 xl:pl-6"></main>
    </div>
  );
}
