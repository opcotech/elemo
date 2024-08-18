import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import type { DateRange } from "react-day-picker";

import { Calendar } from "@/components/ui/calendar";

const meta: Meta = {
  title: "UI/Calendar",
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A date picker component with range selection, multiple months, and customizable styling. Built on top of React Day Picker.",
      },
    },
  },
  tags: ["autodocs"],
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic single date selection
export const Default: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date());

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        className="rounded-md border"
      />
    );
  },
};

// Range selection
export const RangeSelection: Story = {
  render: () => {
    const [dateRange, setDateRange] = useState<DateRange | undefined>({
      from: new Date(),
      to: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000), // 7 days from now
    });

    return (
      <Calendar
        mode="range"
        selected={dateRange}
        onSelect={setDateRange}
        className="rounded-md border"
        numberOfMonths={2}
      />
    );
  },
};

// Multiple date selection
export const MultipleSelection: Story = {
  render: () => {
    const [dates, setDates] = useState<Date[]>([
      new Date(),
      new Date(Date.now() + 24 * 60 * 60 * 1000), // tomorrow
      new Date(Date.now() + 3 * 24 * 60 * 60 * 1000), // 3 days from now
    ]);

    return (
      <Calendar
        mode="multiple"
        selected={dates}
        onSelect={setDates}
        className="rounded-md border"
        required
      />
    );
  },
};

// With dropdown navigation
export const WithDropdownNavigation: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date());

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        captionLayout="dropdown"
        fromYear={2020}
        toYear={2030}
        className="rounded-md border"
      />
    );
  },
};

// Multiple months
export const MultipleMonths: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date());

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        numberOfMonths={3}
        className="rounded-md border"
      />
    );
  },
};

// Disabled dates
export const WithDisabledDates: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>();

    const disabledDays = [
      { before: new Date() }, // Disable past dates
      { dayOfWeek: [0, 6] }, // Disable weekends
    ];

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        disabled={disabledDays}
        className="rounded-md border"
      />
    );
  },
};

// Without outside days
export const WithoutOutsideDays: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date());

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        showOutsideDays={false}
        className="rounded-md border"
      />
    );
  },
};

// Week numbers
export const WithWeekNumbers: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date());

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        showWeekNumber
        className="rounded-md border"
      />
    );
  },
};

// Custom today date
export const CustomToday: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>();
    const customToday = new Date(2024, 11, 25); // Christmas 2024

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        today={customToday}
        className="rounded-md border"
      />
    );
  },
};

// Disabled calendar
export const Disabled: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date());

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        disabled
        className="rounded-md border opacity-50"
      />
    );
  },
};

// Different button variants
export const DifferentButtonVariant: Story = {
  render: () => {
    const [date, setDate] = useState<Date | undefined>(new Date());

    return (
      <Calendar
        mode="single"
        selected={date}
        onSelect={setDate}
        buttonVariant="outline"
        className="rounded-md border"
      />
    );
  },
};

// All features combined
export const FullFeatured: Story = {
  render: () => {
    const [dateRange, setDateRange] = useState<DateRange | undefined>({
      from: new Date(),
      to: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000),
    });

    const disabledDays = [
      { dayOfWeek: [0, 6] }, // Disable weekends
    ];

    return (
      <div className="space-y-4">
        <div className="text-muted-foreground text-sm">
          Range selection with dropdown navigation, week numbers, and disabled
          weekends
        </div>
        <Calendar
          mode="range"
          selected={dateRange}
          onSelect={setDateRange}
          captionLayout="dropdown"
          fromYear={2020}
          toYear={2030}
          showWeekNumber
          disabled={disabledDays}
          numberOfMonths={2}
          className="rounded-md border"
        />
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "A comprehensive example showcasing multiple calendar features combined.",
      },
    },
  },
};
