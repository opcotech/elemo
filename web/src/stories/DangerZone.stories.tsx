import type { Meta, StoryObj } from "@storybook/react-vite";
import { Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DangerZone,
  DangerZoneActions,
  DangerZoneContent,
  DangerZoneDescription,
  DangerZoneHeader,
  DangerZoneTitle,
} from "@/components/ui/danger-zone";

const meta: Meta<typeof DangerZone> = {
  title: "UI/DangerZone",
  component: DangerZone,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A bordered destructive surface for irreversible settings actions. Use instead of Card with a destructive border override.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <DangerZone className="w-[420px]" data-section="example-danger-zone">
      <DangerZoneHeader>
        <DangerZoneTitle>Danger Zone</DangerZoneTitle>
        <DangerZoneDescription>
          Irreversible actions for this resource
        </DangerZoneDescription>
      </DangerZoneHeader>
      <DangerZoneContent className="space-y-2">
        <p className="text-muted-foreground text-sm">
          Deleting this resource permanently removes it. This action cannot be
          undone.
        </p>
        <p className="text-sm font-medium">Consequences:</p>
        <ul className="text-muted-foreground list-inside list-disc space-y-1 text-sm">
          <li>All related access will be revoked</li>
          <li>Associated data will no longer be linked</li>
          <li>This action is permanent and cannot be reversed</li>
        </ul>
      </DangerZoneContent>
      <DangerZoneActions>
        <Button variant="destructive">
          <Trash2 className="size-4" />
          Delete resource
        </Button>
      </DangerZoneActions>
    </DangerZone>
  ),
};
