import type {ReactNode} from 'react';

import {concat} from '@/helpers';

export interface PageProps {
  title?: string;
  className?: string;
  fullWidth?: boolean;
  children: ReactNode;
}

export default function Page({ title, className, fullWidth, children }: PageProps): JSX.Element {
  const containerClass = concat(className, fullWidth ? 'w-full' : 'mx-auto px-4 sm:px-6 lg:px-8');

  return (
    <div className={'py-10'}>
      {title && (
        <header className="mb-10">
          <div className={containerClass}>
            <h1>{title}</h1>
          </div>
        </header>
      )}

      <main className={containerClass}>{children}</main>
    </div>
  );
}
