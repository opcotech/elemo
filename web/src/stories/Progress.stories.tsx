import type { Meta, StoryObj } from "@storybook/react-vite";
import { Download, Pause, Play, RotateCcw, Upload } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Spinner } from "@/components/ui/spinner";

const meta: Meta<typeof Progress> = {
  title: "UI/Progress",
  component: Progress,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Displays an indicator showing the completion progress of a task, typically displayed as a progress bar. Built on top of Radix UI Progress primitive.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    value: {
      control: { type: "range", min: 0, max: 100, step: 1 },
      description: "The progress value (0-100)",
    },
    max: {
      control: "number",
      description: "The maximum progress value",
    },
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the progress",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic progress
export const Default: Story = {
  args: {
    value: 60,
  },
};

// Different values
export const DifferentValues: Story = {
  render: () => (
    <div className="w-full space-y-6">
      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span>0%</span>
          <span>0/100</span>
        </div>
        <Progress value={0} />
      </div>

      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span>25%</span>
          <span>25/100</span>
        </div>
        <Progress value={25} />
      </div>

      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span>50%</span>
          <span>50/100</span>
        </div>
        <Progress value={50} />
      </div>

      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span>75%</span>
          <span>75/100</span>
        </div>
        <Progress value={75} />
      </div>

      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span>100%</span>
          <span>100/100</span>
        </div>
        <Progress value={100} />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Progress bars showing different completion values from 0% to 100%.",
      },
    },
  },
};

// Animated progress
export const Animated: Story = {
  render: () => {
    const [progress, setProgress] = useState(0);
    const [isRunning, setIsRunning] = useState(false);

    useEffect(() => {
      let interval: NodeJS.Timeout;

      if (isRunning && progress < 100) {
        interval = setInterval(() => {
          setProgress((prev) => {
            if (prev >= 100) {
              setIsRunning(false);
              return 100;
            }
            return prev + 1;
          });
        }, 50);
      }

      return () => clearInterval(interval);
    }, [isRunning, progress]);

    const handleStart = () => {
      setIsRunning(true);
    };

    const handlePause = () => {
      setIsRunning(false);
    };

    const handleReset = () => {
      setProgress(0);
      setIsRunning(false);
    };

    return (
      <div className="w-full space-y-4">
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span>Progress</span>
            <span>{progress}%</span>
          </div>
          <Progress value={progress} />
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={handleStart}
            disabled={isRunning || progress >= 100}
          >
            <Play className="h-4 w-4" />
            Start
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={handlePause}
            disabled={!isRunning}
          >
            <Pause className="h-4 w-4" />
            Pause
          </Button>
          <Button size="sm" variant="outline" onClick={handleReset}>
            <RotateCcw className="h-4 w-4" />
            Reset
          </Button>
        </div>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "An animated progress bar with start, pause, and reset controls.",
      },
    },
  },
};

// File upload progress
export const FileUpload: Story = {
  render: () => {
    const [uploadProgress, setUploadProgress] = useState(0);
    const [isUploading, setIsUploading] = useState(false);

    const simulateUpload = () => {
      setIsUploading(true);
      setUploadProgress(0);

      const interval = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev >= 100) {
            clearInterval(interval);
            setIsUploading(false);
            return 100;
          }
          return prev + Math.random() * 10;
        });
      }, 200);
    };

    return (
      <div className="w-full space-y-4">
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Upload className="h-4 w-4" />
            <span className="text-sm font-medium">Uploading document.pdf</span>
          </div>
          <Progress value={uploadProgress} />
          <div className="text-muted-foreground flex justify-between text-xs">
            <span>{uploadProgress.toFixed(1)}% complete</span>
            <span>
              {isUploading
                ? "Uploading..."
                : uploadProgress >= 100
                  ? "Complete"
                  : "Ready"}
            </span>
          </div>
        </div>
        <Button size="sm" onClick={simulateUpload} disabled={isUploading}>
          {isUploading ? (
            <>
              <Spinner size="xs" className="mr-2" />
              Uploading...
            </>
          ) : (
            <>
              <Upload className="h-4 w-4" />
              Start Upload
            </>
          )}
        </Button>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "Progress bar simulating file upload with realistic progress updates.",
      },
    },
  },
};

