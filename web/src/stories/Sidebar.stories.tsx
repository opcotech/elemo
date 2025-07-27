import type { Meta, StoryObj } from "@storybook/react-vite";
import { Home, LogOut, Settings, User } from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";

const meta: Meta<typeof Sidebar> = {
  title: "UI/Sidebar",
  component: Sidebar,
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "A collapsible sidebar component for navigation and content organization. Provides a flexible layout system with support for groups, menus, and responsive behavior.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    side: {
      control: "select",
      options: ["left", "right"],
      description: "Which side the sidebar appears on",
    },
    variant: {
      control: "select",
      options: ["sidebar", "floating", "inset"],
      description: "The visual variant of the sidebar",
    },
    collapsible: {
      control: "select",
      options: ["offcanvas", "icon", "none"],
      description: "How the sidebar can be collapsed",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic sidebar
export const Default: Story = {
  render: () => (
    <SidebarProvider>
      <div className="flex h-screen w-screen">
        <Sidebar variant="inset">
          <SidebarHeader>
            <h2 className="px-4 py-2 text-lg font-semibold">My App</h2>
          </SidebarHeader>
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupLabel>Navigation</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  <SidebarMenuItem>
                    <SidebarMenuButton>
                      <Home className="h-4 w-4" />
                      <span>Home</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton>
                      <User className="h-4 w-4" />
                      <span>Profile</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton>
                      <Settings className="h-4 w-4" />
                      <span>Settings</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarFooter>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton>
                  <LogOut className="h-4 w-4" />
                  <span>Logout</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarFooter>
        </Sidebar>
        <SidebarInset>
          <main className="flex-1 p-6">
            <div className="mb-4 flex items-center gap-2">
              <SidebarTrigger />
              <h1 className="text-2xl font-bold">Main Content</h1>
            </div>
            <p className="text-muted-foreground">
              This is the main content area. Use the trigger button to toggle
              the sidebar.
            </p>
          </main>
        </SidebarInset>
      </div>
    </SidebarProvider>
  ),
};
