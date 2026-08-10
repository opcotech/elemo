import type { Meta, StoryObj } from "@storybook/react-vite";

import { UserAvatar } from "@/components/ui/user-avatar";

const meta: Meta<typeof UserAvatar> = {
  title: "UI/UserAvatar",
  component: UserAvatar,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Avatar with initials fallback and optional name and email display.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    firstName: "Jane",
    lastName: "Cooper",
  },
};

export const WithEmail: Story = {
  args: {
    firstName: "Jane",
    lastName: "Cooper",
    email: "jane.cooper@example.com",
    showEmail: true,
  },
};

export const WithPicture: Story = {
  args: {
    firstName: "Pedro",
    lastName: "Duarte",
    email: "pedro@example.com",
    showEmail: true,
    picture: "https://github.com/shadcn.png",
  },
};

export const Sizes: Story = {
  render: () => (
    <div className="space-y-4">
      <UserAvatar firstName="Small" lastName="User" size="sm" />
      <UserAvatar firstName="Medium" lastName="User" size="md" />
      <UserAvatar firstName="Large" lastName="User" size="lg" />
    </div>
  ),
};
