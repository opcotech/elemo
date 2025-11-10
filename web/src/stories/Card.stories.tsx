import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Bookmark,
  Calendar,
  DollarSign,
  Heart,
  MapPin,
  MoreHorizontal,
  Share,
  Star,
  TrendingUp,
  Users,
} from "lucide-react";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";

const meta: Meta<typeof Card> = {
  title: "UI/Card",
  component: Card,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A flexible card component, perfect for displaying content in an organized way.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic card
export const Default: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Card Title</CardTitle>
        <CardDescription>
          Card description goes here. This provides additional context about the
          card content.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p>
          This is the main content area of the card where you can place any
          content.
        </p>
      </CardContent>
    </Card>
  ),
};

// Card with action
export const WithAction: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Project Update</CardTitle>
        <CardDescription>
          Latest updates on the development progress
        </CardDescription>
        <CardAction>
          <Button variant="ghost" size="icon">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <p>The new features have been implemented and are ready for testing.</p>
      </CardContent>
    </Card>
  ),
};

// Card with footer
export const WithFooter: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Task Management</CardTitle>
        <CardDescription>Keep track of your daily tasks</CardDescription>
      </CardHeader>
      <CardContent>
        <p>You have 5 tasks remaining for today.</p>
      </CardContent>
      <CardFooter>
        <Button variant="outline" className="w-full">
          View All Tasks
        </Button>
      </CardFooter>
    </Card>
  ),
};

// Blog post card
export const BlogPost: Story = {
  render: () => (
    <Card className="w-[400px]">
      <CardHeader>
        <CardTitle>Getting Started with React</CardTitle>
        <CardDescription>
          A comprehensive guide to building your first React application
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          <img
            src="https://images.unsplash.com/photo-1633356122544-f134324a6cee?w=400&h=200&fit=crop"
            alt="React guide"
            className="h-32 w-full rounded-lg object-cover"
          />
          <p className="text-muted-foreground text-sm">
            Learn the fundamentals of React including components, props, state,
            and hooks...
          </p>
          <div className="flex items-center space-x-2">
            <Badge variant="secondary">React</Badge>
            <Badge variant="secondary">Tutorial</Badge>
            <Badge variant="secondary">Beginner</Badge>
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex justify-between">
        <div className="text-muted-foreground flex items-center space-x-2 text-sm">
          <Calendar className="h-4 w-4" />
          <span>Dec 15, 2024</span>
        </div>
        <div className="flex space-x-2">
          <Button variant="ghost" size="icon">
            <Heart className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon">
            <Share className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon">
            <Bookmark className="h-4 w-4" />
          </Button>
        </div>
      </CardFooter>
    </Card>
  ),
};

// User profile card
export const UserProfile: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <div className="flex items-center space-x-4">
          <Avatar>
            <img
              src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop&crop=face"
              alt="Profile"
            />
          </Avatar>
          <div className="flex-1">
            <CardTitle>John Doe</CardTitle>
            <CardDescription>Software Engineer</CardDescription>
          </div>
        </div>
        <CardAction>
          <Button size="sm">Follow</Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="text-muted-foreground flex items-center text-sm">
            <MapPin className="size-4" />
            San Francisco, CA
          </div>
          <div className="text-muted-foreground flex items-center text-sm">
            <Users className="size-4" />
            2.5k followers
          </div>
          <p className="text-sm">
            Passionate about building great user experiences and scalable
            applications.
          </p>
        </div>
      </CardContent>
    </Card>
  ),
};

// Stats card
export const StatsCard: Story = {
  render: () => (
    <Card className="w-[300px]">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
        <DollarSign className="text-muted-foreground h-4 w-4" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">$45,231.89</div>
        <p className="text-muted-foreground text-xs">
          <Badge variant="success" className="gap-1 px-1.5 py-0.5">
            <TrendingUp className="mr-1 h-3 w-3" />
            +20.1%
          </Badge>
          from last month
        </p>
      </CardContent>
    </Card>
  ),
};

// Progress card
export const ProgressCard: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Project Progress</CardTitle>
        <CardDescription>Development milestone tracking</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span>Frontend Development</span>
            <span>80%</span>
          </div>
          <Progress value={80} />
        </div>
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span>Backend Development</span>
            <span>65%</span>
          </div>
          <Progress value={65} />
        </div>
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span>Testing</span>
            <span>40%</span>
          </div>
          <Progress value={40} />
        </div>
      </CardContent>
    </Card>
  ),
};

// Interactive card
export const InteractiveCard: Story = {
  render: () => (
    <Card className="hover:shadow-elegant-lg w-[350px] cursor-pointer transition-all duration-300 hover:scale-[1.02]">
      <CardHeader>
        <CardTitle>Premium Plan</CardTitle>
        <CardDescription>
          Unlock advanced features and priority support
        </CardDescription>
        <CardAction>
          <Badge>Popular</Badge>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          <div className="text-3xl font-bold">
            $29
            <span className="text-muted-foreground text-lg font-normal">
              /month
            </span>
          </div>
          <ul className="space-y-2 text-sm">
            <li className="flex items-center">
              <Star className="size-4 text-yellow-500" />
              Unlimited projects
            </li>
            <li className="flex items-center">
              <Star className="size-4 text-yellow-500" />
              Priority support
            </li>
            <li className="flex items-center">
              <Star className="size-4 text-yellow-500" />
              Advanced analytics
            </li>
          </ul>
        </div>
      </CardContent>
      <CardFooter>
        <Button className="w-full">Upgrade Now</Button>
      </CardFooter>
    </Card>
  ),
};

// Card with separator
export const WithSeparator: Story = {
  render: () => (
    <Card className="w-[350px]">
      <CardHeader>
        <CardTitle>Order Summary</CardTitle>
        <CardDescription>Review your purchase details</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex justify-between">
          <span>Subtotal</span>
          <span>$99.00</span>
        </div>
        <div className="flex justify-between">
          <span>Shipping</span>
          <span>$4.99</span>
        </div>
        <div className="flex justify-between">
          <span>Tax</span>
          <span>$8.32</span>
        </div>
        <Separator />
        <div className="flex justify-between font-semibold">
          <span>Total</span>
          <span>$112.31</span>
        </div>
      </CardContent>
      <CardFooter>
        <Button className="w-full">Proceed to Payment</Button>
      </CardFooter>
    </Card>
  ),
};

// All card types showcase
export const AllCardTypes: Story = {
  render: () => (
    <div className="grid grid-cols-1 gap-6 p-6 md:grid-cols-2 lg:grid-cols-3">
      <Card className="w-[300px]">
        <CardHeader>
          <CardTitle>Simple Card</CardTitle>
          <CardDescription>
            Basic card with title and description
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p>Simple content goes here.</p>
        </CardContent>
      </Card>

      <Card className="w-[300px]">
        <CardHeader>
          <CardTitle>With Action</CardTitle>
          <CardDescription>Card with header action</CardDescription>
          <CardAction>
            <Button variant="ghost" size="icon">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <p>Content with action in header.</p>
        </CardContent>
      </Card>

      <Card className="w-[300px]">
        <CardHeader>
          <CardTitle>With Footer</CardTitle>
          <CardDescription>Card with footer actions</CardDescription>
        </CardHeader>
        <CardContent>
          <p>Content with footer.</p>
        </CardContent>
        <CardFooter>
          <Button variant="outline" className="w-full">
            Action
          </Button>
        </CardFooter>
      </Card>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Various card layouts and configurations displayed together.",
      },
    },
  },
};
