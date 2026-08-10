import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";

const meta: Meta<typeof DeleteConfirmationDialog> = {
  title: "UI/DeleteConfirmationDialog",
  component: DeleteConfirmationDialog,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Reusable destructive confirmation dialog built on AlertDialog.",
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
      <>
        <Button variant="destructive" onClick={() => setOpen(true)}>
          Delete project
        </Button>
        <DeleteConfirmationDialog
          open={open}
          onOpenChange={setOpen}
          title="Delete project?"
          description="This will permanently delete the project and all associated data."
          consequences={[
            "All issues and documents will be removed",
            "Team members will lose access",
            "This action cannot be undone",
          ]}
          onConfirm={() => setOpen(false)}
        />
      </>
    );
  },
};

export const Pending: Story = {
  render: () => (
    <DeleteConfirmationDialog
      open
      onOpenChange={() => undefined}
      title="Deleting..."
      description="Please wait while we delete this resource."
      onConfirm={() => undefined}
      isPending
    />
  ),
};
