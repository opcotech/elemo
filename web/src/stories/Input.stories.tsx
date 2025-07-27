import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Calendar,
  CreditCard,
  Eye,
  EyeOff,
  Lock,
  Mail,
  Phone,
  Search,
  User,
} from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const meta: Meta<typeof Input> = {
  title: "UI/Input",
  component: Input,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A versatile input component with modern styling, support for various types, and excellent accessibility features.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    type: {
      control: "select",
      options: [
        "text",
        "email",
        "password",
        "number",
        "search",
        "tel",
        "url",
        "date",
        "time",
        "file",
      ],
      description: "The input type",
    },
    placeholder: {
      control: "text",
      description: "Placeholder text",
    },
    disabled: {
      control: "boolean",
      description: "Whether the input is disabled",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic input
export const Default: Story = {
  args: {
    placeholder: "Enter your text...",
  },
};

// Input types
export const Email: Story = {
  args: {
    type: "email",
    placeholder: "Enter your email...",
  },
};

export const Password: Story = {
  args: {
    type: "password",
    placeholder: "Enter your password...",
  },
};

export const Number: Story = {
  args: {
    type: "number",
    placeholder: "Enter a number...",
  },
};

export const SearchInput: Story = {
  args: {
    type: "search",
    placeholder: "Search...",
  },
};

export const Tel: Story = {
  args: {
    type: "tel",
    placeholder: "Enter phone number...",
  },
};

export const Date: Story = {
  args: {
    type: "date",
  },
};

// States
export const Disabled: Story = {
  args: {
    placeholder: "Disabled input...",
    disabled: true,
  },
};

export const WithValue: Story = {
  args: {
    value: "Input with value",
    readOnly: true,
  },
};

export const Invalid: Story = {
  args: {
    placeholder: "Invalid input...",
    "aria-invalid": true,
  },
};

// With labels
export const WithLabel: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="email">Email</Label>
      <Input type="email" id="email" placeholder="Email" />
    </div>
  ),
};

export const WithLabelAndDescription: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="username">Username</Label>
      <Input type="text" id="username" placeholder="Enter username" />
      <p className="text-muted-foreground text-sm">
        Your username must be at least 3 characters long.
      </p>
    </div>
  ),
};

// With icons
export const WithIconLeft: Story = {
  render: () => (
    <div className="relative w-full max-w-sm">
      <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
      <Input placeholder="Search..." className="pl-10" />
    </div>
  ),
};

export const WithIconRight: Story = {
  render: () => (
    <div className="relative w-full max-w-sm">
      <Input placeholder="Enter email..." type="email" className="pr-10" />
      <Mail className="text-muted-foreground absolute top-1/2 right-3 h-4 w-4 -translate-y-1/2" />
    </div>
  ),
};

// Password input with toggle
export const PasswordWithToggle: Story = {
  render: () => {
    const [showPassword, setShowPassword] = useState(false);

    return (
      <div className="relative w-full max-w-sm">
        <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
        <Input
          type={showPassword ? "text" : "password"}
          placeholder="Enter password..."
          className="pr-10 pl-10"
        />
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="absolute top-1/2 right-1 h-8 w-8 -translate-y-1/2"
          onClick={() => setShowPassword(!showPassword)}
        >
          {showPassword ? (
            <EyeOff className="h-4 w-4" />
          ) : (
            <Eye className="h-4 w-4" />
          )}
        </Button>
      </div>
    );
  },
};

// Form examples
export const LoginForm: Story = {
  render: () => (
    <div className="grid w-full max-w-sm gap-4">
      <div className="grid gap-2">
        <Label htmlFor="login-email">Email</Label>
        <div className="relative">
          <User className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <Input
            id="login-email"
            type="email"
            placeholder="Enter your email"
            className="pl-10"
          />
        </div>
      </div>
      <div className="grid gap-2">
        <Label htmlFor="login-password">Password</Label>
        <div className="relative">
          <Lock className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <Input
            id="login-password"
            type="password"
            placeholder="Enter your password"
            className="pl-10"
          />
        </div>
      </div>
      <Button className="w-full">Sign In</Button>
    </div>
  ),
};

