import Image from 'next/image';
import { SignInForm } from '@/components/auth/SignInForm';

export const metadata = {
  title: 'Sign in | Elemo'
};

export default function SignInPage() {
  return (
    <div className="flex min-h-full">
      <div className="flex flex-1 flex-col justify-center px-4 py-12 sm:px-6 lg:flex-none lg:px-20 xl:px-24">
        <div className="mx-auto w-full max-w-sm lg:w-96">
          <div>
            <Image
              className="h-12 w-auto"
              src="https://tailwindui.com/img/logos/mark.svg?color=blue&shade=600"
              alt="Elemo Logo"
              width={56}
              height={56}
            />
            <h2 className="mt-6 text-3xl font-bold tracking-tight text-gray-900">Sign in to your account</h2>
            <p className="mt-2 text-gray-600">If you don&apos;t have an account, please contact your administrator.</p>
          </div>

          <SignInForm />
        </div>
      </div>
      <div className="relative hidden w-0 flex-1 lg:block">
        <Image
          className="absolute inset-0 h-full w-full object-cover"
          src="https://images.unsplash.com/photo-1605106901227-991bd663255c?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=1335&q=80"
          alt="Login page welcome image"
          fill
        />
      </div>
    </div>
  );
}
