'use client';

import * as React from 'react';
import type { Meta, StoryObj } from '@storybook/react';
import { IconCircleDot } from '@tabler/icons-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';

type Status = { value: string; label: string };

const statuses: Status[] = [
  { value: 'backlog', label: 'Backlog' },
  { value: 'todo', label: 'Todo' },
  { value: 'in progress', label: 'In Progress' },
  { value: 'done', label: 'Done' },
  { value: 'canceled', label: 'Canceled' },
];

function StatusList({
  setOpen,
  setSelectedStatus,
}: {
  setOpen: (open: boolean) => void;
  setSelectedStatus: (status: Status | null) => void;
}) {
  return (
    <Command>
      <CommandInput placeholder='Filter status...' />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup>
          {statuses.map((status) => (
            <CommandItem
              key={status.value}
              value={status.value}
              onSelect={(value) => {
                setSelectedStatus(statuses.find((priority) => priority.value === value) || null);
                setOpen(false);
              }}
            >
              {status.label}
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
}

function ComboBoxResponsive({ open }: { open: boolean }) {
  const [isOpen, setIsOpen] = React.useState(open);
  const [selectedStatus, setSelectedStatus] = React.useState<Status | null>(null);

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button
          variant='outline'
          className={cn('w-[200px] justify-start space-x-1', !selectedStatus && 'text-muted-foreground')}
        >
          {selectedStatus ? (
            <>{selectedStatus.label}</>
          ) : (
            <>
              <IconCircleDot className='h-4 w-4' />
              <span>set status</span>
            </>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className='w-[200px] p-0' align='start'>
        <StatusList setOpen={setIsOpen} setSelectedStatus={setSelectedStatus} />
      </PopoverContent>
    </Popover>
  );
}

const meta: Meta<typeof ComboBoxResponsive> = {
  title: 'Forms/Combobox',
  component: ComboBoxResponsive,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return <ComboBoxResponsive {...props} />;
  },
};

export default meta;
type Story = StoryObj<typeof ComboBoxResponsive>;

export const Default: Story = {};

export const Open: Story = {
  args: {
    open: true,
  },
};
