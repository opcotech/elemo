'use client';

import * as React from 'react';
import { IconUserCircle } from '@tabler/icons-react';

import { Button } from '@/components/ui/button';
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from '@/components/ui/command';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Navbar,
  NavbarActionsContent,
  NavbarCollapsedContent,
  NavbarContent,
  NavbarItem,
} from '@/components/ui/navbar';

const Header = () => {
  const [commandDialogOpen, setCommandDialogOpen] = React.useState(false);

  function toggleCommandDialog() {
    setCommandDialogOpen((commandDialogOpen) => !commandDialogOpen);
  }

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        toggleCommandDialog();
      }
    };
    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  return (
    <Navbar>
      <NavbarContent>
        <NavbarItem href='#'>Dashboard</NavbarItem>
        <NavbarItem href='#'>Namespaces</NavbarItem>
        <NavbarItem href='#'>Projects</NavbarItem>
        <NavbarItem href='#'>Documents</NavbarItem>
      </NavbarContent>
      <NavbarCollapsedContent>
        <NavbarItem href='#'>Dashboard</NavbarItem>
        <NavbarItem href='#'>Namespaces</NavbarItem>
        <NavbarItem href='#'>Documents</NavbarItem>
      </NavbarCollapsedContent>
      <NavbarActionsContent>
        <Button
          variant='outline'
          className='min-w-[210px] justify-between text-xs text-muted-foreground hover:text-muted-foreground'
          onClick={toggleCommandDialog}
        >
          <span>Search or execute...</span>
          <kbd className='font-mono pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 text-[10px] font-medium uppercase text-muted-foreground opacity-100'>
            <span className='text-xs'>⌘</span>k
          </kbd>
        </Button>
        <CommandDialog open={commandDialogOpen} onOpenChange={setCommandDialogOpen}>
          <CommandInput placeholder='Type a command or search...' />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup heading='Quick actions'>
              <CommandItem>
                <span>New issue</span>
                <CommandShortcut>N+I</CommandShortcut>
              </CommandItem>
            </CommandGroup>
            <CommandSeparator />
          </CommandList>
        </CommandDialog>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant='secondary' size='icon' className='rounded-full'>
              <IconUserCircle stroke={1.5} className='h-5 w-5' />
              <span className='sr-only'>Toggle user menu</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end'>
            <DropdownMenuLabel>My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>Settings</DropdownMenuItem>
            <DropdownMenuItem>Support</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className='focus:bg-destructive/10 focus:text-destructive-dark focus:dark:bg-destructive-light/20 focus:dark:text-red-500'>
              Logout
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </NavbarActionsContent>
    </Navbar>
  );
};
Header.displayName = 'Header';

export { Header };
