import type { Meta, StoryObj } from '@storybook/react';

import { Drawer } from './Drawer';
import { useArgs } from '@storybook/client-api';
import { Button } from './Button';

const meta: Meta<typeof Drawer> = {
  title: 'Elements/Drawer',
  component: Drawer,
  tags: ['autodocs'],
  args: {
    id: 'default',
    title: 'Drawer Title',
    wide: false
  }
};

export default meta;
type Story = StoryObj<typeof Drawer>;

export const Default = (args: Story['args']) => {
  const [{ show }, updateArgs] = useArgs();
  const toggleShow = () => updateArgs({ show: !show });

  return (
    <div>
      <Button onClick={() => updateArgs({ show: !show })}>Open Drawer</Button>
      <Drawer id={args?.id!} title={args?.title!} wide={args?.wide!} show={show} toggle={toggleShow}>
        <p className="text-sm text-gray-500">
          Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed neque velit, lobortis ut magna varius, blandit
          rhoncus sem. Morbi lacinia nisi ac dui fermentum, sed luctus urna tincidunt. Etiam ut feugiat ex. Cras non
          risus mi.
        </p>
      </Drawer>
    </div>
  );
};
