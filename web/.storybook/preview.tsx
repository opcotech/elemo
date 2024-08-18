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
      // Enable accessibility testing globally for all stories
      disable: false,
      // Run a11y tests automatically (not manual)
      manual: false,
      // Configuration for @storybook/addon-a11y
      config: {
        rules: [
          // Essential accessibility rules
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
          "duplicate-id": { options: { allowFailure: true } },
        },
        restoreScroll: true,
        runOnly: {
          type: "tag",
          values: ["wcag2a", "wcag2aa", "wcag21aa", "best-practice"],
        },
      },
      // Context to test (replaces deprecated element parameter)
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
      <div
        className="min-h-[200px] bg-white p-4"
        role="main"
        aria-label="Story content"
      >
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
