import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  Bell,
  Bluetooth,
  Eye,
  EyeOff,
  Lock,
  Moon,
  Plane,
  Settings,
  Shield,
  Smartphone,
  Sun,
  Volume2,
  VolumeX,
  Wifi,
  WifiOff,
} from "lucide-react";
import { useState } from "react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";

const meta: Meta<typeof Switch> = {
  title: "UI/Switch",
  component: Switch,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A control that allows the user to toggle between checked and not checked states. Built on top of Radix UI Switch primitive with modern styling.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    checked: {
      control: "boolean",
      description: "The controlled checked state of the switch",
    },
    defaultChecked: {
      control: "boolean",
      description: "The default checked state when uncontrolled",
    },
    onCheckedChange: {
      action: "onCheckedChange",
      description: "Callback fired when the checked state changes",
    },
    disabled: {
      control: "boolean",
      description: "Whether the switch is disabled",
    },
    required: {
      control: "boolean",
      description: "Whether the switch is required",
    },
    name: {
      control: "text",
      description: "The name of the switch for form submission",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic switch
export const Default: Story = {
  args: {
    defaultChecked: false,
  },
};

// With label
export const WithLabel: Story = {
  render: () => (
    <div className="flex items-center space-x-2">
      <Switch id="airplane-mode" />
      <Label htmlFor="airplane-mode">Airplane mode</Label>
    </div>
  ),
};

// Different states
export const States: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="flex items-center space-x-2">
        <Switch id="default-off" defaultChecked={false} />
        <Label htmlFor="default-off">Default (Off)</Label>
      </div>
      <div className="flex items-center space-x-2">
        <Switch id="default-on" defaultChecked={true} />
        <Label htmlFor="default-on">Default (On)</Label>
      </div>
      <div className="flex items-center space-x-2">
        <Switch id="disabled-off" disabled defaultChecked={false} />
        <Label htmlFor="disabled-off">Disabled (Off)</Label>
      </div>
      <div className="flex items-center space-x-2">
        <Switch id="disabled-on" disabled defaultChecked={true} />
        <Label htmlFor="disabled-on">Disabled (On)</Label>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Switches in different states including default and disabled variants.",
      },
    },
  },
};

// Controlled switch
export const Controlled: Story = {
  render: () => {
    const [checked, setChecked] = useState(false);

    return (
      <div className="space-y-2">
        <div className="text-muted-foreground text-sm">
          Switch is {checked ? "on" : "off"}
        </div>
        <div className="flex items-center space-x-2">
          <Switch
            id="controlled"
            checked={checked}
            onCheckedChange={setChecked}
          />
          <Label htmlFor="controlled">Controlled switch</Label>
        </div>
      </div>
    );
  },
};

// Settings panel
export const SettingsPanel: Story = {
  render: () => (
    <Card className="w-[400px]">
      <CardHeader>
        <CardTitle>Settings</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-4">
          <h4 className="text-sm font-medium">Notifications</h4>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Bell className="h-4 w-4" />
                <Label htmlFor="notifications" className="text-sm">
                  Push notifications
                </Label>
              </div>
              <Switch id="notifications" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Volume2 className="h-4 w-4" />
                <Label htmlFor="sound" className="text-sm">
                  Sound alerts
                </Label>
              </div>
              <Switch id="sound" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Smartphone className="h-4 w-4" />
                <Label htmlFor="mobile" className="text-sm">
                  Mobile notifications
                </Label>
              </div>
              <Switch id="mobile" />
            </div>
          </div>
        </div>

        <Separator />

        <div className="space-y-4">
          <h4 className="text-sm font-medium">Privacy</h4>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Eye className="h-4 w-4" />
                <Label htmlFor="visibility" className="text-sm">
                  Profile visibility
                </Label>
              </div>
              <Switch id="visibility" defaultChecked />
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Shield className="h-4 w-4" />
                <Label htmlFor="analytics" className="text-sm">
                  Analytics tracking
                </Label>
              </div>
              <Switch id="analytics" />
            </div>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Lock className="h-4 w-4" />
                <Label htmlFor="two-factor" className="text-sm">
                  Two-factor authentication
                </Label>
              </div>
              <Switch id="two-factor" defaultChecked />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  ),
};

// Device controls
export const DeviceControls: Story = {
  render: () => {
    const [wifi, setWifi] = useState(true);
    const [bluetooth, setBluetooth] = useState(false);
    const [airplane, setAirplane] = useState(false);
    const [darkMode, setDarkMode] = useState(false);

    return (
      <Card className="w-[350px]">
        <CardHeader>
          <CardTitle>Device Controls</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              {wifi ? (
                <Wifi className="h-5 w-5 text-blue-500" />
              ) : (
                <WifiOff className="text-muted-foreground h-5 w-5" />
              )}
              <div>
                <Label htmlFor="wifi" className="text-sm font-medium">
                  Wi-Fi
                </Label>
                <div className="text-muted-foreground text-xs">
                  {wifi ? "Connected to Home Network" : "Disconnected"}
                </div>
              </div>
            </div>
            <Switch id="wifi" checked={wifi} onCheckedChange={setWifi} />
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <Bluetooth
                className={`h-5 w-5 ${
                  bluetooth ? "text-blue-500" : "text-muted-foreground"
                }`}
              />
              <div>
                <Label htmlFor="bluetooth" className="text-sm font-medium">
                  Bluetooth
                </Label>
                <div className="text-muted-foreground text-xs">
                  {bluetooth ? "2 devices connected" : "No devices"}
                </div>
              </div>
            </div>
            <Switch
              id="bluetooth"
              checked={bluetooth}
              onCheckedChange={setBluetooth}
            />
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <Plane
                className={`h-5 w-5 ${
                  airplane ? "text-orange-500" : "text-muted-foreground"
                }`}
              />
              <div>
                <Label htmlFor="airplane" className="text-sm font-medium">
                  Airplane Mode
                </Label>
                <div className="text-muted-foreground text-xs">
                  {airplane ? "All wireless off" : "Wireless enabled"}
                </div>
              </div>
            </div>
            <Switch
              id="airplane"
              checked={airplane}
              onCheckedChange={setAirplane}
            />
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              {darkMode ? (
                <Moon className="h-5 w-5" />
              ) : (
                <Sun className="h-5 w-5" />
              )}
              <div>
                <Label htmlFor="dark-mode" className="text-sm font-medium">
                  Dark Mode
                </Label>
                <div className="text-muted-foreground text-xs">
                  {darkMode ? "Dark theme active" : "Light theme active"}
                </div>
              </div>
            </div>
            <Switch
              id="dark-mode"
              checked={darkMode}
              onCheckedChange={setDarkMode}
            />
          </div>
        </CardContent>
      </Card>
    );
  },
  parameters: {
    docs: {
      description: {
        story: "Device control panel with switches for common system settings.",
      },
    },
  },
};

