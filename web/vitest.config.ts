import { fileURLToPath } from "node:url";

import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  test: {
    include: [
      "src/**/*.{test,spec}.{ts,tsx}",
      "tests/e2e/utils/**/*.test.ts",
    ],
  },
});
