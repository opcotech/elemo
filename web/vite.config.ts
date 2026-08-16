/// <reference types="vitest/config" />
import path from "node:path";
import { fileURLToPath } from "node:url";

import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import viteReact from "@vitejs/plugin-react";
import { nitro } from "nitro/vite";
import type { Plugin } from "vite";
import { defineConfig } from "vite";

const root = fileURLToPath(new URL(".", import.meta.url));

function linkedomClientStub(): Plugin {
  return {
    name: "linkedom-client-stub",
    enforce: "pre",
    resolveId(source, _importer, options) {
      if (source === "linkedom" && !options.ssr) {
        return path.resolve(root, "src/lib/work/linkedom-browser-stub.ts");
      }
      return undefined;
    },
  };
}

export default defineConfig({
  server: {
    port: 3000,
  },
  resolve: {
    tsconfigPaths: true,
  },
  build: {
    chunkSizeWarningLimit: 1000,
  },
  ssr: {
    noExternal: ["dompurify", "linkedom"],
  },
  plugins: [
    linkedomClientStub(),
    tailwindcss(),
    tanstackStart(),
    nitro(),
    viteReact(),
  ],
});
