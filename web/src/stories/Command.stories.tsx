import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Bell,
  Calculator,
  Calendar,
  ClipboardPaste,
  Copy,
  CreditCard,
  Download,
  Edit,
  File,
  Folder,
  Home,
  Image,
  Mail,
  Music,
  Phone,
  Scissors,
  Search,
  Settings,
  Smile,
  Trash2,
  Upload,
  User,
  Video,
} from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";
import { Spinner } from "@/components/ui/spinner";

const meta: Meta<typeof Command> = {
  title: "UI/Command",
  component: Command,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A fast, composable, unstyled command menu for React. Built on top of Radix UI Dialog and cmdk.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    shouldFilter: {
      control: "boolean",
      description: "Whether the command should filter items based on search",
    },
    filter: {
      description: "Custom filter function for command items",
    },
    defaultValue: {
      control: "text",
      description: "Default search value",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic command menu
export const Default: Story = {
  render: () => (
    <Command className="max-w-md rounded-lg border shadow-md">
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Suggestions">
          <CommandItem>
            <Calendar className="size-4" />
            <span>Calendar</span>
          </CommandItem>
          <CommandItem>
            <Smile className="size-4" />
            <span>Search Emoji</span>
          </CommandItem>
          <CommandItem>
            <Calculator className="size-4" />
            <span>Calculator</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Settings">
          <CommandItem>
            <User className="size-4" />
            <span>Profile</span>
            <CommandShortcut>⌘P</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <CreditCard className="size-4" />
            <span>Billing</span>
            <CommandShortcut>⌘B</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <Settings className="size-4" />
            <span>Settings</span>
            <CommandShortcut>⌘S</CommandShortcut>
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  ),
};

// Command dialog
export const Dialog: Story = {
  render: () => {
    const [open, setOpen] = useState(false);

    return (
      <>
        <Button onClick={() => setOpen(true)}>Open Command Dialog</Button>
        <CommandDialog open={open} onOpenChange={setOpen}>
          <CommandInput placeholder="Type a command or search..." />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup heading="Quick Actions">
              <CommandItem onSelect={() => setOpen(false)}>
                <Search className="size-4" />
                <span>Search Files</span>
                <CommandShortcut>⌘K</CommandShortcut>
              </CommandItem>
              <CommandItem onSelect={() => setOpen(false)}>
                <Mail className="size-4" />
                <span>Send Email</span>
                <CommandShortcut>⌘E</CommandShortcut>
              </CommandItem>
              <CommandItem onSelect={() => setOpen(false)}>
                <Phone className="size-4" />
                <span>Make Call</span>
                <CommandShortcut>⌘C</CommandShortcut>
              </CommandItem>
            </CommandGroup>
            <CommandSeparator />
            <CommandGroup heading="Navigation">
              <CommandItem onSelect={() => setOpen(false)}>
                <Home className="size-4" />
                <span>Go Home</span>
              </CommandItem>
              <CommandItem onSelect={() => setOpen(false)}>
                <Bell className="size-4" />
                <span>Notifications</span>
              </CommandItem>
              <CommandItem onSelect={() => setOpen(false)}>
                <Settings className="size-4" />
                <span>Settings</span>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </CommandDialog>
      </>
    );
  },
};

// File manager command
export const FileManager: Story = {
  render: () => (
    <Command className="max-w-md rounded-lg border shadow-md">
      <CommandInput placeholder="Search files and folders..." />
      <CommandList>
        <CommandEmpty>No files found.</CommandEmpty>
        <CommandGroup heading="Recent Files">
          <CommandItem>
            <File className="size-4" />
            <span>document.pdf</span>
          </CommandItem>
          <CommandItem>
            <Image className="size-4" />
            <span>screenshot.png</span>
          </CommandItem>
          <CommandItem>
            <Music className="size-4" />
            <span>song.mp3</span>
          </CommandItem>
          <CommandItem>
            <Video className="size-4" />
            <span>video.mp4</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Folders">
          <CommandItem>
            <Folder className="size-4" />
            <span>Documents</span>
          </CommandItem>
          <CommandItem>
            <Folder className="size-4" />
            <span>Downloads</span>
          </CommandItem>
          <CommandItem>
            <Folder className="size-4" />
            <span>Pictures</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Actions">
          <CommandItem>
            <Upload className="size-4" />
            <span>Upload File</span>
            <CommandShortcut>⌘U</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <Download className="size-4" />
            <span>Download</span>
            <CommandShortcut>⌘D</CommandShortcut>
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  ),
};

// Text editor commands
export const TextEditor: Story = {
  render: () => (
    <Command className="max-w-md rounded-lg border shadow-md">
      <CommandInput placeholder="Search commands..." />
      <CommandList>
        <CommandEmpty>No commands found.</CommandEmpty>
        <CommandGroup heading="Edit">
          <CommandItem>
            <Copy className="size-4" />
            <span>Copy</span>
            <CommandShortcut>⌘C</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <Scissors className="size-4" />
            <span>Cut</span>
            <CommandShortcut>⌘X</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <ClipboardPaste className="size-4" />
            <span>Paste</span>
            <CommandShortcut>⌘V</CommandShortcut>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Format">
          <CommandItem>
            <span className="mr-2 font-bold">B</span>
            <span>Bold</span>
            <CommandShortcut>⌘B</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <span className="mr-2 italic">I</span>
            <span>Italic</span>
            <CommandShortcut>⌘I</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <span className="mr-2 underline">U</span>
            <span>Underline</span>
            <CommandShortcut>⌘U</CommandShortcut>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="File">
          <CommandItem>
            <File className="size-4" />
            <span>New File</span>
            <CommandShortcut>⌘N</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <Edit className="size-4" />
            <span>Save</span>
            <CommandShortcut>⌘S</CommandShortcut>
          </CommandItem>
          <CommandItem>
            <Trash2 className="size-4" />
            <span>Delete</span>
            <CommandShortcut>⌫</CommandShortcut>
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  ),
};

// Simple search
export const SimpleSearch: Story = {
  render: () => (
    <Command className="max-w-md rounded-lg border shadow-md">
      <CommandInput placeholder="Search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup>
          <CommandItem>Apple</CommandItem>
          <CommandItem>Banana</CommandItem>
          <CommandItem>Cherry</CommandItem>
          <CommandItem>Date</CommandItem>
          <CommandItem>Elderberry</CommandItem>
          <CommandItem>Fig</CommandItem>
          <CommandItem>Grape</CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  ),
};

// Without filtering
export const WithoutFiltering: Story = {
  render: () => (
    <Command
      shouldFilter={false}
      className="max-w-md rounded-lg border shadow-md"
    >
      <CommandInput placeholder="Type to see all items..." />
      <CommandList>
        <CommandEmpty>Type something to see results.</CommandEmpty>
        <CommandGroup heading="Always Visible">
          <CommandItem>Item 1</CommandItem>
          <CommandItem>Item 2</CommandItem>
          <CommandItem>Item 3</CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  ),
};

// Loading state
export const LoadingState: Story = {
  render: () => (
    <Command className="max-w-md rounded-lg border shadow-md">
      <CommandInput placeholder="Search..." />
      <CommandList>
        <CommandEmpty>
          <div className="flex items-center justify-center py-6">
            <Spinner size="xs" className="mr-2" />
            <span>Loading...</span>
          </div>
        </CommandEmpty>
      </CommandList>
    </Command>
  ),
};

// Large list example
export const LargeList: Story = {
  render: () => (
    <Command className="h-96 max-w-md rounded-lg border shadow-md">
      <CommandInput placeholder="Search from many items..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Countries">
          {[
            "Afghanistan",
            "Albania",
            "Algeria",
            "Argentina",
            "Australia",
            "Austria",
            "Bangladesh",
            "Belgium",
            "Brazil",
            "Canada",
            "China",
            "Colombia",
            "Denmark",
            "Egypt",
            "Finland",
            "France",
            "Germany",
            "Greece",
            "India",
            "Indonesia",
            "Ireland",
            "Italy",
            "Japan",
            "Kenya",
            "Malaysia",
            "Mexico",
            "Netherlands",
            "Norway",
            "Pakistan",
            "Philippines",
            "Poland",
            "Portugal",
            "Russia",
            "South Africa",
            "Spain",
            "Sweden",
            "Switzerland",
            "Thailand",
            "Turkey",
            "Ukraine",
            "United Kingdom",
            "United States",
            "Vietnam",
          ].map((country) => (
            <CommandItem key={country}>
              <span>{country}</span>
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A command menu with many items to demonstrate scrolling and search functionality.",
      },
    },
  },
};
