// For more info, see https://github.com/storybookjs/eslint-plugin-storybook#configuration-flat-config-format
import storybook from "eslint-plugin-storybook";

import { tanstackConfig } from "@tanstack/eslint-config";

export default [...tanstackConfig, {
  ignores: [
    "*.config.js",
    "*.config.ts",
    ".nitro/**",
    ".output/**",
    ".prettierrc.js",
    ".storybook/**",
    "build/**",
    "dist/**",
    "node_modules/**",
    "package.json",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
    "postcss.config.ts",
    "public/**",
    "src/lib/client/**",
    "storybook-static/**",
    "vite.config.ts",
  ],
}, {
  files: ["src/**/*.{js,jsx,ts,tsx}", "tests/**/*.{js,jsx,ts,tsx}"],
  languageOptions: {
    globals: {
      console: "readonly",
      window: "readonly",
      document: "readonly",
      navigator: "readonly",
      localStorage: "readonly",
      sessionStorage: "readonly",
      fetch: "readonly",
    },
  },
  rules: {
    // Core ESLint rules
    "no-unused-vars": "off",
    "@typescript-eslint/no-unused-vars": ["warn"],
    "react/react-in-jsx-scope": "off",
    "react/jsx-no-bind": "off",

    // Import order
    "import/order": [
      "error",
      {
        groups: [
          "builtin",
          "external",
          "internal",
          ["parent", "sibling", "index"],
        ],
        alphabetize: { order: "asc", caseInsensitive: true },
        "newlines-between": "always",
      },
    ],

    // TypeScript specific rules
    "@typescript-eslint/consistent-type-imports": [
      "error",
      {
        prefer: "type-imports",
        disallowTypeAnnotations: false,
      },
    ],

    // Disable some strict rules that may be too aggressive
    "@typescript-eslint/no-unnecessary-condition": "off",
    "@typescript-eslint/no-unnecessary-type-assertion": "off",
    "no-shadow": "warn",
    "@typescript-eslint/array-type": ["error", { default: "array" }],
  },
}, ...storybook.configs["flat/recommended"]];