// Form integration
export const FormIntegration: Story = {
  render: () => (
    <form className="w-[400px] space-y-6">
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Account Preferences</h3>

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Label htmlFor="newsletter" className="text-sm">
              Subscribe to newsletter
            </Label>
            <Switch id="newsletter" name="newsletter" defaultChecked />
          </div>

          <div className="flex items-center justify-between">
            <Label htmlFor="marketing" className="text-sm">
              Marketing emails
            </Label>
            <Switch id="marketing" name="marketing" />
          </div>

          <div className="flex items-center justify-between">
            <Label htmlFor="updates" className="text-sm">
              Product updates
            </Label>
            <Switch id="updates" name="updates" defaultChecked />
          </div>
        </div>

        <Separator />

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Label htmlFor="public-profile" className="text-sm">
              Make profile public
            </Label>
            <Switch id="public-profile" name="publicProfile" />
          </div>

          <div className="flex items-center justify-between">
            <Label htmlFor="online-status" className="text-sm">
              Show online status
            </Label>
            <Switch id="online-status" name="onlineStatus" defaultChecked />
          </div>
        </div>

        <Separator />

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <Label htmlFor="required-terms" className="text-sm">
                Accept terms and conditions *
              </Label>
              <div className="text-muted-foreground text-xs">
                Required to use the service
              </div>
            </div>
            <Switch id="required-terms" name="terms" required />
          </div>
        </div>
      </div>
    </form>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Switches integrated within a form with proper labels and validation.",
      },
    },
  },
};

// Sizes and styling
export const SizesAndStyling: Story = {
  render: () => (
    <div className="space-y-6">
      <div className="space-y-3">
        <h4 className="text-sm font-medium">Different Sizes</h4>
        <div className="space-y-2">
          <div className="flex items-center space-x-2">
            <Switch id="small" className="scale-75" />
            <Label htmlFor="small" className="text-sm">
              Small switch
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <Switch id="normal" />
            <Label htmlFor="normal" className="text-sm">
              Normal switch
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <Switch id="large" className="scale-125" />
            <Label htmlFor="large" className="text-sm">
              Large switch
            </Label>
          </div>
        </div>
      </div>

      <Separator />

      <div className="space-y-3">
        <h4 className="text-sm font-medium">Custom Colors</h4>
        <div className="space-y-2">
          <div className="flex items-center space-x-2">
            <Switch
              id="success"
              defaultChecked
              className="data-[state=checked]:bg-green-600"
            />
            <Label htmlFor="success" className="text-sm">
              Success color
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <Switch
              id="warning"
              defaultChecked
              className="data-[state=checked]:bg-yellow-600"
            />
            <Label htmlFor="warning" className="text-sm">
              Warning color
            </Label>
          </div>
          <div className="flex items-center space-x-2">
            <Switch
              id="danger"
              defaultChecked
              className="data-[state=checked]:bg-red-600"
            />
            <Label htmlFor="danger" className="text-sm">
              Danger color
            </Label>
          </div>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Switches with different sizes and custom color schemes.",
      },
    },
  },
};

// Interactive examples
export const InteractiveExamples: Story = {
  render: () => {
    const [sound, setSound] = useState(true);
    const [visibility, setVisibility] = useState(false);
    const [autoSave, setAutoSave] = useState(true);

    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            {sound ? (
              <Volume2 className="h-4 w-4" />
            ) : (
              <VolumeX className="h-4 w-4" />
            )}
            <Label htmlFor="sound-toggle" className="text-sm">
              Sound {sound ? "On" : "Off"}
            </Label>
          </div>
          <Switch
            id="sound-toggle"
            checked={sound}
            onCheckedChange={setSound}
          />
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            {visibility ? (
              <Eye className="h-4 w-4" />
            ) : (
              <EyeOff className="h-4 w-4" />
            )}
            <Label htmlFor="visibility-toggle" className="text-sm">
              Profile {visibility ? "Visible" : "Hidden"}
            </Label>
          </div>
          <Switch
            id="visibility-toggle"
            checked={visibility}
            onCheckedChange={setVisibility}
          />
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            {autoSave ? (
              <Shield className="h-4 w-4 text-green-500" />
            ) : (
              <Settings className="h-4 w-4" />
            )}
            <Label htmlFor="auto-save" className="text-sm">
              Auto-save {autoSave ? "Enabled" : "Disabled"}
            </Label>
          </div>
          <Switch
            id="auto-save"
            checked={autoSave}
            onCheckedChange={setAutoSave}
          />
        </div>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "Interactive switches that update their associated icons and text based on state.",
      },
    },
  },
};
