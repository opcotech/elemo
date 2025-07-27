import type { Meta, StoryObj } from "@storybook/react-vite";

import { ThemeProvider } from "@/components/theme-provider";
import { ThemeSwitcher } from "@/components/theme-switcher";

const meta: Meta<typeof ThemeSwitcher> = {
  title: "Components/ThemeSwitcher",
  component: ThemeSwitcher,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A theme switcher component that allows users to toggle between light, dark, and system themes.",
      },
    },
  },
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <ThemeProvider>
        <Story />
      </ThemeProvider>
    ),
  ],
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "ghost", "outline"],
      description: "The visual variant of the button",
    },
    size: {
      control: "select",
      options: ["default", "sm", "lg", "icon"],
      description: "The size of the button",
    },
  },
};

export default meta;
type Story = StoryObj<typeof ThemeSwitcher>;

// Default theme switcher
export const Default: Story = {
  args: {},
};

// Different variants
export const Variants: Story = {
  render: () => (
    <div className="flex gap-4">
      <ThemeSwitcher variant="default" />
      <ThemeSwitcher variant="ghost" />
      <ThemeSwitcher variant="outline" />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Theme switcher in different button variants.",
      },
    },
  },
};

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <ThemeSwitcher size="sm" />
      <ThemeSwitcher size="default" />
      <ThemeSwitcher size="lg" />
      <ThemeSwitcher size="icon" />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Theme switcher in different sizes.",
      },
    },
  },
};

// In navigation context
export const InNavigation: Story = {
  render: () => (
    <div className="flex w-80 items-center justify-between rounded-lg border p-4">
      <div className="flex items-center gap-2">
        <div className="bg-primary h-8 w-8 rounded" />
        <span className="font-medium">Elemo</span>
      </div>
      <div className="flex items-center gap-2">
        <ThemeSwitcher />
        <div className="bg-muted h-8 w-8 rounded-full" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Theme switcher in a navigation bar context.",
      },
    },
  },
};
