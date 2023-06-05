'use client';

import { SignOutForm } from '@/components/auth/SignOutForm';

export default function SignOutPage() {
  return (
    <div className="h-screen w-screen flex items-center">
      <div className="max-w-xl mx-auto text-center">
        <h2 className="mb-4">Are you sure you want to sign out?</h2>
        <p className="mb-10">
          By signing out, your session will be terminated and you will be redirected to login page.
        </p>

        <SignOutForm />
      </div>
    </div>
  );
}
