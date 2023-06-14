import { UserService } from '@/lib/api';
import { ChangePasswordForm } from '@/components/settings/ChangePasswordForm';

export const dynamic = 'force-dynamic';
export const metadata = {
  title: 'Security settings | Elemo'
};

async function getData() {
  const user = await UserService.v1UserGet('me');
  return { user };
}

export default async function SecuritySettings() {
  const { user } = await getData();

  return (
    <div className="space-y-8 divide-y divide-gray-100">
      <section>
        <h2 className="text-base font-medium leading-7 text-gray-900">Password</h2>
        <p className="mt-1 text-sm leading-6 text-gray-500">
          Change your password. Please make sure that it is hard to guess to keep your account safe.
        </p>

        <div className="mt-10 space-y-6 text-sm leading-6">
          <ChangePasswordForm userId={user.id} />
        </div>
      </section>
    </div>
  );
}
