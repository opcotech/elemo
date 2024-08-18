import type { Meta, StoryObj } from "@storybook/react-vite";
import { useEffect, useState } from "react";

import { BreadcrumbNav, BreadcrumbProvider } from "@/components/breadcrumb";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";

const meta: Meta<typeof BreadcrumbNav> = {
  title: "Components/Breadcrumb",
  component: BreadcrumbNav,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A breadcrumb navigation component that displays the current page hierarchy and allows navigation to parent pages.",
      },
    },
  },
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <BreadcrumbProvider>
        <Story />
      </BreadcrumbProvider>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof BreadcrumbNav>;

// Simple breadcrumb
export const Simple: Story = {
  render: () => {
    const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

    useEffect(() => {
      setBreadcrumbsFromItems([
        { label: "Home", href: "/", isNavigatable: true },
        { label: "Dashboard", isNavigatable: false },
      ]);
    }, []);

    return <BreadcrumbNav />;
  },
};

// Complex breadcrumb with multiple levels
export const Complex: Story = {
  render: () => {
    const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

    useEffect(() => {
      setBreadcrumbsFromItems([
        { label: "Home", href: "/", isNavigatable: true },
        { label: "Projects", href: "/projects", isNavigatable: true },
        {
          label: "Project Alpha",
          href: "/projects/alpha",
          isNavigatable: true,
        },
        { label: "Tasks", href: "/projects/alpha/tasks", isNavigatable: true },
        { label: "Task Details", isNavigatable: false },
      ]);
    }, []);

    return <BreadcrumbNav />;
  },
};

// Non-navigatable breadcrumbs (using / ... / pattern)
export const NonNavigatable: Story = {
  render: () => {
    const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

    useEffect(() => {
      setBreadcrumbsFromItems([
        { label: "Home", href: "/", isNavigatable: true },
        { label: "...", isNavigatable: false }, // Non-navigatable separator
        { label: "Settings", href: "/settings", isNavigatable: true },
        { label: "Profile", isNavigatable: false },
      ]);
    }, []);

    return <BreadcrumbNav />;
  },
  parameters: {
    docs: {
      description: {
        story:
          "Breadcrumb with non-navigatable segments using '...' to indicate skipped levels.",
      },
    },
  },
};

// Interactive breadcrumb demo
export const Interactive: Story = {
  render: () => {
    const { setBreadcrumbsFromItems, clearBreadcrumbs } = useBreadcrumbUtils();
    const [, setCurrentPath] = useState("dashboard");

    const paths = {
      dashboard: [{ label: "Dashboard", isNavigatable: false }],
      projects: [
        { label: "Dashboard", href: "/dashboard", isNavigatable: true },
        { label: "Projects", isNavigatable: false },
      ],
      projectDetails: [
        { label: "Dashboard", href: "/dashboard", isNavigatable: true },
        { label: "Projects", href: "/projects", isNavigatable: true },
        { label: "Project Alpha", isNavigatable: false },
      ],
      tasks: [
        { label: "Dashboard", href: "/dashboard", isNavigatable: true },
        { label: "Projects", href: "/projects", isNavigatable: true },
        {
          label: "Project Alpha",
          href: "/projects/alpha",
          isNavigatable: true,
        },
        { label: "Tasks", isNavigatable: false },
      ],
    };

    const handlePathChange = (path: keyof typeof paths) => {
      setCurrentPath(path);
      setBreadcrumbsFromItems(paths[path]);
    };

    return (
      <div className="space-y-4">
        <BreadcrumbNav />
        <div className="flex gap-2">
          <button
            onClick={() => handlePathChange("dashboard")}
            className="rounded border px-3 py-1 text-sm hover:bg-gray-100"
          >
            Dashboard
          </button>
          <button
            onClick={() => handlePathChange("projects")}
            className="rounded border px-3 py-1 text-sm hover:bg-gray-100"
          >
            Projects
          </button>
          <button
            onClick={() => handlePathChange("projectDetails")}
            className="rounded border px-3 py-1 text-sm hover:bg-gray-100"
          >
            Project Details
          </button>
          <button
            onClick={() => handlePathChange("tasks")}
            className="rounded border px-3 py-1 text-sm hover:bg-gray-100"
          >
            Tasks
          </button>
          <button
            onClick={clearBreadcrumbs}
            className="rounded border px-3 py-1 text-sm hover:bg-gray-100"
          >
            Clear
          </button>
        </div>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "Interactive demo showing different breadcrumb states. Click the buttons to see how breadcrumbs change.",
      },
    },
  },
};

// Empty breadcrumb (should not render)
export const Empty: Story = {
  render: () => {
    const { clearBreadcrumbs } = useBreadcrumbUtils();

    useEffect(() => {
      clearBreadcrumbs();
    }, []);

    return (
      <div className="space-y-2">
        <BreadcrumbNav />
        <p className="text-sm text-gray-500">
          No breadcrumbs set (component should not render)
        </p>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "When no breadcrumbs are set, the component should not render anything.",
      },
    },
  },
};
