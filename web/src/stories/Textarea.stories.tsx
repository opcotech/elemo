import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  AlertCircle,
  Check,
  Info,
  MessageCircle,
  Save,
  Send,
  X,
} from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

const meta: Meta<typeof Textarea> = {
  title: "UI/Textarea",
  component: Textarea,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A multi-line text input component for longer text content. Built with modern styling and accessibility features.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    placeholder: {
      control: "text",
      description: "Placeholder text for the textarea",
    },
    disabled: {
      control: "boolean",
      description: "Whether the textarea is disabled",
    },
    rows: {
      control: "number",
      description: "Number of visible text lines",
    },
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the textarea",
    },
    value: {
      control: "text",
      description: "The controlled value of the textarea",
    },
    defaultValue: {
      control: "text",
      description: "The default value when uncontrolled",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic textarea
export const Default: Story = {
  args: {
    placeholder: "Type your message here...",
  },
};

// With label
export const WithLabel: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="message">Message</Label>
      <Textarea id="message" placeholder="Type your message here..." />
    </div>
  ),
};

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="w-full max-w-md space-y-4">
      <div className="grid gap-1.5">
        <Label>Small (2 rows)</Label>
        <Textarea placeholder="Small textarea..." rows={2} />
      </div>
      <div className="grid gap-1.5">
        <Label>Medium (4 rows)</Label>
        <Textarea placeholder="Medium textarea..." rows={4} />
      </div>
      <div className="grid gap-1.5">
        <Label>Large (6 rows)</Label>
        <Textarea placeholder="Large textarea..." rows={6} />
      </div>
      <div className="grid gap-1.5">
        <Label>Extra Large (8 rows)</Label>
        <Textarea placeholder="Extra large textarea..." rows={8} />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Textareas in different sizes by adjusting the number of rows.",
      },
    },
  },
};

// Disabled state
export const Disabled: Story = {
  render: () => (
    <div className="grid w-full max-w-sm items-center gap-1.5">
      <Label htmlFor="disabled">Disabled</Label>
      <Textarea
        id="disabled"
        placeholder="This textarea is disabled"
        disabled
        defaultValue="This is a disabled textarea with some default content."
      />
    </div>
  ),
};

// With character count
export const WithCharacterCount: Story = {
  render: () => {
    const [text, setText] = useState("");
    const maxLength = 280;

    return (
      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="tweet">Tweet</Label>
        <Textarea
          id="tweet"
          placeholder="What's happening?"
          value={text}
          onChange={(e) => setText(e.target.value)}
          maxLength={maxLength}
          rows={3}
        />
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">Share your thoughts</span>
          <span
            className={`${text.length > maxLength * 0.9 ? "text-red-500" : "text-muted-foreground"}`}
          >
            {text.length}/{maxLength}
          </span>
        </div>
      </div>
    );
  },
};

// Form validation states
export const ValidationStates: Story = {
  render: () => (
    <div className="w-full max-w-sm space-y-4">
      <div className="grid gap-1.5">
        <Label htmlFor="success">Success</Label>
        <Textarea
          id="success"
          placeholder="Valid input..."
          className="border-green-500 focus:border-green-500 focus:ring-green-500"
          defaultValue="This is a valid message."
        />
        <div className="flex items-center rounded bg-green-600 px-2 py-1 text-sm text-white">
          <Check className="mr-1 h-4 w-4 text-white" />
          Message looks good!
        </div>
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="warning">Warning</Label>
        <Textarea
          id="warning"
          placeholder="Warning input..."
          className="border-yellow-500 focus:border-yellow-500 focus:ring-yellow-500"
          defaultValue="This message might need attention."
        />
        <div className="flex items-center text-sm text-yellow-600">
          <AlertCircle className="mr-1 h-4 w-4" />
          Please review your message.
        </div>
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="error">Error</Label>
        <Textarea
          id="error"
          placeholder="Invalid input..."
          className="border-red-500 focus:border-red-500 focus:ring-red-500"
          defaultValue="This message has an error."
        />
        <div className="flex items-center text-sm text-red-600">
          <X className="mr-1 h-4 w-4" />
          Message is required and must be at least 10 characters.
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Textareas showing different validation states with appropriate styling and feedback messages.",
      },
    },
  },
};

// Comment form
export const CommentForm: Story = {
  render: () => {
    const [comment, setComment] = useState("");

    return (
      <div className="w-full max-w-md space-y-4">
        <div className="space-y-2">
          <Label htmlFor="comment">Leave a comment</Label>
          <Textarea
            id="comment"
            placeholder="Share your thoughts..."
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            rows={4}
          />
          <div className="text-muted-foreground text-xs">
            Be respectful and constructive in your feedback.
          </div>
        </div>
        <div className="flex items-center justify-between">
          <div className="text-muted-foreground text-sm">
            {comment.length} characters
          </div>
          <div className="flex space-x-2">
            <Button variant="outline" size="sm">
              Cancel
            </Button>
            <Button size="sm" disabled={comment.length === 0}>
              <MessageCircle className="mr-2 h-4 w-4" />
              Post Comment
            </Button>
          </div>
        </div>
      </div>
    );
  },
};

