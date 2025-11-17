import type { Meta, StoryObj } from "@storybook/react-vite";
import type { ComponentProps } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
} from "recharts";
import type { PieLabelRenderProps } from "recharts";

import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";

const meta: Meta<typeof ChartContainer> = {
  title: "UI/Chart",
  component: ChartContainer,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A set of chart components built on top of Recharts with consistent styling and theming support.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    config: {
      description: "Chart configuration object for styling and data mapping",
    },
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the chart container",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Sample data
const barData = [
  { name: "Jan", desktop: 186, mobile: 80 },
  { name: "Feb", desktop: 305, mobile: 200 },
  { name: "Mar", desktop: 237, mobile: 120 },
  { name: "Apr", desktop: 73, mobile: 190 },
  { name: "May", desktop: 209, mobile: 130 },
  { name: "Jun", desktop: 214, mobile: 140 },
];

const lineData = [
  { name: "Jan", visitors: 4000, pageViews: 2400 },
  { name: "Feb", visitors: 3000, pageViews: 1398 },
  { name: "Mar", visitors: 2000, pageViews: 9800 },
  { name: "Apr", visitors: 2780, pageViews: 3908 },
  { name: "May", visitors: 1890, pageViews: 4800 },
  { name: "Jun", visitors: 2390, pageViews: 3800 },
];

const pieData = [
  { name: "Desktop", value: 400, fill: "#8884d8" },
  { name: "Mobile", value: 300, fill: "#82ca9d" },
  { name: "Tablet", value: 200, fill: "#ffc658" },
  { name: "Other", value: 100, fill: "#ff7c7c" },
];

const areaData = [
  { name: "Jan", revenue: 4000, profit: 2400 },
  { name: "Feb", revenue: 3000, profit: 1398 },
  { name: "Mar", revenue: 2000, profit: 9800 },
  { name: "Apr", revenue: 2780, profit: 3908 },
  { name: "May", revenue: 1890, profit: 4800 },
  { name: "Jun", revenue: 2390, profit: 3800 },
];

const chartConfig = {
  desktop: {
    label: "Desktop",
    color: "#8884d8",
  },
  mobile: {
    label: "Mobile",
    color: "#82ca9d",
  },
  visitors: {
    label: "Visitors",
    color: "#8884d8",
  },
  pageViews: {
    label: "Page Views",
    color: "#82ca9d",
  },
  revenue: {
    label: "Revenue",
    color: "#8884d8",
  },
  profit: {
    label: "Profit",
    color: "#82ca9d",
  },
};

const renderTooltip = (props: ComponentProps<typeof ChartTooltipContent>) => (
  <ChartTooltipContent {...props} />
);

// Bar Chart
export const BarChartExample: Story = {
  render: () => (
    <ChartContainer config={chartConfig} className="min-h-[200px] w-full">
      <BarChart data={barData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="name" />
        <YAxis />
        <ChartTooltip content={renderTooltip} />
        <ChartLegend content={<ChartLegendContent />} />
        <Bar dataKey="desktop" fill="var(--color-desktop)" />
        <Bar dataKey="mobile" fill="var(--color-mobile)" />
      </BarChart>
    </ChartContainer>
  ),
};

// Line Chart
export const LineChartExample: Story = {
  render: () => (
    <ChartContainer config={chartConfig} className="min-h-[200px] w-full">
      <LineChart data={lineData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="name" />
        <YAxis />
        <ChartTooltip content={renderTooltip} />
        <ChartLegend content={<ChartLegendContent />} />
        <Line
          type="monotone"
          dataKey="visitors"
          stroke="var(--color-visitors)"
          strokeWidth={2}
        />
        <Line
          type="monotone"
          dataKey="pageViews"
          stroke="var(--color-pageViews)"
          strokeWidth={2}
        />
      </LineChart>
    </ChartContainer>
  ),
};

// Pie Chart
export const PieChartExample: Story = {
  render: () => (
    <ChartContainer config={chartConfig} className="min-h-[200px] w-full">
      <PieChart>
        <ChartTooltip content={renderTooltip} />
        <Pie
          data={pieData}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={({ name, percent }: PieLabelRenderProps) => {
            const percentValue =
              typeof percent === "number" ? percent : undefined;
            return percentValue !== undefined
              ? `${String(name)} ${(percentValue * 100).toFixed(0)}%`
              : String(name);
          }}
          outerRadius={80}
          fill="#8884d8"
          dataKey="value"
        >
          {pieData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.fill} />
          ))}
        </Pie>
      </PieChart>
    </ChartContainer>
  ),
};

