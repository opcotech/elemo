import type { Meta, StoryObj } from "@storybook/react-vite";
import { BreadcrumbNav } from "@/components/breadcrumb";
import { withRouter } from "../../.storybook/with-router";

const meta: Meta<typeof BreadcrumbNav> = {
  title: "Components/Breadcrumb",
  component: BreadcrumbNav,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Breadcrumb navigation driven by route `staticData.breadcrumb` values. Home is prepended automatically when missing.",
      },
    },
  },
  tags: ["autodocs"],
  decorators: [withRouter],
};

export default meta;
type Story = StoryObj<typeof BreadcrumbNav>;

/** Default story at `/`: BreadcrumbNav renders a Home crumb from the current path. */
export const Home: Story = {
  render: () => <BreadcrumbNav />,
};
