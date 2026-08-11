import type { Preview } from "@storybook/react-vite";
import { withThemeByClassName } from "@storybook/addon-themes";
import "../src/styles/app.css";

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: "^on[A-Z].*" },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    docs: {
      toc: true,
    },
    a11y: {
      // Failures are enforced on the maintained UI/Elemo catalog via
      // parameters.a11y.test = "error" in those story metas. Legacy stories
      // stay on "todo" so violations remain visible without masking.
      test: "todo",
      disable: false,
      manual: false,
      config: {
        rules: [
          { id: "autocomplete-valid", enabled: true },
          { id: "button-name", enabled: true },
          { id: "color-contrast", enabled: true },
          { id: "focus-order-semantics", enabled: true },
          { id: "form-field-multiple-labels", enabled: true },
          { id: "frame-title", enabled: true },
          { id: "image-alt", enabled: true },
          { id: "input-image-alt", enabled: true },
          { id: "label", enabled: true },
          { id: "link-name", enabled: true },
          { id: "aria-valid-attr", enabled: true },
          { id: "aria-valid-attr-value", enabled: true },
          { id: "aria-roles", enabled: true },
          { id: "tabindex", enabled: true },
          { id: "duplicate-id", enabled: true },
          { id: "heading-order", enabled: true },
          { id: "landmark-unique", enabled: true },
          { id: "list", enabled: true },
          { id: "listitem", enabled: true },
          { id: "region", enabled: true },
        ],
      },
      options: {
        checks: {
          "color-contrast": { options: { noScroll: true } },
        },
        restoreScroll: true,
        runOnly: {
          type: "tag",
          values: ["wcag2a", "wcag2aa", "wcag21aa", "best-practice"],
        },
      },
      context: "#storybook-root",
    },
    backgrounds: {
      disable: true,
    },
    layout: "centered",
  },
  decorators: [
    withThemeByClassName({
      themes: {
        light: "",
        dark: "dark",
      },
      defaultTheme: "light",
    }),
    (Story) => (
      <div className="bg-background text-foreground min-h-50 p-4">
        <Story />
      </div>
    ),
  ],
  globalTypes: {
    theme: {
      description: "Global theme for components",
      defaultValue: "light",
      toolbar: {
        title: "Theme",
        icon: "paintbrush",
        items: ["light", "dark"],
        dynamicTitle: true,
      },
    },
  },
};

export default preview;
