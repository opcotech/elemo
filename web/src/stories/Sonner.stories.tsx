import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Bell,
  Calendar,
  Download,
  Heart,
  Mail,
  Settings,
  Share,
  Star,
  Trash2,
  Upload,
  User,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Toaster } from "@/components/ui/sonner";

const meta: Meta<typeof Toaster> = {
  title: "UI/Sonner",
  component: Toaster,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "An opinionated toast component for React. Provides beautiful, customizable toast notifications with animations and positioning options.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    position: {
      control: "select",
      options: [
        "top-left",
        "top-center",
        "top-right",
        "bottom-left",
        "bottom-center",
        "bottom-right",
      ],
      description: "Position of the toast notifications",
    },
    richColors: {
      control: "boolean",
      description: "Whether to use rich colors for different toast types",
    },
    closeButton: {
      control: "boolean",
      description: "Whether to show close button on toasts",
    },
    duration: {
      control: "number",
      description: "Default duration for toasts in milliseconds",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic toast types
export const Default: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("Event has been created", {
              description: "Sunday, December 03, 2023 at 9:00 AM",
            })
          }
        >
          Default Toast
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.success("Payment successful", {
              description: "Your payment has been processed successfully.",
            })
          }
        >
          Success Toast
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.error("Something went wrong", {
              description: "There was a problem with your request.",
            })
          }
        >
          Error Toast
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.warning("Storage almost full", {
              description: "You're running out of storage space.",
            })
          }
        >
          Warning Toast
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.info("New feature available", {
              description: "Check out the new dashboard improvements.",
            })
          }
        >
          Info Toast
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};

// With actions
export const WithActions: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("Friend request sent", {
              description: "Your friend request has been sent to Alex.",
              action: {
                label: "Undo",
                onClick: () => toast("Friend request cancelled"),
              },
            })
          }
        >
          With Action
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Email sent successfully", {
              description: "Your message has been delivered to the recipient.",
              action: {
                label: "View",
                onClick: () => toast("Opening email..."),
              },
            })
          }
        >
          With View Action
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.error("Failed to delete item", {
              description: "The item could not be deleted at this time.",
              action: {
                label: "Retry",
                onClick: () => toast("Retrying deletion..."),
              },
            })
          }
        >
          Error with Retry
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};

// Custom icons and styling
export const CustomStyling: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("Message received", {
              description: "You have a new message from Sarah",
              icon: <Mail className="h-4 w-4" />,
            })
          }
        >
          <Mail className="mr-2 h-4 w-4" />
          Message
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Download complete", {
              description: "Your file has been downloaded successfully",
              icon: <Download className="h-4 w-4" />,
            })
          }
        >
          <Download className="mr-2 h-4 w-4" />
          Download
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Upload started", {
              description: "Your files are being uploaded",
              icon: <Upload className="h-4 w-4" />,
            })
          }
        >
          <Upload className="mr-2 h-4 w-4" />
          Upload
        </Button>
        <Button
          variant="destructive"
          onClick={() =>
            toast("Item deleted", {
              description: "The item has been moved to trash",
              icon: <Trash2 className="h-4 w-4" />,
              action: {
                label: "Undo",
                onClick: () => toast("Item restored"),
              },
            })
          }
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};

// Loading and promise toasts
export const LoadingToasts: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() => {
            toast.loading("Saving changes...", {
              description: "Please wait while we save your data",
            });
            setTimeout(() => {
              toast.success("Changes saved", {
                description: "Your data has been saved successfully",
              });
            }, 2000);
          }}
        >
          Save with Loading
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            const promise = new Promise((resolve) => {
              setTimeout(resolve, 3000);
            });

            toast.promise(promise, {
              loading: "Uploading file...",
              success: "File uploaded successfully",
              error: "Failed to upload file",
            });
          }}
        >
          Promise Toast
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            const toastId = toast.loading("Processing...", {
              description: "This might take a few seconds",
            });

            setTimeout(() => {
              toast.success("Processing complete", {
                id: toastId,
                description: "Your request has been processed",
              });
            }, 3000);
          }}
        >
          Update Toast
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};

// Rich colors
export const RichColors: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast.success("Account created", {
              description: "Welcome to our platform!",
            })
          }
        >
          Success
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.error("Connection failed", {
              description: "Unable to connect to the server",
            })
          }
        >
          Error
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.warning("Session expiring", {
              description: "Your session will expire in 5 minutes",
            })
          }
        >
          Warning
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.info("Update available", {
              description: "A new version is available for download",
            })
          }
        >
          Info
        </Button>
      </div>
      <Toaster richColors />
    </div>
  ),
};

// Custom duration
export const CustomDuration: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("Quick notification", {
              description: "This toast disappears quickly",
              duration: 1000,
            })
          }
        >
          1 Second
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Standard notification", {
              description: "This toast lasts 4 seconds",
              duration: 4000,
            })
          }
        >
          4 Seconds
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Long notification", {
              description: "This toast stays for 10 seconds",
              duration: 10000,
            })
          }
        >
          10 Seconds
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Persistent notification", {
              description: "This toast won't disappear automatically",
              duration: Infinity,
            })
          }
        >
          Persistent
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};