// Area Chart
export const AreaChartExample: Story = {
  render: () => (
    <ChartContainer config={chartConfig} className="min-h-[200px] w-full">
      <AreaChart data={areaData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="name" />
        <YAxis />
        <ChartTooltip content={renderTooltip} />
        <ChartLegend content={<ChartLegendContent />} />
        <Area
          type="monotone"
          dataKey="revenue"
          stackId="1"
          stroke="var(--color-revenue)"
          fill="var(--color-revenue)"
        />
        <Area
          type="monotone"
          dataKey="profit"
          stackId="1"
          stroke="var(--color-profit)"
          fill="var(--color-profit)"
        />
      </AreaChart>
    </ChartContainer>
  ),
};

// Simple Bar Chart
export const SimpleBarChart: Story = {
  render: () => (
    <ChartContainer config={chartConfig} className="min-h-[200px] w-full">
      <BarChart data={barData}>
        <XAxis dataKey="name" />
        <YAxis />
        <ChartTooltip content={renderTooltip} />
        <Bar dataKey="desktop" fill="var(--color-desktop)" />
      </BarChart>
    </ChartContainer>
  ),
};

// Responsive Chart
export const ResponsiveChart: Story = {
  render: () => (
    <div className="h-64 w-full">
      <ChartContainer config={chartConfig} className="h-full w-full">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={lineData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" />
            <YAxis />
            <ChartTooltip content={renderTooltip} />
            <Line
              type="monotone"
              dataKey="visitors"
              stroke="var(--color-visitors)"
              strokeWidth={3}
              dot={{ r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </ChartContainer>
    </div>
  ),
};

// Multiple Chart Types
export const MultipleCharts: Story = {
  render: () => (
    <div className="grid w-full grid-cols-1 gap-6 md:grid-cols-2">
      <ChartContainer config={chartConfig} className="min-h-[200px]">
        <BarChart data={barData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" />
          <YAxis />
          <ChartTooltip content={renderTooltip} />
          <Bar dataKey="desktop" fill="var(--color-desktop)" />
        </BarChart>
      </ChartContainer>

      <ChartContainer config={chartConfig} className="min-h-[200px]">
        <LineChart data={lineData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" />
          <YAxis />
          <ChartTooltip content={renderTooltip} />
          <Line
            type="monotone"
            dataKey="visitors"
            stroke="var(--color-visitors)"
            strokeWidth={2}
          />
        </LineChart>
      </ChartContainer>

      <ChartContainer config={chartConfig} className="min-h-[200px]">
        <PieChart>
          <ChartTooltip content={renderTooltip} />
          <Pie
            data={pieData}
            cx="50%"
            cy="50%"
            outerRadius={60}
            fill="#8884d8"
            dataKey="value"
          >
            {pieData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.fill} />
            ))}
          </Pie>
        </PieChart>
      </ChartContainer>

      <ChartContainer config={chartConfig} className="min-h-[200px]">
        <AreaChart data={areaData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" />
          <YAxis />
          <ChartTooltip content={renderTooltip} />
          <Area
            type="monotone"
            dataKey="revenue"
            stroke="var(--color-revenue)"
            fill="var(--color-revenue)"
            fillOpacity={0.6}
          />
        </AreaChart>
      </ChartContainer>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Multiple chart types displayed in a grid layout.",
      },
    },
  },
};

// Custom Styled Chart
export const CustomStyled: Story = {
  render: () => (
    <ChartContainer
      config={chartConfig}
      className="min-h-[300px] w-full rounded-lg border p-4"
    >
      <BarChart data={barData}>
        <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
        <XAxis
          dataKey="name"
          tick={{ fontSize: 12 }}
          axisLine={{ stroke: "#374151" }}
        />
        <YAxis tick={{ fontSize: 12 }} axisLine={{ stroke: "#374151" }} />
        <ChartTooltip
          content={renderTooltip}
          cursor={{ fill: "rgba(59, 130, 246, 0.1)" }}
        />
        <ChartLegend content={<ChartLegendContent />} />
        <Bar
          dataKey="desktop"
          fill="var(--color-desktop)"
          radius={[4, 4, 0, 0]}
        />
        <Bar
          dataKey="mobile"
          fill="var(--color-mobile)"
          radius={[4, 4, 0, 0]}
        />
      </BarChart>
    </ChartContainer>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A customized chart with rounded bars, custom styling, and enhanced visual appeal.",
      },
    },
  },
};
