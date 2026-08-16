import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";

import { DatePicker } from "@/components/ui/date-picker";
import { Label } from "@/components/ui/label";

const meta: Meta<typeof DatePicker> = {
  title: "UI/DatePicker",
  component: DatePicker,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Date picker combining a popover trigger button and calendar.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => {
    const [date, setDate] = useState<Date | null>(null);

    return (
      <div className="w-[280px] space-y-2">
        <Label>Pick a date</Label>
        <DatePicker date={date} onDateChange={setDate} />
      </div>
    );
  },
};

export const WithValue: Story = {
  render: () => {
    const [date, setDate] = useState<Date | null>(new Date());

    return (
      <div className="w-[280px] space-y-2">
        <Label>Event date</Label>
        <DatePicker date={date} onDateChange={setDate} />
      </div>
    );
  },
};

export const Disabled: Story = {
  render: () => (
    <div className="w-[280px] space-y-2">
      <Label>Disabled</Label>
      <DatePicker disabled placeholder="Unavailable" />
    </div>
  ),
};

export const Clearable: Story = {
  render: () => {
    const [date, setDate] = useState<Date | null>(new Date());

    return (
      <div className="w-[280px] space-y-2">
        <Label>Due date</Label>
        <DatePicker
          clearable
          date={date}
          onDateChange={setDate}
          placeholder="Due date"
        />
      </div>
    );
  },
};
