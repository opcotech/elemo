import * as React from 'react';

import { cn } from '@/lib/utils';

export interface KbdProps extends React.HTMLAttributes<HTMLElement> {}

function Kbd({ className, ...props }: KbdProps) {
  return (
    <kbd
      className={cn(
        'font-mono pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 text-[10px] font-medium uppercase text-muted-foreground',
        className
      )}
      {...props}
    >
      <span className='text-xs'>⌘</span>c
    </kbd>
  );
}
Kbd.displayName = 'Kbd';

export { Kbd };
