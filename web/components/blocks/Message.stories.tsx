import type { Meta, StoryObj } from '@storybook/react';

import { Message } from './Message';

const meta: Meta<typeof Message> = {
  title: 'Elements/Message',
  component: Message,
  tags: ['autodocs']
};

export default meta;
type Story = StoryObj<typeof Message>;

export const Sample: Story = {
  args: {
    type: 'info',
    title: 'Message Title',
    message: 'Message content',
    dismissAfter: null
  }
};

export const Info: Story = {
  args: {
    type: 'info',
    title: 'Info',
    message: 'This is an info message.',
    dismissAfter: null
  }
};

export const Success: Story = {
  args: {
    type: 'success',
    title: 'Success',
    message: 'This is a success message.',
    dismissAfter: null
  }
};

export const Warning: Story = {
  args: {
    type: 'warning',
    title: 'Warning',
    message: 'This is a warning message.',
    dismissAfter: null
  }
};

export const Error: Story = {
  args: {
    type: 'error',
    title: 'Error',
    message: 'This is an error message.',
    dismissAfter: null
  }
};
