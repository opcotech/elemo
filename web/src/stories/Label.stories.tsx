import type { Meta, StoryObj } from "@storybook/react-vite";
import { Asterisk, Info } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

const meta: Meta<typeof Label> = {
  title: "UI/Label",
  component: Label,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Renders an accessible label associated with controls. Built on top of Radix UI Label primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    htmlFor: {
      control: "text",
      description: "The id of the element the label is associated with",
    },
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the label",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic label
export const Default: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="email">Email</Label>
      <Input type="email" id="email" placeholder="Email" />
    </div>
  ),
};

// Required field label
export const Required: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="name" className="flex items-center gap-1">
        Name
        <Asterisk className="h-3 w-3 text-red-500" />
      </Label>
      <Input type="text" id="name" placeholder="Enter your name" />
    </div>
  ),
};

// Label with description
export const WithDescription: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="username">Username</Label>
      <Input type="text" id="username" placeholder="Username" />
      <p className="text-muted-foreground text-sm">
        This will be your public display name.
      </p>
    </div>
  ),
};

// Label with helper text
export const WithHelperText: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="password" className="flex items-center gap-2">
        Password
        <Info className="text-muted-foreground h-4 w-4" />
      </Label>
      <Input type="password" id="password" placeholder="Password" />
      <p className="text-muted-foreground text-xs">
        Must be at least 8 characters long and include a number.
      </p>
    </div>
  ),
};

// Checkbox with label
export const WithCheckbox: Story = {
  render: () => (
    <div className="flex items-center space-x-2">
      <Checkbox id="terms" />
      <Label
        htmlFor="terms"
        className="text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Accept terms and conditions
      </Label>
    </div>
  ),
};

// Textarea with label
export const WithTextarea: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="message">Your message</Label>
      <Textarea id="message" placeholder="Type your message here." />
    </div>
  ),
};

// Disabled label
export const Disabled: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="disabled-input" className="opacity-50">
        Disabled field
      </Label>
      <Input type="text" id="disabled-input" placeholder="Disabled" disabled />
    </div>
  ),
};

// Error state
export const ErrorState: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="error-input" className="text-red-600">
        Email
      </Label>
      <Input
        type="email"
        id="error-input"
        placeholder="Email"
        className="border-red-500 focus:border-red-500 focus:ring-red-500"
      />
      <p className="text-sm text-red-600">
        Please enter a valid email address.
      </p>
    </div>
  ),
};

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="small" className="text-xs">
          Small label
        </Label>
        <Input type="text" id="small" placeholder="Small input" />
      </div>

      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="default" className="text-sm">
          Default label
        </Label>
        <Input type="text" id="default" placeholder="Default input" />
      </div>

      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="large" className="text-base">
          Large label
        </Label>
        <Input type="text" id="large" placeholder="Large input" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Labels in different sizes to match various input controls.",
      },
    },
  },
};

// Complex form layout
export const FormLayout: Story = {
  render: () => (
    <div className="max-w-md space-y-6">
      <div className="grid gap-2">
        <Label htmlFor="first-name" className="flex items-center gap-1">
          First Name
          <Asterisk className="h-3 w-3 text-red-500" />
        </Label>
        <Input type="text" id="first-name" placeholder="John" />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="last-name" className="flex items-center gap-1">
          Last Name
          <Asterisk className="h-3 w-3 text-red-500" />
        </Label>
        <Input type="text" id="last-name" placeholder="Doe" />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="email-form">Email</Label>
        <Input type="email" id="email-form" placeholder="john@example.com" />
        <p className="text-muted-foreground text-xs">
          We'll never share your email with anyone else.
        </p>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="bio">Bio</Label>
        <Textarea id="bio" placeholder="Tell us about yourself" rows={3} />
        <p className="text-muted-foreground text-xs">Maximum 500 characters.</p>
      </div>

      <div className="flex items-center space-x-2">
        <Checkbox id="newsletter" />
        <Label htmlFor="newsletter" className="text-sm">
          Subscribe to our newsletter
        </Label>
      </div>

      <div className="flex items-center space-x-2">
        <Checkbox id="privacy" />
        <Label htmlFor="privacy" className="flex items-center gap-1 text-sm">
          I agree to the privacy policy
          <Asterisk className="h-3 w-3 text-red-500" />
        </Label>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A complex form layout showcasing various label patterns and use cases.",
      },
    },
  },
};

// Label without htmlFor (accessibility warning)
export const WithoutHtmlFor: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label className="rounded bg-amber-600 px-2 py-1 text-white">
        Label without htmlFor (not recommended)
      </Label>
      <Input type="text" placeholder="Input without proper association" />
      <p className="rounded bg-amber-600 px-2 py-1 text-xs text-white">
        ⚠️ This label is not properly associated with the input
      </p>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Example showing a label without htmlFor attribute - not recommended for accessibility.",
      },
    },
  },
};

// Custom styling
export const CustomStyling: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label
          htmlFor="custom1"
          className="rounded bg-blue-600 px-2 py-1 font-semibold tracking-wide text-white uppercase"
        >
          Custom Label
        </Label>
        <Input type="text" id="custom1" placeholder="Input" />
      </div>

      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label
          htmlFor="custom2"
          className="rounded-md bg-green-600 px-2 py-1 text-sm text-white"
        >
          Custom Success
        </Label>
        <Input type="text" id="custom2" placeholder="Input" />
      </div>

      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label
          htmlFor="custom3"
          className="border-l-4 border-purple-600 pl-3 font-medium text-purple-600"
        >
          Bordered Label
        </Label>
        <Input type="text" id="custom3" placeholder="Input" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Labels with custom styling to demonstrate design flexibility.",
      },
    },
  },
};
