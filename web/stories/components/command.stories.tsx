import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';

import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command';

function Commands() {
  return (
    <>
      <CommandInput placeholder='Type a command or search...' />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading='Issues'>
          <CommandItem>
            <span>New issue</span>
          </CommandItem>
          <CommandItem>
            <span>Recent issues</span>
          </CommandItem>
          <CommandItem>
            <span>Search issue</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
      </CommandList>
    </>
  );
}

function CommandMenu() {
  const [open, setOpen] = React.useState(false);

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'c') {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };
    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  return (
    <>
      <p className='text-sm text-muted-foreground'>
        Press{' '}
        <kbd className='font-mono pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 text-[10px] font-medium uppercase text-muted-foreground opacity-100'>
          <span className='text-xs'>⌘</span>c
        </kbd>
      </p>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <Commands />
      </CommandDialog>
    </>
  );
}

const meta: Meta<typeof Command> = {
  title: 'Components/Command',
  component: Command,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: () => (
    <Command className='rounded-lg border shadow-md'>
      <Commands />
    </Command>
  ),
};

export default meta;
type Story = StoryObj<typeof Command>;

export const Default: Story = {};

export const DialogCommand: Story = {
  render: (props) => <CommandMenu {...props} />,
};
