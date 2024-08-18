import type { Meta, StoryObj } from '@storybook/react';

import { Calendar } from '@/components/ui/calendar';

const today = new Date('2024-04-01');
const tomorrow = new Date('2024-04-02');
const range: Date[] = [new Date('2024-04-29'), new Date('2024-04-30'), new Date('2024-05-01')];

const meta: Meta<typeof Calendar> = {
  title: 'Elements/Calendar',
  component: Calendar,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  args: {
    mode: 'single',
    className: 'rounded-md border',
    today: today,
    selected: tomorrow,
  },
};

export default meta;
type Story = StoryObj<typeof Calendar>;

export const Default: Story = {};

export const Multiple: Story = {
  args: {
    mode: 'multiple',
    selected: range,
    numberOfMonths: 2,
  },
};
