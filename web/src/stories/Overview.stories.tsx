import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  AlertTriangle,
  CheckCircle,
  Download,
  Edit,
  Info,
  Mail,
  Plus,
  Search,
  Settings,
  Trash2,
  User,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const meta: Meta = {
  title: "Overview",
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "A comprehensive overview of all available UI components in the Elemo design system.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const AllComponents: Story = {
  render: () => (
    <div className="bg-background min-h-screen space-y-12 p-8">
      {/* Header */}
      <div className="space-y-4 text-center">
        <h1 className="text-4xl font-bold">Elemo Design System</h1>
        <p className="text-muted-foreground text-xl">
          Modern, minimal UI components built with Tailwind CSS v4 and Radix UI
        </p>
      </div>

      {/* Buttons Section */}
      <section className="space-y-6">
        <h2 className="text-2xl font-semibold">Buttons</h2>
        <div className="space-y-4">
          <div className="flex flex-wrap gap-3">
            <Button>Default</Button>
            <Button variant="destructive">Destructive</Button>
            <Button variant="outline">Outline</Button>
            <Button variant="secondary">Secondary</Button>
            <Button variant="ghost">Ghost</Button>
          </div>
          <div className="flex items-center gap-3">
            <Button size="sm">Small</Button>
            <Button>Default</Button>
            <Button size="lg">Large</Button>
            <Button size="icon">
              <Settings className="h-4 w-4" />
            </Button>
          </div>
          <div className="flex gap-3">
            <Button>
              <Plus className="h-4 w-4" />
              Add New
            </Button>
            <Button variant="outline">
              <Edit className="h-4 w-4" />
              Edit
            </Button>
            <Button variant="destructive">
              <Trash2 className="h-4 w-4" />
              Delete
            </Button>
          </div>
        </div>
      </section>

      {/* Badges Section */}
      <section className="space-y-6">
        <h2 className="text-2xl font-semibold">Badges</h2>
        <div className="flex flex-wrap gap-2">
          <Badge>Default</Badge>
          <Badge variant="secondary">Secondary</Badge>
          <Badge variant="destructive">Destructive</Badge>
          <Badge variant="outline">Outline</Badge>
          <Badge variant="success">
            <CheckCircle className="h-3 w-3" />
            Success
          </Badge>
          <Badge variant="warning">
            <AlertTriangle className="h-3 w-3" />
            Warning
          </Badge>
        </div>
      </section>

      {/* Alerts Section */}
      <section className="space-y-6">
        <h2 className="text-2xl font-semibold">Alerts</h2>
        <div className="max-w-2xl space-y-4">
          <Alert variant="info">
            <Info className="h-4 w-4" />
            <AlertTitle>Information</AlertTitle>
            <AlertDescription>
              This is an informational alert with info styling.
            </AlertDescription>
          </Alert>
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>
              This is a destructive alert indicating an error.
            </AlertDescription>
          </Alert>
          <Alert variant="success">
            <CheckCircle className="h-4 w-4" />
            <AlertTitle>Success</AlertTitle>
            <AlertDescription>
              This is a success alert indicating a positive outcome.
            </AlertDescription>
          </Alert>
          <Alert variant="warning">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Warning</AlertTitle>
            <AlertDescription>
              This is a warning alert indicating caution is needed.
            </AlertDescription>
          </Alert>
        </div>
      </section>

      {/* Form Elements Section */}
      <section className="space-y-6">
        <h2 className="text-2xl font-semibold">Form Elements</h2>
        <div className="grid max-w-6xl grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {/* Input Examples */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Inputs</h3>
            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input id="name" placeholder="Enter your name" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <div className="relative">
                  <Mail className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                  <Input
                    id="email"
                    type="email"
                    placeholder="email@example.com"
                    className="pl-10"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="search">Search</Label>
                <div className="relative">
                  <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                  <Input
                    id="search"
                    type="search"
                    placeholder="Search..."
                    className="pl-10"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Checkbox Examples */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Checkboxes & Switches</h3>
            <div className="space-y-3">
              <div className="flex items-center space-x-2">
                <Checkbox id="newsletter" />
                <Label htmlFor="newsletter">Newsletter</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox id="marketing" defaultChecked />
                <Label htmlFor="marketing">Marketing emails</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Switch id="notifications" />
                <Label htmlFor="notifications">Push notifications</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Switch id="dark-mode" defaultChecked />
                <Label htmlFor="dark-mode">Dark mode</Label>
              </div>
            </div>
          </div>

          {/* Progress */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">Progress</h3>
            <div className="space-y-3">
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Upload progress</span>
                  <span>75%</span>
                </div>
                <Progress value={75} />
              </div>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Profile completion</span>
                  <span>45%</span>
                </div>
                <Progress value={45} />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Cards Section */}
      <section className="space-y-6">
        <h2 className="text-2xl font-semibold">Cards</h2>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {/* Simple Card */}
          <Card>
            <CardHeader>
              <CardTitle>Simple Card</CardTitle>
              <CardDescription>
                A basic card with title and description.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p>
                This is the card content area where you can place any
                information.
              </p>
            </CardContent>
          </Card>

          {/* User Profile Card */}
          <Card>
            <CardHeader>
              <div className="flex items-center space-x-4">
                <Avatar>
                  <div className="bg-primary text-primary-foreground flex h-full w-full items-center justify-center">
                    <User className="h-5 w-5" />
                  </div>
                </Avatar>
                <div>
                  <CardTitle>John Doe</CardTitle>
                  <CardDescription>Software Engineer</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="text-muted-foreground flex items-center text-sm">
                  <Mail className="mr-2 h-4 w-4" />
                  john@example.com
                </div>
                <div className="flex space-x-1">
                  <Badge variant="secondary">React</Badge>
                  <Badge variant="secondary">TypeScript</Badge>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Action Card */}
          <Card>
            <CardHeader>
              <CardTitle>Project Tasks</CardTitle>
              <CardDescription>
                Manage your project tasks efficiently.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm">Completed</span>
                  <Badge variant="success">8/12</Badge>
                </div>
                <Progress value={67} />
              </div>
            </CardContent>
            <CardFooter>
              <Button className="w-full">
                <Plus className="h-4 w-4" />
                Add Task
              </Button>
            </CardFooter>
          </Card>
        </div>
      </section>

      {/* Tabs Section */}
      <section className="space-y-6">
        <h2 className="text-2xl font-semibold">Tabs</h2>
        <div className="max-w-2xl">
          <Tabs defaultValue="overview" className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="analytics">Analytics</TabsTrigger>
              <TabsTrigger value="settings">Settings</TabsTrigger>
            </TabsList>
            <TabsContent value="overview" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Overview</CardTitle>
                  <CardDescription>
                    Get a quick overview of your project metrics.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-3 gap-4 text-center">
                    <div>
                      <div className="text-2xl font-bold">24</div>
                      <div className="text-muted-foreground text-sm">
                        Active Tasks
                      </div>
                    </div>
                    <div>
                      <div className="text-2xl font-bold">8</div>
                      <div className="text-muted-foreground text-sm">
                        Team Members
                      </div>
                    </div>
                    <div>
                      <div className="text-2xl font-bold">95%</div>
                      <div className="text-muted-foreground text-sm">
                        Completion
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="analytics" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Analytics</CardTitle>
                  <CardDescription>
                    Detailed analytics and performance metrics.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <p>Analytics content would go here...</p>
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="settings" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Settings</CardTitle>
                  <CardDescription>
                    Configure your project settings.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Email notifications</Label>
                      <div className="text-muted-foreground text-sm">
                        Receive email updates about your projects.
                      </div>
                    </div>
                    <Switch />
                  </div>
                  <Separator />
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label>Push notifications</Label>
                      <div className="text-muted-foreground text-sm">
                        Get push notifications on your devices.
                      </div>
                    </div>
                    <Switch defaultChecked />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </section>

      {/* Interactive Example */}
      <section className="space-y-6">
        <h2 className="text-2xl font-semibold">Interactive Example</h2>
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>Create Account</CardTitle>
            <CardDescription>
              Fill out the form below to create your account.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="signup-name">Full Name</Label>
              <Input id="signup-name" placeholder="Enter your full name" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="signup-email">Email</Label>
              <Input
                id="signup-email"
                type="email"
                placeholder="Enter your email"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="signup-password">Password</Label>
              <Input
                id="signup-password"
                type="password"
                placeholder="Create a password"
              />
            </div>
            <div className="flex items-center space-x-2">
              <Checkbox id="terms" />
              <Label htmlFor="terms" className="text-sm">
                I agree to the{" "}
                <a href="#" className="underline">
                  terms and conditions
                </a>
              </Label>
            </div>
          </CardContent>
          <CardFooter className="flex flex-col space-y-2">
            <Button className="w-full">
              <User className="h-4 w-4" />
              Create Account
            </Button>
            <Button variant="outline" className="w-full">
              <Download className="h-4 w-4" />
              Sign in instead
            </Button>
          </CardFooter>
        </Card>
      </section>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A comprehensive showcase of all UI components in the Elemo design system, demonstrating how they work together to create beautiful, modern interfaces.",
      },
    },
  },
};

