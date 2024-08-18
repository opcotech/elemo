import * as React from 'react';
import type { Meta, StoryObj } from '@storybook/react';

import { CodeBlock } from '@/components/ui/code-block';
import { CopyButton } from '@/components/ui/copy-button';

const code = JSON.stringify({ title: 'success', message: 'sample code to copy' }, null, 2);

function Example(props: React.HTMLAttributes<HTMLPreElement>) {
  const ref = React.useRef(null);

  return (
    <CodeBlock {...props}>
      <CopyButton
        targetRef={ref}
        className='absolute right-1 top-1 text-white/70 hover:bg-muted-light/10 hover:text-white'
      />
      <span ref={ref}>{code}</span>
    </CodeBlock>
  );
}

const meta: Meta<typeof CodeBlock> = {
  title: 'Elements/Code Block',
  component: CodeBlock,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  render: (props) => {
    return <CodeBlock {...props}>{code}</CodeBlock>;
  },
};

export default meta;
type Story = StoryObj<typeof CodeBlock>;

export const Default: Story = {};

export const WithCopyButton: Story = {
  render: (props) => <Example {...props} />,
};
