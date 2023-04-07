'use client';

import {signIn, signOut, useSession} from 'next-auth/react';

export default function Home() {
  const {data: session} = useSession();

  return (
    <main>
      <h1 className="text-3xl font-bold underline">Hello world!</h1>
      {session ? (
        <>
          <p className="text-xl">Welcome {session.user?.email}!</p>
          <button onClick={() => signOut()}>Sign out</button>
        </>
      ) : (
        <>
          Not signed in <br/>
          <button onClick={() => signIn()}>Sign in</button>
        </>
      )}
    </main>
  );
}
