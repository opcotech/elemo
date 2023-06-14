import { UserService } from '@/lib/api';
import { UpdateUserProfileForm } from '@/components/settings/UpdateUserProfileForm';
import { UpdateUserContactForm } from '@/components/settings/UpdateUserContactForm';
import { UpdateUserAddressForm } from '@/components/settings/UpdateUserAddressForm';

export const dynamic = 'force-dynamic';
export const metadata = {
  title: 'Settings | Elemo'
};

async function getData() {
  const user = await UserService.v1UserGet('me');
  return { user };
}

export default async function Settings() {
  const { user } = await getData();

  return (
    <div className="space-y-8 divide-y divide-gray-100">
      <section>
        <h2 className="text-base font-medium leading-7 text-gray-900">Profile</h2>
        <p className="mt-1 text-sm leading-6 text-gray-500">This information will be displayed on your profile.</p>

        <div className="mt-10 space-y-6 text-sm leading-6">
          <UpdateUserProfileForm
            userId={user.id}
            defaultValues={{
              username: user.username,
              first_name: user.first_name || undefined,
              last_name: user.last_name || undefined,
              picture: user.picture || undefined,
              title: user.title || undefined,
              bio: user.bio || undefined,
              languages: user.languages
            }}
          />
        </div>
      </section>

      <section className="pt-8">
        <h2 className="text-base font-medium leading-7 text-gray-900">Contact</h2>
        <p className="mt-1 text-sm leading-6 text-gray-500">
          Share your contact information with other users to make it easier to collaborate.
        </p>

        <div className="mt-10 space-y-6 text-sm leading-6">
          <UpdateUserContactForm
            userId={user.id}
            defaultValues={{
              email: user.email,
              phone: user.phone || undefined,
              links: user.links || undefined
            }}
          />
        </div>
      </section>

      <section className="pt-8">
        <h2 className="text-base font-medium leading-7 text-gray-900">Address</h2>
        <p className="mt-1 text-sm leading-6 text-gray-500">
          Set your work location to let your teammates know where you are working from.
        </p>

        <div className="mt-10 space-y-6 text-sm leading-6">
          <UpdateUserAddressForm
            userId={user.id}
            defaultValues={{
              address: user.address || undefined
            }}
          />
        </div>
      </section>
    </div>
  );
}
