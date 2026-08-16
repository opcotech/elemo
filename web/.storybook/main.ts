import path from "node:path";
import { fileURLToPath } from "node:url";

import type { StorybookConfig } from "@storybook/react-vite";
import { mergeConfig } from "vite";
import type { PluginOption } from "vite";
import tailwindcss from "@tailwindcss/vite";
import viteReact from "@vitejs/plugin-react";

const dirname = path.dirname(fileURLToPath(import.meta.url));

function isAppOnlyPlugin(plugin: PluginOption): boolean {
  if (!plugin || typeof plugin !== "object") {
    return false;
  }
  if (Array.isArray(plugin)) {
    return plugin.some(isAppOnlyPlugin);
  }
  const name = "name" in plugin ? String(plugin.name ?? "") : "";
  return /tanstack|nitro|start-manifest|server-fn/i.test(name);
}

const config: StorybookConfig = {
  stories: ["../src/**/*.mdx", "../src/**/*.stories.@(js|jsx|mjs|ts|tsx)"],
  addons: [
    "@storybook/addon-docs",
    "@storybook/addon-a11y",
    "@storybook/addon-themes",
  ],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  typescript: {
    check: false,
    reactDocgen: false,
  },
  async viteFinal(config) {
    const filteredPlugins = (config.plugins ?? []).filter(
      (plugin) => !isAppOnlyPlugin(plugin)
    );

    return mergeConfig(
      {
        ...config,
        // Do not inherit the app Vite config (TanStack Start / Nitro).
        configFile: false,
        plugins: filteredPlugins,
      },
      {
        plugins: [tailwindcss(), viteReact()],
        resolve: {
          tsconfigPaths: true,
          alias: [
            {
              // String aliases for the package root append subpaths onto the
              // stub file (`tanstack-start-stub.ts/server-only`).
              find: /^@tanstack\/react-start(?:\/.*)?$/,
              replacement: path.resolve(dirname, "./tanstack-start-stub.ts"),
            },
            {
              find: "@",
              replacement: path.resolve(dirname, "../src"),
            },
          ],
        },
      }
    );
  },
};

export default config;
