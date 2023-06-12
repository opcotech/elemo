import type { Meta, StoryObj } from '@storybook/react';
import { useArgs } from '@storybook/client-api';

import { Modal } from './Modal';
import { Button } from './Button';

const meta: Meta<typeof Modal> = {
  title: 'Elements/Modal',
  component: Modal,
  tags: ['autodocs'],
  args: {
    open: false,
    title: 'Modal Title',
    size: 'md'
  }
};

export default meta;
type Story = StoryObj<typeof Modal>;

export const Default = (args: Story['args']) => {
  const [{ open }, updateArgs] = useArgs();
  const toggleOpenState = () => updateArgs({ open: !open });

  return (
    <div>
      <Button onClick={() => updateArgs({ open: !open })}>Open Modal</Button>
      <Modal
        open={open}
        setOpen={toggleOpenState}
        title={args?.title!}
        size={args?.size!}
        actions={<Button onClick={() => alert('Got it!')}>Got it!</Button>}
      >
        <p className="text-sm text-gray-500">
          Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed neque velit, lobortis ut magna varius, blandit
          rhoncus sem. Morbi lacinia nisi ac dui fermentum, sed luctus urna tincidunt. Etiam ut feugiat ex. Cras non
          risus mi.
        </p>
      </Modal>
    </div>
  );
};
