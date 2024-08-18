import type { Meta, StoryObj } from "@storybook/react-vite";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";

const meta: Meta<typeof Skeleton> = {
  title: "UI/Skeleton",
  component: Skeleton,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A placeholder component used to show a loading state while content is being fetched or processed. Provides visual feedback to users during loading states.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the skeleton",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic skeleton
export const Default: Story = {
  render: () => (
    <div className="flex items-center space-x-4">
      <Skeleton className="h-12 w-12 rounded-full" />
      <div className="space-y-2">
        <Skeleton className="h-4 w-[250px]" />
        <Skeleton className="h-4 w-[200px]" />
      </div>
    </div>
  ),
};

// Different shapes
export const Shapes: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="space-y-2">
        <h4 className="text-sm font-medium">Rectangles</h4>
        <Skeleton className="h-4 w-[250px]" />
        <Skeleton className="h-4 w-[200px]" />
        <Skeleton className="h-4 w-[150px]" />
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-medium">Squares</h4>
        <div className="flex space-x-2">
          <Skeleton className="h-12 w-12" />
          <Skeleton className="h-16 w-16" />
          <Skeleton className="h-20 w-20" />
        </div>
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-medium">Circles</h4>
        <div className="flex space-x-2">
          <Skeleton className="h-8 w-8 rounded-full" />
          <Skeleton className="h-12 w-12 rounded-full" />
          <Skeleton className="h-16 w-16 rounded-full" />
        </div>
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-medium">Rounded Rectangles</h4>
        <Skeleton className="h-8 w-[200px] rounded-md" />
        <Skeleton className="h-10 w-[150px] rounded-lg" />
        <Skeleton className="h-12 w-[100px] rounded-xl" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Skeletons in various shapes including rectangles, squares, circles, and rounded rectangles.",
      },
    },
  },
};

// Card skeleton
export const CardSkeleton: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader className="space-y-2">
        <div className="flex items-center space-x-4">
          <Skeleton className="h-12 w-12 rounded-full" />
          <div className="space-y-2">
            <Skeleton className="h-4 w-[150px]" />
            <Skeleton className="h-4 w-[100px]" />
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <Skeleton className="h-[200px] w-full rounded-md" />
        <div className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-[80%]" />
          <Skeleton className="h-4 w-[60%]" />
        </div>
        <div className="flex space-x-2">
          <Skeleton className="h-8 w-[80px] rounded-md" />
          <Skeleton className="h-8 w-[60px] rounded-md" />
        </div>
      </CardContent>
    </Card>
  ),
};

// Article skeleton
export const ArticleSkeleton: Story = {
  render: () => (
    <div className="w-[600px] space-y-6">
      <div className="space-y-2">
        <Skeleton className="h-8 w-[80%]" />
        <Skeleton className="h-6 w-[60%]" />
      </div>

      <div className="flex items-center space-x-4">
        <Skeleton className="h-10 w-10 rounded-full" />
        <div className="space-y-2">
          <Skeleton className="h-4 w-[120px]" />
          <Skeleton className="h-3 w-[80px]" />
        </div>
      </div>

      <Skeleton className="h-[300px] w-full rounded-lg" />

      <div className="space-y-3">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-[90%]" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-[85%]" />
      </div>

      <div className="space-y-3">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-[95%]" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-[70%]" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A skeleton layout for an article with title, author, image, and content paragraphs.",
      },
    },
  },
};

// User profile skeleton
export const UserProfileSkeleton: Story = {
  render: () => (
    <div className="w-[400px] space-y-6">
      <div className="flex items-center space-x-4">
        <Skeleton className="h-20 w-20 rounded-full" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-6 w-[150px]" />
          <Skeleton className="h-4 w-[100px]" />
          <Skeleton className="h-4 w-[200px]" />
        </div>
      </div>

      <Separator />

      <div className="space-y-4">
        <div className="space-y-2">
          <Skeleton className="h-5 w-[100px]" />
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1 text-center">
              <Skeleton className="mx-auto h-6 w-[50px]" />
              <Skeleton className="mx-auto h-3 w-[60px]" />
            </div>
            <div className="space-y-1 text-center">
              <Skeleton className="mx-auto h-6 w-[50px]" />
              <Skeleton className="mx-auto h-3 w-[70px]" />
            </div>
          </div>
        </div>

        <Separator />

        <div className="space-y-3">
          <Skeleton className="h-5 w-[120px]" />
          <div className="space-y-2">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-[85%]" />
            <Skeleton className="h-4 w-[60%]" />
          </div>
        </div>

        <div className="flex space-x-2">
          <Skeleton className="h-9 w-[100px] rounded-md" />
          <Skeleton className="h-9 w-[80px] rounded-md" />
        </div>
      </div>
    </div>
  ),
};

// Table skeleton
export const TableSkeleton: Story = {
  render: () => (
    <div className="w-[600px] space-y-4">
      <div className="space-y-2">
        <Skeleton className="h-6 w-[150px]" />
        <Skeleton className="h-4 w-[300px]" />
      </div>

      <div className="rounded-lg border">
        <div className="border-b p-4">
          <div className="flex space-x-4">
            <Skeleton className="h-4 w-[100px]" />
            <Skeleton className="h-4 w-[120px]" />
            <Skeleton className="h-4 w-[80px]" />
            <Skeleton className="h-4 w-[90px]" />
            <Skeleton className="h-4 w-[60px]" />
          </div>
        </div>

        {Array.from({ length: 5 }, (_, i) => (
          <div key={i} className="border-b p-4 last:border-b-0">
            <div className="flex items-center space-x-4">
              <Skeleton className="h-8 w-8 rounded-full" />
              <Skeleton className="h-4 w-[120px]" />
              <Skeleton className="h-4 w-[80px]" />
              <Skeleton className="h-4 w-[90px]" />
              <Skeleton className="h-6 w-[60px] rounded-md" />
            </div>
          </div>
        ))}
      </div>
    </div>
  ),
};

