'use client';

import { useEffect } from 'react';

import Button from '@/components/Button';
import Page from '@/components/Page';

export default function Error({ error, reset }: { error: Error; reset: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <Page className="text-center">
      <h2>Something went wrong!</h2>
      <h3 className="mb-3">
        <>
          {error.message} - {error.cause}
        </>
      </h3>
      <pre className="text-left mb-6">{error.stack}</pre>
      <Button onClick={reset}>Try again</Button>
    </Page>
  );
}
