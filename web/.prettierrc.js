export default {
  // Basic formatting
  semi: true,
  trailingComma: "es5",
  singleQuote: false,
  tabWidth: 2,
  useTabs: false,
  printWidth: 80,

  // JSX specific
  jsxSingleQuote: false,
  bracketSpacing: true,
  bracketSameLine: false,
  arrowParens: "always",

  // End of line
  endOfLine: "lf",

  // Tailwind CSS plugin for class sorting
  plugins: ["prettier-plugin-tailwindcss"],

  // File-specific overrides
  overrides: [
    {
      files: "*.md",
      options: {
        printWidth: 100,
        proseWrap: "preserve",
      },
    },
    {
      files: "*.json",
      options: {
        printWidth: 120,
        tabWidth: 2,
      },
    },
    {
      files: "*.{ts,tsx,js,jsx}",
      options: {
        parser: "typescript",
      },
    },
  ],
};