export const ContactForm: Story = {
  render: () => (
    <div className="grid w-full max-w-md gap-4">
      <div className="grid gap-2">
        <Label htmlFor="contact-name">Full Name</Label>
        <Input id="contact-name" placeholder="John Doe" />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="contact-email">Email</Label>
        <div className="relative">
          <Mail className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <Input
            id="contact-email"
            type="email"
            placeholder="john@example.com"
            className="pl-10"
          />
        </div>
      </div>
      <div className="grid gap-2">
        <Label htmlFor="contact-phone">Phone</Label>
        <div className="relative">
          <Phone className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <Input
            id="contact-phone"
            type="tel"
            placeholder="+1 (555) 123-4567"
            className="pl-10"
          />
        </div>
      </div>
    </div>
  ),
};

// Specialized inputs
export const SearchWithIcon: Story = {
  render: () => (
    <div className="relative w-full max-w-md">
      <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
      <Input
        type="search"
        placeholder="Search products, customers, orders..."
        className="pr-4 pl-10"
      />
    </div>
  ),
};

export const CreditCardInput: Story = {
  render: () => (
    <div className="grid w-full max-w-sm gap-4">
      <div className="grid gap-2">
        <Label htmlFor="card-number">Card Number</Label>
        <div className="relative">
          <CreditCard className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <Input
            id="card-number"
            placeholder="1234 5678 9012 3456"
            className="pl-10"
            maxLength={19}
          />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="grid gap-2">
          <Label htmlFor="card-expiry">Expiry</Label>
          <Input id="card-expiry" placeholder="MM/YY" maxLength={5} />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="card-cvc">CVC</Label>
          <Input id="card-cvc" placeholder="123" maxLength={4} />
        </div>
      </div>
    </div>
  ),
};

export const DateTimeInputs: Story = {
  render: () => (
    <div className="grid w-full max-w-sm gap-4">
      <div className="grid gap-2">
        <Label htmlFor="date-input">Date</Label>
        <div className="relative">
          <Calendar className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <Input id="date-input" type="date" className="pl-10" />
        </div>
      </div>
      <div className="grid gap-2">
        <Label htmlFor="time-input">Time</Label>
        <Input id="time-input" type="time" />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="datetime-input">Date & Time</Label>
        <Input id="datetime-input" type="datetime-local" />
      </div>
    </div>
  ),
};

// File upload
export const FileUpload: Story = {
  render: () => (
    <div className="grid w-full max-w-sm gap-2">
      <Label htmlFor="file-upload">Upload File</Label>
      <Input id="file-upload" type="file" accept="image/*" />
      <p className="text-muted-foreground text-sm">
        Upload an image file (JPG, PNG, GIF)
      </p>
    </div>
  ),
};

// All input types showcase
export const AllInputTypes: Story = {
  render: () => (
    <div className="grid w-full max-w-2xl gap-4">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="grid gap-2">
          <Label>Text Input</Label>
          <Input placeholder="Text input" />
        </div>
        <div className="grid gap-2">
          <Label>Email Input</Label>
          <Input type="email" placeholder="email@example.com" />
        </div>
        <div className="grid gap-2">
          <Label>Password Input</Label>
          <Input type="password" placeholder="Password" />
        </div>
        <div className="grid gap-2">
          <Label>Number Input</Label>
          <Input type="number" placeholder="123" />
        </div>
        <div className="grid gap-2">
          <Label>Search Input</Label>
          <Input type="search" placeholder="Search..." />
        </div>
        <div className="grid gap-2">
          <Label>Tel Input</Label>
          <Input type="tel" placeholder="Phone number" />
        </div>
        <div className="grid gap-2">
          <Label>URL Input</Label>
          <Input type="url" placeholder="https://example.com" />
        </div>
        <div className="grid gap-2">
          <Label>Date Input</Label>
          <Input type="date" />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "All available input types displayed together.",
      },
    },
  },
};

// Error states
export const ErrorStates: Story = {
  render: () => (
    <div className="grid w-full max-w-sm gap-4">
      <div className="grid gap-2">
        <Label htmlFor="error-input" className="text-destructive">
          Email (with error)
        </Label>
        <Input
          id="error-input"
          type="email"
          placeholder="Enter email"
          aria-invalid={true}
          className="border-destructive focus-visible:ring-destructive/20"
        />
        <p className="text-destructive text-sm">
          Please enter a valid email address.
        </p>
      </div>
      <div className="grid gap-2">
        <Label htmlFor="success-input" className="text-success">
          Email (valid)
        </Label>
        <Input
          id="success-input"
          type="email"
          value="user@example.com"
          className="border-success focus-visible:ring-success/20"
          readOnly
        />
        <p className="text-success text-sm">Email address is valid.</p>
      </div>
    </div>
  ),
};
