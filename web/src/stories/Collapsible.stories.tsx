import type { Meta, StoryObj } from "@storybook/react-vite";
import { ChevronsUpDown } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

const meta: Meta<typeof Collapsible> = {
  title: "UI/Collapsible",
  component: Collapsible,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "An expandable/collapsible region built on Base UI Collapsible.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => {
    const [open, setOpen] = useState(false);

    return (
      <Collapsible
        open={open}
        onOpenChange={setOpen}
        className="w-[350px] space-y-2"
      >
        <div className="flex items-center justify-between gap-4">
          <h4 className="text-sm font-semibold">3 starred repositories</h4>
          <CollapsibleTrigger
            render={<Button variant="ghost" size="icon-sm" />}
          >
            <ChevronsUpDown className="h-4 w-4" />
            <span className="sr-only">Toggle</span>
          </CollapsibleTrigger>
        </div>
        <div className="rounded-lg border px-4 py-2 text-sm">@elemo/core</div>
        <CollapsibleContent className="space-y-2">
          <div className="rounded-lg border px-4 py-2 text-sm">@elemo/web</div>
          <div className="rounded-lg border px-4 py-2 text-sm">@elemo/api</div>
        </CollapsibleContent>
      </Collapsible>
    );
  },
};
