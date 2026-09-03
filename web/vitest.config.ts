import { fileURLToPath } from "node:url";

import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
      "@elemo/plugin-sdk": fileURLToPath(
        new URL("./packages/plugin-sdk/src/index.ts", import.meta.url)
      ),
      "@elemo/plugin-ui": fileURLToPath(
        new URL("./packages/plugin-ui/src/index.ts", import.meta.url)
      ),
    },
  },
  test: {
    include: [
      "src/**/*.{test,spec}.{ts,tsx}",
      "tests/e2e/utils/**/*.test.ts",
    ],
  },
});