// Dashboard skeleton
export const DashboardSkeleton: Story = {
  render: () => (
    <div className="w-[800px] space-y-6">
      <div className="space-y-2">
        <Skeleton className="h-8 w-[200px]" />
        <Skeleton className="h-4 w-[400px]" />
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        {Array.from({ length: 3 }, (_, i) => (
          <Card key={i}>
            <CardContent className="space-y-2 p-6">
              <div className="flex items-center justify-between">
                <Skeleton className="h-4 w-[80px]" />
                <Skeleton className="h-4 w-4 rounded" />
              </div>
              <Skeleton className="h-8 w-[100px]" />
              <Skeleton className="h-3 w-[120px]" />
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        <Card>
          <CardHeader className="space-y-2">
            <Skeleton className="h-5 w-[150px]" />
            <Skeleton className="h-4 w-[200px]" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-[200px] w-full rounded-md" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="space-y-2">
            <Skeleton className="h-5 w-[120px]" />
            <Skeleton className="h-4 w-[180px]" />
          </CardHeader>
          <CardContent className="space-y-3">
            {Array.from({ length: 4 }, (_, i) => (
              <div key={i} className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <Skeleton className="h-6 w-6 rounded-full" />
                  <Skeleton className="h-4 w-[100px]" />
                </div>
                <Skeleton className="h-4 w-[50px]" />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A complete dashboard skeleton with metrics cards and chart placeholders.",
      },
    },
  },
};

// Chat message skeleton
export const ChatMessageSkeleton: Story = {
  render: () => (
    <div className="w-[400px] space-y-4">
      <div className="flex items-start space-x-3">
        <Skeleton className="h-8 w-8 rounded-full" />
        <div className="flex-1 space-y-2">
          <div className="flex items-center space-x-2">
            <Skeleton className="h-4 w-[80px]" />
            <Skeleton className="h-3 w-[60px]" />
          </div>
          <Skeleton className="h-10 w-[250px] rounded-lg" />
        </div>
      </div>

      <div className="flex flex-row-reverse items-start space-x-3">
        <Skeleton className="h-8 w-8 rounded-full" />
        <div className="flex-1 items-end space-y-2">
          <div className="flex items-center justify-end space-x-2">
            <Skeleton className="h-3 w-[60px]" />
            <Skeleton className="h-4 w-[60px]" />
          </div>
          <Skeleton className="ml-auto h-16 w-[200px] rounded-lg" />
        </div>
      </div>

      <div className="flex items-start space-x-3">
        <Skeleton className="h-8 w-8 rounded-full" />
        <div className="flex-1 space-y-2">
          <div className="flex items-center space-x-2">
            <Skeleton className="h-4 w-[80px]" />
            <Skeleton className="h-3 w-[60px]" />
          </div>
          <div className="space-y-1">
            <Skeleton className="h-6 w-[180px] rounded-lg" />
            <Skeleton className="h-6 w-[220px] rounded-lg" />
          </div>
        </div>
      </div>
    </div>
  ),
};

// List skeleton
export const ListSkeleton: Story = {
  render: () => (
    <div className="w-[400px] space-y-1">
      {Array.from({ length: 6 }, (_, i) => (
        <div
          key={i}
          className="flex items-center space-x-3 rounded-lg border p-3"
        >
          <Skeleton className="h-10 w-10 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-[150px]" />
            <Skeleton className="h-3 w-[100px]" />
          </div>
          <Skeleton className="h-6 w-[60px] rounded-md" />
        </div>
      ))}
    </div>
  ),
};

// Form skeleton
export const FormSkeleton: Story = {
  render: () => (
    <div className="w-[400px] space-y-6">
      <div className="space-y-2">
        <Skeleton className="h-6 w-[120px]" />
        <Skeleton className="h-4 w-[250px]" />
      </div>

      <div className="space-y-4">
        <div className="space-y-2">
          <Skeleton className="h-4 w-[80px]" />
          <Skeleton className="h-10 w-full rounded-md" />
        </div>

        <div className="space-y-2">
          <Skeleton className="h-4 w-[60px]" />
          <Skeleton className="h-10 w-full rounded-md" />
        </div>

        <div className="space-y-2">
          <Skeleton className="h-4 w-[100px]" />
          <Skeleton className="h-24 w-full rounded-md" />
        </div>

        <div className="space-y-2">
          <Skeleton className="h-4 w-[90px]" />
          <Skeleton className="h-10 w-full rounded-md" />
        </div>

        <div className="flex items-center space-x-2">
          <Skeleton className="h-4 w-4 rounded" />
          <Skeleton className="h-4 w-[200px]" />
        </div>

        <div className="flex space-x-2 pt-4">
          <Skeleton className="h-10 w-[100px] rounded-md" />
          <Skeleton className="h-10 w-[80px] rounded-md" />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "A form skeleton with various input types and buttons.",
      },
    },
  },
};
