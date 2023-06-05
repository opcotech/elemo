'use client';

import { Link } from '@/components/blocks/Link';
import { authOptions } from '@/lib/auth';
import { useSearchParams } from 'next/navigation';

type Error = {
  error: string;
  error_description: string;
};

const ERRORS: Record<ErrorType, Error> = {
  default: {
    error: 'Unknown error',
    error_description: 'An unknown error occurred. The issue may be temporary, so please try again later.'
  },
  configuration: {
    error: 'Configuration error',
    error_description:
      'An error occurred while loading the configuration. The issue is permanent, so please contact the administrator.'
  },
  accessdenied: {
    error: 'Access denied',
    error_description:
      'The access to the requested resource is denied. Please make sure you are logged in and have the required permissions.'
  },
  verification: {
    error: 'Email not verified',
    error_description: 'Email not verified. Please check your inbox for a verification email.'
  }
};

export type ErrorType = 'default' | 'configuration' | 'accessdenied' | 'verification';

export default function ErrorPage() {
  const signInUrl = authOptions.pages!.signIn;

  const searchParams = useSearchParams();
  const error = searchParams.get('error') as ErrorType | null;
  const { error: errorName, error_description: errorDescription } = (error && ERRORS[error]) || ERRORS.default;

  return (
    <div className="h-screen w-screen flex items-center">
      <div className="max-w-xl mx-auto text-center">
        <h2 className="mb-4">{errorName}</h2>
        <p className="mb-10">{errorDescription}</p>

        {signInUrl && error != 'configuration' && (
          <p>
            Try to <Link href={signInUrl}>sign in again</Link>.
          </p>
        )}

        {error == 'configuration' && (
          <p>
            Go back to the <Link href="/">home page</Link>.
          </p>
        )}
      </div>
    </div>
  );
}
