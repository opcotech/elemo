import type { Meta, StoryObj } from "@storybook/react-vite";

import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";

const meta: Meta<typeof Checkbox> = {
  title: "UI/Checkbox",
  component: Checkbox,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A checkbox component built on top of Radix UI with modern styling and accessibility features.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    checked: {
      control: "boolean",
      description: "Whether the checkbox is checked",
    },
    disabled: {
      control: "boolean",
      description: "Whether the checkbox is disabled",
    },
    "aria-invalid": {
      control: "boolean",
      description: "Whether the checkbox is in an invalid state",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic checkbox
export const Default: Story = {
  args: {},
};

export const Checked: Story = {
  args: {
    checked: true,
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};

export const DisabledChecked: Story = {
  args: {
    disabled: true,
    checked: true,
  },
};

export const Invalid: Story = {
  args: {
    "aria-invalid": true,
  },
};

// With labels
export const WithLabel: Story = {
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

export const WithDescription: Story = {
  render: () => (
    <div className="items-top flex space-x-2">
      <Checkbox id="terms2" />
      <div className="grid gap-1.5 leading-none">
        <Label
          htmlFor="terms2"
          className="text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Accept terms and conditions
        </Label>
        <p className="text-muted-foreground text-sm">
          You agree to our Terms of Service and Privacy Policy.
        </p>
      </div>
    </div>
  ),
};

// Form examples
export const FormExample: Story = {
  render: () => (
    <div className="w-[300px] space-y-4">
      <div className="flex items-center space-x-2">
        <Checkbox id="marketing" />
        <Label
          htmlFor="marketing"
          className="text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Marketing emails
        </Label>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox id="newsletter" defaultChecked />
        <Label
          htmlFor="newsletter"
          className="text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Newsletter subscription
        </Label>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox id="notifications" />
        <Label
          htmlFor="notifications"
          className="text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Push notifications
        </Label>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox id="security" defaultChecked disabled />
        <Label
          htmlFor="security"
          className="text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Security alerts (required)
        </Label>
      </div>
    </div>
  ),
};

// Checkbox list
export const CheckboxList: Story = {
  render: () => (
    <div className="w-[250px] space-y-3">
      <div className="text-sm font-medium">Select your interests:</div>
      <div className="space-y-2">
        <div className="flex items-center space-x-2">
          <Checkbox id="tech" />
          <Label htmlFor="tech" className="text-sm">
            Technology
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Checkbox id="design" defaultChecked />
          <Label htmlFor="design" className="text-sm">
            Design
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Checkbox id="business" />
          <Label htmlFor="business" className="text-sm">
            Business
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Checkbox id="science" defaultChecked />
          <Label htmlFor="science" className="text-sm">
            Science
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Checkbox id="arts" />
          <Label htmlFor="arts" className="text-sm">
            Arts & Culture
          </Label>
        </div>
      </div>
    </div>
  ),
};

// States showcase
export const AllStates: Story = {
  render: () => (
    <div className="w-[300px] space-y-4">
      <div className="flex items-center space-x-2">
        <Checkbox id="unchecked" />
        <Label htmlFor="unchecked" className="text-sm">
          Unchecked
        </Label>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox id="checked" checked />
        <Label htmlFor="checked" className="text-sm">
          Checked
        </Label>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox id="disabled-unchecked" disabled />
        <Label htmlFor="disabled-unchecked" className="text-sm">
          Disabled (Unchecked)
        </Label>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox id="disabled-checked" disabled checked />
        <Label htmlFor="disabled-checked" className="text-sm">
          Disabled (Checked)
        </Label>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox id="invalid" aria-invalid />
        <Label htmlFor="invalid" className="text-destructive text-sm">
          Invalid state
        </Label>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "All available checkbox states displayed together.",
      },
    },
  },
};

// Error state example
export const WithError: Story = {
  render: () => (
    <div className="w-[300px] space-y-2">
      <div className="flex items-center space-x-2">
        <Checkbox id="error-checkbox" aria-invalid />
        <Label htmlFor="error-checkbox" className="text-destructive text-sm">
          I agree to the terms
        </Label>
      </div>
      <p className="text-destructive text-sm">
        You must accept the terms and conditions to continue.
      </p>
    </div>
  ),
};
