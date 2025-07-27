import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  AlertTriangle,
  CheckCircle,
  Info,
  Lightbulb,
  Terminal,
  XCircle,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

const meta: Meta<typeof Alert> = {
  title: "UI/Alert",
  component: Alert,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A flexible alert component for displaying important messages with optional icons, titles, and descriptions.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "destructive"],
      description: "The visual style variant of the alert",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic variants
export const Default: Story = {
  render: () => (
    <Alert className="w-[400px]">
      <Info className="h-4 w-4" />
      <AlertTitle>Heads up!</AlertTitle>
      <AlertDescription>
        You can add components to your app using the cli.
      </AlertDescription>
    </Alert>
  ),
};

export const Destructive: Story = {
  render: () => (
    <Alert variant="destructive" className="w-[400px]">
      <XCircle className="h-4 w-4" />
      <AlertTitle>Error</AlertTitle>
      <AlertDescription>
        Your session has expired. Please log in again.
      </AlertDescription>
    </Alert>
  ),
};

// Different message types
export const Success: Story = {
  render: () => (
    <Alert variant="success" className="w-[400px]">
      <CheckCircle className="h-4 w-4" />
      <AlertTitle>Success!</AlertTitle>
      <AlertDescription>
        Your changes have been saved successfully.
      </AlertDescription>
    </Alert>
  ),
};

export const Warning: Story = {
  render: () => (
    <Alert variant="warning" className="w-[400px]">
      <AlertTriangle className="h-4 w-4" />
      <AlertTitle>Warning</AlertTitle>
      <AlertDescription>
        This action cannot be undone. Please proceed with caution.
      </AlertDescription>
    </Alert>
  ),
};

export const InfoAlert: Story = {
  render: () => (
    <Alert variant="info" className="w-[400px]">
      <Info className="h-4 w-4" />
      <AlertTitle>Information</AlertTitle>
      <AlertDescription>
        We've updated our privacy policy. Please review the changes.
      </AlertDescription>
    </Alert>
  ),
};

// Without icon
export const WithoutIcon: Story = {
  render: () => (
    <Alert className="w-[400px]">
      <AlertTitle>System Maintenance</AlertTitle>
      <AlertDescription>
        Scheduled maintenance will occur on Sunday at 2:00 AM UTC.
      </AlertDescription>
    </Alert>
  ),
};

// Title only
export const TitleOnly: Story = {
  render: () => (
    <Alert className="w-[400px]">
      <Terminal className="h-4 w-4" />
      <AlertTitle>Command completed successfully</AlertTitle>
    </Alert>
  ),
};

// Description only
export const DescriptionOnly: Story = {
  render: () => (
    <Alert className="w-[400px]">
      <Lightbulb className="h-4 w-4" />
      <AlertDescription>
        Tip: You can use keyboard shortcuts to navigate faster.
      </AlertDescription>
    </Alert>
  ),
};

// Complex content
export const WithComplexContent: Story = {
  render: () => (
    <Alert className="w-[500px]">
      <Info className="h-4 w-4" />
      <AlertTitle>Update Available</AlertTitle>
      <AlertDescription>
        <p>A new version of the application is available.</p>
        <ul className="mt-2 list-inside list-disc space-y-1">
          <li>Improved performance</li>
          <li>Bug fixes</li>
          <li>New features</li>
        </ul>
        <p className="mt-2">
          <a href="#" className="underline hover:no-underline">
            View release notes
          </a>
        </p>
      </AlertDescription>
    </Alert>
  ),
};

// All variants showcase
export const AllVariants: Story = {
  render: () => (
    <div className="w-[500px] space-y-4">
      <Alert>
        <Info className="h-4 w-4" />
        <AlertTitle>Default Alert</AlertTitle>
        <AlertDescription>
          This is a default alert with informational styling.
        </AlertDescription>
      </Alert>

      <Alert variant="destructive">
        <XCircle className="h-4 w-4" />
        <AlertTitle>Destructive Alert</AlertTitle>
        <AlertDescription>
          This is a destructive alert indicating an error or dangerous action.
        </AlertDescription>
      </Alert>

      <Alert variant="success">
        <CheckCircle className="h-4 w-4" />
        <AlertTitle>Success Alert</AlertTitle>
        <AlertDescription>
          This is a success alert indicating a positive outcome.
        </AlertDescription>
      </Alert>

      <Alert variant="warning">
        <AlertTriangle className="h-4 w-4" />
        <AlertTitle>Warning Alert</AlertTitle>
        <AlertDescription>
          This is a warning alert indicating caution is needed.
        </AlertDescription>
      </Alert>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "All available alert variants and custom styled alerts displayed together.",
      },
    },
  },
};