export const ComponentGrid: Story = {
  render: () => (
    <div className="grid grid-cols-1 gap-6 p-8 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {/* Button Examples */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Buttons</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Button className="w-full">Primary Button</Button>
          <Button variant="outline" className="w-full">
            Outline Button
          </Button>
          <Button variant="ghost" className="w-full">
            Ghost Button
          </Button>
        </CardContent>
      </Card>

      {/* Badge Examples */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Badges</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-2">
            <Badge>Default</Badge>
            <Badge variant="secondary">Secondary</Badge>
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge variant="success">Success</Badge>
            <Badge variant="destructive">Error</Badge>
          </div>
        </CardContent>
      </Card>

      {/* Form Examples */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Form Elements</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Input placeholder="Text input" />
          <div className="flex items-center space-x-2">
            <Checkbox id="example" />
            <Label htmlFor="example">Checkbox</Label>
          </div>
          <div className="flex items-center space-x-2">
            <Switch id="switch" />
            <Label htmlFor="switch">Switch</Label>
          </div>
        </CardContent>
      </Card>

      {/* Progress Example */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Progress</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Progress</span>
              <span>60%</span>
            </div>
            <Progress value={60} />
          </div>
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Loading</span>
              <span>80%</span>
            </div>
            <Progress value={80} />
          </div>
        </CardContent>
      </Card>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A grid layout showcasing different component categories in an organized manner.",
      },
    },
  },
};
