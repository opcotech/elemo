import { toCapitalCase } from '@/lib/helpers';
import { ApiError, SystemLicense, SystemService, SystemVersion } from '@/lib/api';

export const dynamic = 'force-dynamic';
export const metadata = {
  title: 'System settings | Elemo'
};

async function getData() {
  let license: SystemLicense | undefined = undefined;
  let version: SystemVersion | undefined = undefined;

  try {
    license = await SystemService.v1SystemLicense();
  } catch (e) {
    const err = e as ApiError;
    if (err.status !== 403) throw err.message;
  }

  try {
    version = await SystemService.v1SystemVersion();
  } catch (e) {
    throw (e as ApiError).message;
  }

  return {
    license,
    version
  };
}

export default async function SystemSettings() {
  const { license, version } = await getData();

  return (
    <div className="space-y-8 divide-y divide-gray-100">
      <section>
        <h2 className="text-base font-medium leading-7 text-gray-900">System roles</h2>
        <p className="mt-1 text-sm leading-6 text-gray-500">
          Assign pre-defined roles to users. These roles are used to determine what a user can do in the system and what
          they can see. These roles cannot be edited or deleted. For fine-grained control, use organization or project
          specific roles.
        </p>

        <div className="mt-8 space-y-6 text-sm leading-6">TODO</div>
      </section>

      {license && (
        <section className="pt-8">
          <h2 className="text-base font-medium leading-7 text-gray-900">License</h2>
          <p className="mt-1 text-sm leading-6 text-gray-500">View the license information for this instance.</p>

          <dl className="mt-8 space-y-6 text-sm leading-6">
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">License ID</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{license.id}</div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Licensee</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{license.organization}</div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Contact email</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{license.email}</div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Expiry date</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{new Date(license.expires_at).toLocaleString()}</div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Quotas</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">
                  <ul>
                    {Object.keys(license.quotas).map((quota, i) => (
                      <li key={i}>
                        {toCapitalCase(quota.replace('_', ' '))} ({(license.quotas as any)[quota]})
                      </li>
                    ))}
                  </ul>
                </div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Features</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">
                  <ul>
                    {license.features.map((feature, i) => (
                      <li key={i}>{toCapitalCase(feature.replace('_', ' '))}</li>
                    ))}
                  </ul>
                </div>
              </dd>
            </div>
          </dl>
        </section>
      )}

      {version && (
        <section className="pt-8">
          <h2 className="text-base font-medium leading-7 text-gray-900">Version</h2>
          <p className="mt-1 text-sm leading-6 text-gray-500">View the current version of the application.</p>

          <dl className="mt-8 space-y-6 text-sm leading-6">
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Semantic version</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{version.version}</div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">VCS commit</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{version.commit}</div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Release date</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{new Date(version.date).toLocaleString()}</div>
              </dd>
            </div>
            <div className="sm:flex">
              <dt className="font-medium text-gray-900 sm:w-40 sm:flex-none sm:pr-6">Go version</dt>
              <dd className="mt-1 flex justify-between gap-x-6 sm:mt-0 sm:flex-auto">
                <div className="text-gray-900">{version.go_version}</div>
              </dd>
            </div>
          </dl>
        </section>
      )}
    </div>
  );
}
