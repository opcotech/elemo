import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";

import { SearchInput } from "@/components/ui/search-input";

const meta: Meta<typeof SearchInput> = {
  title: "UI/SearchInput",
  component: SearchInput,
  parameters: {
    a11y: { test: "error" },
    layout: "centered",
    docs: {
      description: {
        component:
          "Search field with a leading decorative icon and an accessible name derived from the placeholder or aria-label.",
      },
    },
  },
  tags: ["autodocs", "verification"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => {
    const [value, setValue] = useState("");

    return (
      <SearchInput
        value={value}
        onChange={setValue}
        placeholder="Search projects..."
        className="w-[320px]"
      />
    );
  },
};

export const WithValue: Story = {
  render: () => {
    const [value, setValue] = useState("elemo");

    return (
      <SearchInput
        value={value}
        onChange={setValue}
        placeholder="Search..."
        className="w-[320px]"
      />
    );
  },
};

export const Disabled: Story = {
  render: () => (
    <SearchInput
      value=""
      onChange={() => undefined}
      placeholder="Search disabled"
      disabled
      className="w-[320px]"
    />
  ),
};