// Position variants
export const Positioning: Story = {
  render: () => {
    const [position, setPosition] = useState<
      | "top-left"
      | "top-center"
      | "top-right"
      | "bottom-left"
      | "bottom-center"
      | "bottom-right"
    >("bottom-right");

    const positions = [
      "top-left",
      "top-center",
      "top-right",
      "bottom-left",
      "bottom-center",
      "bottom-right",
    ] as const;

    return (
      <div className="space-y-4">
        <div className="flex flex-wrap gap-2">
          {positions.map((pos) => (
            <Button
              key={pos}
              variant={position === pos ? "default" : "outline"}
              size="sm"
              onClick={() => {
                setPosition(pos);
                toast(`Toast position: ${pos}`, {
                  description: `Notifications will appear in ${pos} corner`,
                });
              }}
            >
              {pos}
            </Button>
          ))}
        </div>
        <div className="text-center">
          <Button
            onClick={() =>
              toast("Test notification", {
                description: `Positioned at ${position}`,
              })
            }
          >
            Show Toast
          </Button>
        </div>
        <Toaster position={position} />
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "Toast notifications can be positioned in different corners of the screen.",
      },
    },
  },
};

// Application notifications
export const ApplicationNotifications: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("New comment", {
              description: "John commented on your post",
              icon: <Bell className="h-4 w-4" />,
              action: {
                label: "View",
                onClick: () => toast("Opening comment..."),
              },
            })
          }
        >
          <Bell className="mr-2 h-4 w-4" />
          Notification
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Profile updated", {
              description: "Your profile information has been saved",
              icon: <User className="h-4 w-4" />,
            })
          }
        >
          <User className="mr-2 h-4 w-4" />
          Profile
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Event reminder", {
              description: "Meeting starts in 15 minutes",
              icon: <Calendar className="h-4 w-4" />,
              action: {
                label: "Join",
                onClick: () => toast("Joining meeting..."),
              },
            })
          }
        >
          <Calendar className="mr-2 h-4 w-4" />
          Reminder
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Settings saved", {
              description: "Your preferences have been updated",
              icon: <Settings className="h-4 w-4" />,
            })
          }
        >
          <Settings className="mr-2 h-4 w-4" />
          Settings
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};

// Social interactions
export const SocialInteractions: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("Post liked", {
              description: "You liked Sarah's photo",
              icon: <Heart className="h-4 w-4 text-red-500" />,
              action: {
                label: "Unlike",
                onClick: () => toast("Post unliked"),
              },
            })
          }
        >
          <Heart className="mr-2 h-4 w-4" />
          Like
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Post shared", {
              description: "You shared Alex's post to your timeline",
              icon: <Share className="h-4 w-4" />,
            })
          }
        >
          <Share className="mr-2 h-4 w-4" />
          Share
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast("Added to favorites", {
              description: "This item has been saved to your favorites",
              icon: <Star className="h-4 w-4 text-yellow-500" />,
              action: {
                label: "View",
                onClick: () => toast("Opening favorites..."),
              },
            })
          }
        >
          <Star className="mr-2 h-4 w-4" />
          Favorite
        </Button>
      </div>
      <Toaster />
    </div>
  ),
};

// Complex notifications
export const ComplexNotifications: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("System backup completed", {
              description:
                "All your data has been safely backed up to the cloud. Next backup scheduled for tomorrow at 3 AM.",
              duration: 6000,
              action: {
                label: "Details",
                onClick: () => toast("Opening backup details..."),
              },
            })
          }
        >
          Backup Complete
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.warning("Storage limit reached", {
              description:
                "You've used 95% of your storage quota. Upgrade your plan to continue uploading files.",
              duration: 8000,
              action: {
                label: "Upgrade",
                onClick: () => toast("Redirecting to upgrade page..."),
              },
            })
          }
        >
          Storage Warning
        </Button>
        <Button
          variant="destructive"
          onClick={() =>
            toast.error("Security alert", {
              description:
                "Unusual login activity detected from a new device. If this wasn't you, please secure your account immediately.",
              duration: 10000,
              action: {
                label: "Secure Account",
                onClick: () => toast("Opening security settings..."),
              },
            })
          }
        >
          Security Alert
        </Button>
      </div>
      <Toaster />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Complex notifications with longer descriptions and important actions.",
      },
    },
  },
};

// With close button
export const WithCloseButton: Story = {
  render: () => (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() =>
            toast("Important announcement", {
              description: "Please read the new terms of service",
              duration: Infinity,
            })
          }
        >
          Persistent Toast
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.success("Task completed", {
              description: "Your workflow has been executed successfully",
            })
          }
        >
          Success Toast
        </Button>
        <Button
          variant="outline"
          onClick={() =>
            toast.error("Operation failed", {
              description: "Unable to complete the requested operation",
            })
          }
        >
          Error Toast
        </Button>
      </div>
      <Toaster closeButton />
    </div>
  ),
};