// Download progress
export const DownloadProgress: Story = {
  render: () => (
    <div className="w-full space-y-4">
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Download className="h-4 w-4" />
          <span className="text-sm font-medium">Downloading update.zip</span>
        </div>
        <Progress value={67} />
        <div className="text-muted-foreground flex justify-between text-xs">
          <span>67% • 1.2 MB of 1.8 MB</span>
          <span>2 min remaining</span>
        </div>
      </div>
    </div>
  ),
};

// Multiple progress bars
export const MultipleProgress: Story = {
  render: () => (
    <div className="w-full space-y-6">
      <div className="space-y-2">
        <Label className="text-sm font-medium">CPU Usage</Label>
        <Progress value={45} className="h-2" />
        <div className="text-muted-foreground text-xs">45% used</div>
      </div>

      <div className="space-y-2">
        <Label className="text-sm font-medium">Memory Usage</Label>
        <Progress value={78} className="h-2" />
        <div className="text-muted-foreground text-xs">
          7.8 GB of 10 GB used
        </div>
      </div>

      <div className="space-y-2">
        <Label className="text-sm font-medium">Disk Space</Label>
        <Progress value={23} className="h-2" />
        <div className="text-muted-foreground text-xs">
          115 GB of 500 GB used
        </div>
      </div>

      <div className="space-y-2">
        <Label className="text-sm font-medium">Network Usage</Label>
        <Progress value={89} className="h-2" />
        <div className="text-muted-foreground text-xs">
          89% of bandwidth used
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Multiple progress bars showing different system metrics.",
      },
    },
  },
};

// Custom colors
export const CustomColors: Story = {
  render: () => (
    <div className="w-full space-y-6">
      <div className="space-y-2">
        <Label className="text-sm">Default</Label>
        <Progress value={60} />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Success (Green)</Label>
        <Progress value={85} className="[&>div]:bg-green-500" />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Warning (Yellow)</Label>
        <Progress value={45} className="[&>div]:bg-yellow-500" />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Danger (Red)</Label>
        <Progress value={90} className="[&>div]:bg-red-500" />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Custom (Purple)</Label>
        <Progress value={70} className="[&>div]:bg-purple-500" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Progress bars with different color schemes for various states.",
      },
    },
  },
};

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="w-full space-y-6">
      <div className="space-y-2">
        <Label className="text-sm">Extra Small</Label>
        <Progress value={60} className="h-1" />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Small</Label>
        <Progress value={60} className="h-2" />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Default</Label>
        <Progress value={60} />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Large</Label>
        <Progress value={60} className="h-6" />
      </div>

      <div className="space-y-2">
        <Label className="text-sm">Extra Large</Label>
        <Progress value={60} className="h-8" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Progress bars in different heights/sizes.",
      },
    },
  },
};

// Indeterminate progress
export const Indeterminate: Story = {
  render: () => (
    <div className="w-full space-y-4">
      <div className="space-y-2">
        <Label className="text-sm font-medium">Loading...</Label>
        <div className="relative">
          <Progress value={0} className="overflow-hidden" />
          <div className="via-primary absolute inset-0 animate-pulse bg-gradient-to-r from-transparent to-transparent"></div>
        </div>
        <div className="text-muted-foreground text-xs">
          Please wait while we process your request
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "An indeterminate progress bar for when progress cannot be measured.",
      },
    },
  },
};

// Stepped progress
export const SteppedProgress: Story = {
  render: () => {
    const steps = ["Account", "Profile", "Preferences", "Review", "Complete"];
    const currentStep = 2;

    return (
      <div className="w-full space-y-4">
        <div className="flex justify-between text-sm">
          {steps.map((step, index) => (
            <div
              key={step}
              className={`text-center ${
                index <= currentStep
                  ? "text-primary font-medium"
                  : "text-muted-foreground"
              }`}
            >
              <div
                className={`mx-auto mb-1 flex h-8 w-8 items-center justify-center rounded-full border-2 ${
                  index < currentStep
                    ? "bg-primary border-primary text-primary-foreground"
                    : index === currentStep
                      ? "border-primary text-primary"
                      : "border-muted text-muted-foreground"
                }`}
              >
                {index + 1}
              </div>
              <div className="text-xs">{step}</div>
            </div>
          ))}
        </div>
        <Progress value={(currentStep / (steps.length - 1)) * 100} />
        <div className="text-muted-foreground text-center text-sm">
          Step {currentStep + 1} of {steps.length}: {steps[currentStep]}
        </div>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "A stepped progress indicator showing completion of multi-step processes.",
      },
    },
  },
};
