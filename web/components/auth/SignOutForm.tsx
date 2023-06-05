'use client';

import { FormEvent, useEffect, useState } from 'react';
import { getCsrfToken } from 'next-auth/react';
import { Button } from '@/components/blocks/Button';
import { Link } from '@/components/blocks/Link';

export function SignOutForm() {
  const [csrfToken, setCSRFToken] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    async function getCSRFToken() {
      const csrfToken = await getCsrfToken();
      setCSRFToken(csrfToken || '');
    }

    getCSRFToken();
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    event.currentTarget.submit();
  }

  return (
    <form className="justify-center" method="POST" action={'/api/auth/signout'} onSubmit={handleSubmit}>
      <input type="hidden" name="csrfToken" value={csrfToken} />
      <Button loading={submitting} type="submit">
        Sign out
      </Button>
      <span className="mx-2">or</span>
      <Link onClick={() => window.history.back()}>go back.</Link>
    </form>
  );
}
