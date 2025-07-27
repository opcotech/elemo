import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Apple,
  Banana,
  Cherry,
  Globe,
  Laptop,
  Monitor,
  Moon,
  Settings,
  Sun,
  Users,
} from "lucide-react";
import { useState } from "react";

import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const meta: Meta<typeof Select> = {
  title: "UI/Select",
  component: Select,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A dropdown select component with customizable options and styling. Built on top of Radix UI Select primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    value: {
      control: "text",
      description: "The controlled value of the select",
    },
    defaultValue: {
      control: "text",
      description: "The default value when uncontrolled",
    },
    onValueChange: {
      action: "onValueChange",
      description: "Callback fired when the value changes",
    },
    disabled: {
      control: "boolean",
      description: "Whether the select is disabled",
    },
    name: {
      control: "text",
      description: "The name of the select for form submission",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic select
export const Default: Story = {
  render: () => (
    <Select>
      <SelectTrigger className="w-[180px]">
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="apple">Apple</SelectItem>
        <SelectItem value="banana">Banana</SelectItem>
        <SelectItem value="blueberry">Blueberry</SelectItem>
        <SelectItem value="grapes">Grapes</SelectItem>
        <SelectItem value="pineapple">Pineapple</SelectItem>
      </SelectContent>
    </Select>
  ),
};

// With icons
export const WithIcons: Story = {
  render: () => (
    <Select>
      <SelectTrigger className="w-[200px]">
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="apple">
          <div className="flex items-center gap-2">
            <Apple className="h-4 w-4" />
            Apple
          </div>
        </SelectItem>
        <SelectItem value="banana">
          <div className="flex items-center gap-2">
            <Banana className="h-4 w-4" />
            Banana
          </div>
        </SelectItem>
        <SelectItem value="cherry">
          <div className="flex items-center gap-2">
            <Cherry className="h-4 w-4" />
            Cherry
          </div>
        </SelectItem>
      </SelectContent>
    </Select>
  ),
};

// With label
export const WithLabel: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="fruit-select">Favorite Fruit</Label>
      <Select>
        <SelectTrigger id="fruit-select">
          <SelectValue placeholder="Select a fruit" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="apple">Apple</SelectItem>
          <SelectItem value="banana">Banana</SelectItem>
          <SelectItem value="orange">Orange</SelectItem>
          <SelectItem value="grape">Grape</SelectItem>
        </SelectContent>
      </Select>
    </div>
  ),
};

// Disabled select
export const Disabled: Story = {
  render: () => (
    <Select disabled>
      <SelectTrigger className="w-[180px]">
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="apple">Apple</SelectItem>
        <SelectItem value="banana">Banana</SelectItem>
        <SelectItem value="orange">Orange</SelectItem>
      </SelectContent>
    </Select>
  ),
};

// Disabled option
export const DisabledOption: Story = {
  render: () => (
    <Select>
      <SelectTrigger className="w-[180px]">
        <SelectValue placeholder="Select a fruit" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="apple">Apple</SelectItem>
        <SelectItem value="banana" disabled>
          Banana (Out of stock)
        </SelectItem>
        <SelectItem value="orange">Orange</SelectItem>
        <SelectItem value="grape">Grape</SelectItem>
      </SelectContent>
    </Select>
  ),
};

// Controlled select
export const Controlled: Story = {
  render: () => {
    const [value, setValue] = useState("");

    return (
      <div className="space-y-2">
        <Label>Selected: {value || "None"}</Label>
        <Select value={value} onValueChange={setValue}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Select a fruit" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="apple">Apple</SelectItem>
            <SelectItem value="banana">Banana</SelectItem>
            <SelectItem value="orange">Orange</SelectItem>
            <SelectItem value="grape">Grape</SelectItem>
          </SelectContent>
        </Select>
      </div>
    );
  },
};

// Theme selector
export const ThemeSelector: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label>Theme</Label>
      <Select defaultValue="system">
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="light">
            <div className="flex items-center gap-2">
              <Sun className="h-4 w-4" />
              Light
            </div>
          </SelectItem>
          <SelectItem value="dark">
            <div className="flex items-center gap-2">
              <Moon className="h-4 w-4" />
              Dark
            </div>
          </SelectItem>
          <SelectItem value="system">
            <div className="flex items-center gap-2">
              <Laptop className="h-4 w-4" />
              System
            </div>
          </SelectItem>
        </SelectContent>
      </Select>
    </div>
  ),
};

