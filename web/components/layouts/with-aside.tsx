'use client';

import * as React from 'react';
import { IconArrowLeft, IconArrowRight } from '@tabler/icons-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Header } from '@/components/header';
import { Shell, ShellAside, ShellContent } from '@/components/ui/shell';
import { ThemeToggle } from '@/components/theme/theme-toggle';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import Link from 'next/link';

const ASIDE_STATE_KEY = 'aside_state';

export type AsideItem = {
  label: string;
  href: string;
  icon: React.ReactNode;
};

export interface LayoutProps {
  asideItems: AsideItem[];
  children: React.ReactNode;
}

const Layout = ({ asideItems, children }: Readonly<LayoutProps>) => {
  const [asideOpen, setAsideOpen] = React.useState(false);

  React.useEffect(() => {
    setAsideOpen(localStorage.getItem(ASIDE_STATE_KEY) === 'true');
  }, []);

  function toggleAsideState() {
    const newState = !asideOpen;
    setAsideOpen(newState);
    localStorage.setItem(ASIDE_STATE_KEY, String(newState));
  }

  return (
    <Shell>
      <Header />
      <ShellAside size={asideOpen ? 'lg' : 'default'}>
        <TooltipProvider delayDuration={300}>
          {asideItems.map(({ icon, label, href }) => (
            <Tooltip key={href}>
              <TooltipTrigger asChild>
                <Button
                  asChild
                  variant='ghost'
                  size={asideOpen ? 'default' : 'icon'}
                  className={cn(
                    'w-full hover:bg-accent/10 hover:text-accent-foreground hover:dark:bg-accent/10 hover:dark:text-accent',
                    asideOpen ? 'justify-start' : null
                  )}
                >
                  <Link href={href}>
                    {icon}
                    {asideOpen && <span className='ml-2'>{label}</span>}
                  </Link>
                </Button>
              </TooltipTrigger>
              <TooltipContent side='right' className={asideOpen ? 'hidden' : ''}>
                {label}
              </TooltipContent>
            </Tooltip>
          ))}
        </TooltipProvider>
        <div className='mt-auto flex w-full flex-col space-y-1'>
          <ThemeToggle />
          <Button size='icon' variant='ghost' onClick={toggleAsideState} aria-label='toggle side panel width'>
            {asideOpen ? <IconArrowLeft className='h-4 w-4' /> : <IconArrowRight className='h-4 w-4' />}
          </Button>
        </div>
      </ShellAside>
      <div className={cn('transition-all duration-200 ease-out', asideOpen ? 'ml-64' : 'ml-16')}>
        <ShellContent>{children}</ShellContent>
      </div>
    </Shell>
  );
};
Layout.displayName = 'LayoutWithAside';

export { Layout };
