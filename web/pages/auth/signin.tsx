import Image from 'next/image';
import {getCsrfToken} from 'next-auth/react';
import Button from '@/components/Button';
import type {GetServerSidePropsContext, InferGetServerSidePropsType} from 'next';
import {useState} from 'react';
import type {FormEvent} from 'react';

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

const errors: Record<SignInErrorTypes, string> = {
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

export default function SignInPage({csrfToken, error: err}: InferGetServerSidePropsType<typeof getServerSideProps>) {
  const error = err && (errors[err] ?? errors.default);

  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    event.currentTarget.submit();
  }

  return (
    <div className="flex min-h-full">
      <div className="flex flex-1 flex-col justify-center px-4 py-12 sm:px-6 lg:flex-none lg:px-20 xl:px-24">
        <div className="mx-auto w-full max-w-sm lg:w-96">
          <div>
            <Image
              className="h-12 w-auto"
              src="https://tailwindui.com/img/logos/mark.svg?color=blue&shade=600"
              alt="Elemo"
              width={56}
              height={56}
            />
            <h2 className="mt-6 text-3xl font-bold tracking-tight text-gray-900">Sign in to your account</h2>
            <p className="mt-2 text-gray-600">If you don&apos;t have an account, please contact your administrator.</p>
          </div>

          <div className="mt-8">
            {error && (
              <div className="rounded-md bg-red-50 p-4">
                <div className="">
                  <h3 className="text-base font-medium text-red-800">Failed to sign in!</h3>
                  <div className="mt-2 text-base text-red-700">
                    <p>{error}</p>
                  </div>
                </div>
              </div>
            )}
            <div className="mt-6">
              <form className="space-y-6" method="POST" action="/api/auth/callback/credentials" onSubmit={handleSubmit}>
                <input name="csrfToken" type="hidden" defaultValue={csrfToken}/>
                <div>
                  <label htmlFor="username" className="block font-medium leading-6 text-gray-900">
                    Email address
                  </label>
                  <div className="mt-2">
                    <input
                      id="username"
                      name="username"
                      type="email"
                      autoComplete="email"
                      defaultValue={'gabor@elemo.app'}
                      required
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
                      autoComplete="current-password"
                      defaultValue={'AppleTree123'}
                      required
                      className="block w-full rounded-md border-0 py-1.5 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:leading-6"
                      disabled={submitting}
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
                  <Button type="submit" loading={submitting} className="w-full">
                    Sign in
                  </Button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
      <div className="relative hidden w-0 flex-1 lg:block">
        <Image
          className="absolute inset-0 h-full w-full object-cover"
          src="https://images.unsplash.com/photo-1605106901227-991bd663255c?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=1335&q=80"
          alt=""
          fill
        />
      </div>
    </div>
  );
}

export async function getServerSideProps(context: GetServerSidePropsContext) {
  return {
    props: {
      error: (context.query?.error as SignInErrorTypes) || null,
      csrfToken: await getCsrfToken(context)
    }
  };
}