// Country selector
export const CountrySelector: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label>Country</Label>
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select a country" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="us">🇺🇸 United States</SelectItem>
          <SelectItem value="ca">🇨🇦 Canada</SelectItem>
          <SelectItem value="mx">🇲🇽 Mexico</SelectItem>
          <SelectItem value="gb">🇬🇧 United Kingdom</SelectItem>
          <SelectItem value="fr">🇫🇷 France</SelectItem>
          <SelectItem value="de">🇩🇪 Germany</SelectItem>
          <SelectItem value="it">🇮🇹 Italy</SelectItem>
          <SelectItem value="es">🇪🇸 Spain</SelectItem>
          <SelectItem value="jp">🇯🇵 Japan</SelectItem>
          <SelectItem value="kr">🇰🇷 South Korea</SelectItem>
          <SelectItem value="cn">🇨🇳 China</SelectItem>
          <SelectItem value="in">🇮🇳 India</SelectItem>
          <SelectItem value="au">🇦🇺 Australia</SelectItem>
          <SelectItem value="br">🇧🇷 Brazil</SelectItem>
        </SelectContent>
      </Select>
    </div>
  ),
};

// Settings selector
export const SettingsSelector: Story = {
  render: () => (
    <div className="w-full max-w-sm space-y-4">
      <div className="grid gap-1.5">
        <Label>Language</Label>
        <Select defaultValue="en">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="en">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                English
              </div>
            </SelectItem>
            <SelectItem value="es">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                Español
              </div>
            </SelectItem>
            <SelectItem value="fr">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                Français
              </div>
            </SelectItem>
            <SelectItem value="de">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                Deutsch
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-1.5">
        <Label>Display Mode</Label>
        <Select defaultValue="auto">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="auto">
              <div className="flex items-center gap-2">
                <Monitor className="h-4 w-4" />
                Auto
              </div>
            </SelectItem>
            <SelectItem value="compact">
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                Compact
              </div>
            </SelectItem>
            <SelectItem value="comfortable">
              <div className="flex items-center gap-2">
                <Settings className="h-4 w-4" />
                Comfortable
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-1.5">
        <Label>Color Scheme</Label>
        <Select defaultValue="blue">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="blue">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-blue-500"></div>
                Blue
              </div>
            </SelectItem>
            <SelectItem value="green">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-green-500"></div>
                Green
              </div>
            </SelectItem>
            <SelectItem value="purple">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-purple-500"></div>
                Purple
              </div>
            </SelectItem>
            <SelectItem value="red">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-red-500"></div>
                Red
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Multiple select components for various settings configurations.",
      },
    },
  },
};

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="grid gap-1.5">
        <Label>Small</Label>
        <Select>
          <SelectTrigger className="h-8 w-[150px] text-xs">
            <SelectValue placeholder="Small select" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
            <SelectItem value="option3">Option 3</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-1.5">
        <Label>Default</Label>
        <Select>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Default select" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
            <SelectItem value="option3">Option 3</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-1.5">
        <Label>Large</Label>
        <Select>
          <SelectTrigger className="h-12 w-[220px] text-base">
            <SelectValue placeholder="Large select" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
            <SelectItem value="option3">Option 3</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Select components in different sizes.",
      },
    },
  },
};

// With descriptions
export const WithDescriptions: Story = {
  render: () => (
    <Select>
      <SelectTrigger className="w-[280px]">
        <SelectValue placeholder="Choose a plan" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="free">
          <div className="flex flex-col items-start">
            <div className="font-medium">Free Plan</div>
            <div className="text-muted-foreground text-xs">
              Perfect for getting started
            </div>
          </div>
        </SelectItem>
        <SelectItem value="pro">
          <div className="flex flex-col items-start">
            <div className="font-medium">Pro Plan</div>
            <div className="text-muted-foreground text-xs">
              For growing businesses
            </div>
          </div>
        </SelectItem>
        <SelectItem value="enterprise">
          <div className="flex flex-col items-start">
            <div className="font-medium">Enterprise Plan</div>
            <div className="text-muted-foreground text-xs">
              Advanced features and support
            </div>
          </div>
        </SelectItem>
      </SelectContent>
    </Select>
  ),
};

// Form integration
export const FormIntegration: Story = {
  render: () => (
    <form className="w-full max-w-sm space-y-4">
      <div className="grid gap-1.5">
        <Label htmlFor="category">Category *</Label>
        <Select name="category" required>
          <SelectTrigger id="category">
            <SelectValue placeholder="Select category" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="electronics">Electronics</SelectItem>
            <SelectItem value="clothing">Clothing</SelectItem>
            <SelectItem value="books">Books</SelectItem>
            <SelectItem value="home">Home & Garden</SelectItem>
            <SelectItem value="sports">Sports</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="priority">Priority</Label>
        <Select name="priority" defaultValue="medium">
          <SelectTrigger id="priority">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="low">
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-green-500"></div>
                Low
              </div>
            </SelectItem>
            <SelectItem value="medium">
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-yellow-500"></div>
                Medium
              </div>
            </SelectItem>
            <SelectItem value="high">
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-red-500"></div>
                High
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="status">Status</Label>
        <Select name="status" defaultValue="draft">
          <SelectTrigger id="status">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="draft">Draft</SelectItem>
            <SelectItem value="review">In Review</SelectItem>
            <SelectItem value="approved">Approved</SelectItem>
            <SelectItem value="published">Published</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </form>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Select components integrated within a form with proper labels and validation.",
      },
    },
  },
};