// Message composer
export const MessageComposer: Story = {
  render: () => {
    const [message, setMessage] = useState("");

    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Send className="h-5 w-5" />
            Send Message
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="to">To</Label>
            <input
              id="to"
              className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:ring focus:ring-offset-2 focus:outline-none"
              placeholder="recipient@example.com"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="subject">Subject</Label>
            <input
              id="subject"
              className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:ring focus:ring-offset-2 focus:outline-none"
              placeholder="Message subject"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="message-body">Message</Label>
            <Textarea
              id="message-body"
              placeholder="Type your message here..."
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              rows={5}
            />
          </div>
          <div className="flex items-center justify-between">
            <Badge variant="secondary" className="text-xs">
              {message.length > 0 ? `${message.length} characters` : "Empty"}
            </Badge>
            <div className="flex space-x-2">
              <Button variant="outline" size="sm">
                Save Draft
              </Button>
              <Button size="sm" disabled={message.length === 0}>
                <Send className="mr-2 h-4 w-4" />
                Send
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  },
};

// Feedback form
export const FeedbackForm: Story = {
  render: () => {
    const [feedback, setFeedback] = useState("");
    const [category, setCategory] = useState("");

    return (
      <div className="w-full max-w-md space-y-4">
        <div className="space-y-2">
          <Label htmlFor="feedback-category">Category</Label>
          <select
            id="feedback-category"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="border-input bg-background ring-offset-background focus:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:ring focus:ring-offset-2 focus:outline-none"
          >
            <option value="">Select a category</option>
            <option value="bug">Bug Report</option>
            <option value="feature">Feature Request</option>
            <option value="improvement">Improvement</option>
            <option value="other">Other</option>
          </select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="feedback-text">Feedback</Label>
          <Textarea
            id="feedback-text"
            placeholder="Please describe your feedback in detail..."
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            rows={6}
          />
          <div className="text-muted-foreground flex items-center text-xs">
            <Info className="mr-1 h-3 w-3" />
            The more details you provide, the better we can help you.
          </div>
        </div>
        <Button
          className="w-full"
          disabled={feedback.length === 0 || category === ""}
        >
          Submit Feedback
        </Button>
      </div>
    );
  },
};

// Auto-resizing textarea
export const AutoResizing: Story = {
  render: () => {
    const [value, setValue] = useState("");

    return (
      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="auto-resize">Auto-resizing Textarea</Label>
        <Textarea
          id="auto-resize"
          placeholder="Start typing and watch this textarea grow..."
          value={value}
          onChange={(e) => setValue(e.target.value)}
          style={{
            minHeight: "80px",
            height: Math.max(80, value.split("\n").length * 24 + 32) + "px",
          }}
          className="resize-none overflow-hidden"
        />
        <div className="text-muted-foreground text-xs">
          This textarea automatically adjusts its height based on content.
        </div>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "A textarea that automatically adjusts its height based on the content length.",
      },
    },
  },
};

// Rich text preview
export const WithPreview: Story = {
  render: () => {
    const [activeTab, setActiveTab] = useState<"write" | "preview">("write");
    const [content, setContent] = useState(
      "# Hello World\n\nThis is **bold** text and this is *italic* text.\n\n- List item 1\n- List item 2\n- List item 3"
    );

    return (
      <div className="w-full max-w-md space-y-4">
        <div className="flex space-x-1 border-b">
          <button
            onClick={() => setActiveTab("write")}
            className={`border-b-2 px-3 py-2 text-sm font-medium ${
              activeTab === "write"
                ? "border-primary text-primary"
                : "text-muted-foreground hover:text-foreground border-transparent"
            }`}
          >
            Write
          </button>
          <button
            onClick={() => setActiveTab("preview")}
            className={`border-b-2 px-3 py-2 text-sm font-medium ${
              activeTab === "preview"
                ? "border-primary text-primary"
                : "text-muted-foreground hover:text-foreground border-transparent"
            }`}
          >
            Preview
          </button>
        </div>

        {activeTab === "write" ? (
          <Textarea
            placeholder="Write your markdown here..."
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={8}
            className="font-mono text-sm"
          />
        ) : (
          <div className="min-h-[200px] rounded-md border p-3 text-sm">
            <div className="space-y-2">
              <h1 className="text-xl font-bold">Hello World</h1>
              <p>
                This is <strong>bold</strong> text and this is <em>italic</em>{" "}
                text.
              </p>
              <ul className="list-inside list-disc space-y-1">
                <li>List item 1</li>
                <li>List item 2</li>
                <li>List item 3</li>
              </ul>
            </div>
          </div>
        )}

        <div className="flex items-center justify-between">
          <div className="text-muted-foreground text-xs">
            Supports Markdown formatting
          </div>
          <Button size="sm">
            <Save className="mr-2 h-4 w-4" />
            Save
          </Button>
        </div>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "A textarea with write and preview tabs, useful for markdown editors.",
      },
    },
  },
};

// Custom styling
export const CustomStyling: Story = {
  render: () => (
    <div className="w-full max-w-sm space-y-4">
      <div className="grid gap-1.5">
        <Label>Minimal Style</Label>
        <Textarea
          placeholder="Minimal textarea..."
          className="border-muted focus:border-primary rounded-none border-0 border-b-2 focus:ring-0"
          rows={3}
        />
      </div>

      <div className="grid gap-1.5">
        <Label>Rounded Style</Label>
        <Textarea
          placeholder="Rounded textarea..."
          className="rounded-xl border-2"
          rows={3}
        />
      </div>

      <div className="grid gap-1.5">
        <Label>Colored Background</Label>
        <Textarea
          placeholder="Colored background..."
          className="border-blue-700 bg-blue-600 text-white focus:border-blue-800 focus:ring-blue-800"
          rows={3}
        />
      </div>

      <div className="grid gap-1.5">
        <Label>No Resize</Label>
        <Textarea
          placeholder="No resize handles..."
          className="resize-none"
          rows={3}
        />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Textareas with various custom styling options.",
      },
    },
  },
};
