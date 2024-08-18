import * as React from 'react';
import type { Meta, StoryObj } from '@storybook/react';

import { CopyButton, type CopyButtonProps } from '@/components/ui/copy-button';

function Example(props: Omit<CopyButtonProps, 'targetRef'>) {
  const ref = React.useRef(null);

  return (
    <div className='flex space-x-1'>
      <p ref={ref}>Copy this text</p>
      <CopyButton targetRef={ref} {...props} />
    </div>
  );
}

const meta: Meta<typeof CopyButton> = {
  title: 'Elements/Copy Button',
  component: CopyButton,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return <Example {...props} />;
  },
};

export default meta;
type Story = StoryObj<typeof CopyButton>;

export const Default: Story = {};
