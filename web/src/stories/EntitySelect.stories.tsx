import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";

import { EntitySelect } from "@/components/ui/entity-select";
import { mockPeople, mockWorkItems } from "@/lib/mock-data";

const meta: Meta<typeof EntitySelect> = {
  title: "UI/EntitySelect",
  component: EntitySelect,
  parameters: {
    a11y: { test: "error" },
    layout: "centered",
    docs: {
      description: {
        component:
          "Accessible entity picker built on Select with title, description, and optional avatar identity.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const WorkItems: Story = {
  render: () => {
    const [value, setValue] = useState<string>(mockWorkItems[0].id);

    return (
      <div className="w-80">
        <EntitySelect
          aria-label="Work item"
          value={value}
          onValueChange={setValue}
          options={mockWorkItems.slice(0, 5).map((item) => ({
            value: item.id,
            title: `${item.key} ${item.title}`,
            description: item.summary,
          }))}
        />
      </div>
    );
  },
};

export const People: Story = {
  render: () => {
    const [value, setValue] = useState<string | undefined>(mockPeople[0]?.id);

    return (
      <div className="w-80">
        <EntitySelect
          aria-label="Assignee"
          value={value}
          onValueChange={setValue}
          placeholder="Select a person"
          options={mockPeople.slice(0, 5).map((person) => ({
            value: person.id,
            title: person.displayName,
            description: person.email,
            avatarFallback: person.displayName.slice(0, 2).toUpperCase(),
          }))}
        />
      </div>
    );
  },
};
