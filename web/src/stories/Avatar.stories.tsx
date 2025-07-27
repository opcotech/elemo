import type { Meta, StoryObj } from "@storybook/react-vite";
import { User } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

const meta: Meta<typeof Avatar> = {
  title: "UI/Avatar",
  component: Avatar,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "An image element with a fallback for representing the user. Built on top of Radix UI Avatar primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the avatar",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic avatar with image
export const Default: Story = {
  render: () => (
    <Avatar>
      <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
      <AvatarFallback>CN</AvatarFallback>
    </Avatar>
  ),
};

// Avatar with fallback (broken image)
export const WithFallback: Story = {
  render: () => (
    <Avatar>
      <AvatarImage src="https://broken-link.jpg" alt="@user" />
      <AvatarFallback>JD</AvatarFallback>
    </Avatar>
  ),
};

// Fallback only
export const FallbackOnly: Story = {
  render: () => (
    <Avatar>
      <AvatarFallback>AB</AvatarFallback>
    </Avatar>
  ),
};

// Fallback with icon
export const FallbackWithIcon: Story = {
  render: () => (
    <Avatar>
      <AvatarFallback>
        <User className="h-4 w-4" />
      </AvatarFallback>
    </Avatar>
  ),
};

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Avatar className="h-6 w-6">
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback className="text-xs">XS</AvatarFallback>
      </Avatar>
      <Avatar className="h-8 w-8">
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback className="text-sm">SM</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback>MD</AvatarFallback>
      </Avatar>
      <Avatar className="h-12 w-12">
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback>LG</AvatarFallback>
      </Avatar>
      <Avatar className="h-16 w-16">
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback className="text-lg">XL</AvatarFallback>
      </Avatar>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Avatars in different sizes from extra small to extra large.",
      },
    },
  },
};

// Team avatars
export const Team: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Avatar>
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback>CN</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarImage src="https://github.com/vercel.png" alt="@vercel" />
        <AvatarFallback>VC</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback>AB</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback>
          <User className="h-4 w-4" />
        </AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback>+3</AvatarFallback>
      </Avatar>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "A collection of avatars representing a team or group.",
      },
    },
  },
};

// Stacked avatars
export const Stacked: Story = {
  render: () => (
    <div className="flex -space-x-2">
      <Avatar className="border-background border-2">
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback>CN</AvatarFallback>
      </Avatar>
      <Avatar className="border-background border-2">
        <AvatarImage src="https://github.com/vercel.png" alt="@vercel" />
        <AvatarFallback>VC</AvatarFallback>
      </Avatar>
      <Avatar className="border-background border-2">
        <AvatarFallback>AB</AvatarFallback>
      </Avatar>
      <Avatar className="border-background border-2">
        <AvatarFallback>CD</AvatarFallback>
      </Avatar>
      <Avatar className="border-background border-2">
        <AvatarFallback>+5</AvatarFallback>
      </Avatar>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Overlapping avatars with borders to show multiple users.",
      },
    },
  },
};

// Custom colors
export const CustomColors: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Avatar>
        <AvatarFallback className="bg-red-500 text-white">RD</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback className="bg-blue-500 text-white">BL</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback className="bg-green-500 text-white">GR</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback className="bg-purple-500 text-white">PR</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback className="bg-orange-500 text-white">OR</AvatarFallback>
      </Avatar>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Avatars with custom background colors for fallbacks.",
      },
    },
  },
};

// Status indicators
export const WithStatus: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <div className="relative">
        <Avatar>
          <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
          <AvatarFallback>CN</AvatarFallback>
        </Avatar>
        <div className="border-background absolute right-0 bottom-0 h-3 w-3 rounded-full border-2 bg-green-500"></div>
      </div>
      <div className="relative">
        <Avatar>
          <AvatarFallback>AB</AvatarFallback>
        </Avatar>
        <div className="border-background absolute right-0 bottom-0 h-3 w-3 rounded-full border-2 bg-yellow-500"></div>
      </div>
      <div className="relative">
        <Avatar>
          <AvatarFallback>CD</AvatarFallback>
        </Avatar>
        <div className="border-background absolute right-0 bottom-0 h-3 w-3 rounded-full border-2 bg-gray-400"></div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Avatars with status indicators showing online, away, and offline states.",
      },
    },
  },
};

// Square avatars
export const Square: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Avatar className="rounded-lg">
        <AvatarImage src="https://github.com/shadcn.png" alt="@shadcn" />
        <AvatarFallback className="rounded-lg">CN</AvatarFallback>
      </Avatar>
      <Avatar className="rounded-md">
        <AvatarFallback className="rounded-md">AB</AvatarFallback>
      </Avatar>
      <Avatar className="rounded-none">
        <AvatarFallback className="rounded-none">CD</AvatarFallback>
      </Avatar>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Avatars with different border radius values for non-circular shapes.",
      },
    },
  },
};
