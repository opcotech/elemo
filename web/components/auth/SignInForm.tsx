'use client';

import { useSearchParams } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';
import { Button } from '@/components/blocks/Button';
import { getCsrfToken } from 'next-auth/react';

export type SignInErrorTypes =
  | 'Signin'
  | 'OAuthSignin'
  | 'OAuthCallback'
  | 'OAuthCreateAccount'
  | 'EmailCreateAccount'
  | 'Callback'
  | 'OAuthAccountNotLinked'
  | 'EmailSignin'
  | 'CredentialsSignin'
  | 'SessionRequired'
  | 'default';

const ERRORS: Record<SignInErrorTypes, string> = {
  Signin: 'Try signing in with a different account.',
  OAuthSignin: 'Try signing in with a different account.',
  OAuthCallback: 'Try signing in with a different account.',
  OAuthCreateAccount: 'Try signing in with a different account.',
  EmailCreateAccount: 'Try signing in with a different account.',
  Callback: 'Try signing in with a different account.',
  OAuthAccountNotLinked: 'To confirm your identity, sign in with the same account you used originally.',
  EmailSignin: 'The e-mail could not be sent.',
  CredentialsSignin: 'Sign in failed. Check the details you provided are correct.',
  SessionRequired: 'Please sign in to access this page.',
  default: 'Unable to sign in.'
};

export function SignInForm() {
  const form = useRef<HTMLFormElement>(null);
  const [csrfToken, setCSRFToken] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const searchParams = useSearchParams();
  const error = searchParams.get('error') as SignInErrorTypes | null;

  useEffect(() => {
    getCsrfToken().then((token) => setCSRFToken(token || ''));
  }, []);

  async function handleSubmit() {
    setSubmitting(true);
    form.current?.submit();
  }

  return (
    <div className="mt-8">
      {error && (
        <div className="rounded-md bg-red-50 p-4">
          <div className="">
            <h3 className="text-base font-medium text-red-800">Failed to sign in!</h3>
            <div className="mt-2 text-base text-red-700">
              <p>{ERRORS[error]}</p>
            </div>
          </div>
        </div>
      )}

      <div className="mt-6">
        <form ref={form} className="space-y-6" method="POST" action="/api/auth/callback/credentials">
          <input name="csrfToken" type="hidden" defaultValue={csrfToken} />
          <div>
            <label htmlFor="username" className="block font-medium leading-6 text-gray-900">
              Email address
            </label>
            <div className="mt-2">
              <input
                id="username"
                name="username"
                type="email"
                required
                disabled={submitting}
                autoComplete="email"
                placeholder={'you@company.com'}
                className="block w-full rounded-md border-0 py-1.5 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:leading-6"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label htmlFor="password" className="block font-medium leading-6 text-gray-900">
              Password
            </label>
            <div className="mt-2">
              <input
                id="password"
                name="password"
                type="password"
                required
                disabled={submitting}
                autoComplete="current-password"
                placeholder={'********'}
                className="block w-full rounded-md border-0 py-1.5 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:leading-6"
              />
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <input
                id="remember-me"
                name="remember-me"
                type="checkbox"
                className="h-4 w-4 rounded border-gray-300 text-blue-500 focus:ring-blue-600"
                disabled={submitting}
              />
              <label htmlFor="remember-me" className="ml-2 block text-gray-900">
                Remember me
              </label>
            </div>

            <div>
              <a href="#" className="font-medium text-blue-500 hover:text-blue-600">
                Forgot your password?
              </a>
            </div>
          </div>

          <div>
            <Button type="submit" loading={submitting} className="w-full" onSubmit={handleSubmit}>
              Sign in
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
