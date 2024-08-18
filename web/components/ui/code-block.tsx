'use client';

import * as React from 'react';

import { cn } from '@/lib/utils';

const CodeBlock = React.forwardRef<HTMLPreElement, React.HTMLAttributes<HTMLPreElement>>(
  ({ className, children, ...props }, ref) => (
    <pre className={cn('relative mt-2 w-full rounded-md bg-slate-800 p-4', className)} {...props}>
      <code ref={ref} className='text-white'>
        {children}
      </code>
    </pre>
  )
);
CodeBlock.displayName = 'CodeBlock';

export { CodeBlock };
