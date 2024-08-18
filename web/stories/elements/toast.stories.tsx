import type { Meta, StoryObj } from '@storybook/react';

import {
  Toast,
  ToastAction,
  ToastActionElement,
  ToastClose,
  ToastDescription,
  ToastProps,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from '@/components/ui/toast';

interface ToastDemoProps {
  title: string;
  description: string;
  action: ToastActionElement;
  variant: ToastProps['variant'];
}

const meta: Meta<ToastDemoProps> = {
  title: 'Elements/Toast',
  tags: ['autodocs'],
  args: {
    title: 'To-do item added',
    description: 'A new item with the title "Make some toast" was added to your to-do list.',
  },
  render: ({ title, description, action, ...props }) => {
    return (
      <div className='py-10'>
        <ToastProvider>
          <Toast {...props} duration={Infinity}>
            <div className='grid gap-1'>
              {title && <ToastTitle>{title}</ToastTitle>}
              {description && <ToastDescription>{description}</ToastDescription>}
            </div>
            <div className='pr-1.5'>{action}</div>
            <ToastClose />
          </Toast>
          <ToastViewport />
        </ToastProvider>
      </div>
    );
  },
};

export default meta;
type Story = StoryObj<ToastDemoProps>;

export const Default: Story = {
  args: {
    title: 'To-do item added',
  },
};

export const WithAction: Story = {
  args: {
    title: 'To-do item added',
    action: <ToastAction altText='Update the newly created item'>Update</ToastAction>,
  },
};

export const Destructive: Story = {
  args: {
    title: 'Cannot load to-do list',
    description: 'Some errors happened when loading your to-do list. Please try again later.',
    action: <ToastAction altText='Try loading to-do list'>Try again</ToastAction>,
    variant: 'destructive',
  },
};
