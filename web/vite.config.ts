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
const linkedomBrowserStub = path.resolve(
  root,
  "src/lib/work/linkedom-browser-stub.ts"
);
const linkedomCanvasShimId = "\0linkedom-canvas-shim";
const linkedomCanvasShimSource = `
class Canvas {
  constructor(width, height) {
    this.width = width;
    this.height = height;
  }
  getContext() { return null; }
  toDataURL() { return ""; }
}
export function createCanvas(width, height) {
  return new Canvas(width, height);
}
export default { createCanvas };
`;

function isLinkedomCanvasSpecifier(source: string, importer?: string): boolean {
  const sourcePath = source.replace(/\\/g, "/");
  const importerPath = importer?.replace(/\\/g, "/") ?? "";
  if (!sourcePath.endsWith("commonjs/canvas.cjs")) {
    return false;
  }
  return (
    importerPath.includes("/linkedom/") || sourcePath.includes("/linkedom/")
  );
}

function linkedomViteCompat(): Plugin {
  return {
    name: "linkedom-vite-compat",
    enforce: "pre",
    applyToEnvironment: () => true,
    configEnvironment(name) {
      if (name === "client") {
        return {};
      }
      return {
        resolve: {
          external: ["linkedom"],
        },
      };
    },
    resolveId(source, importer, options) {
      const isClient =
        this.environment?.name === "client" ||
        (this.environment?.name == null && !options.ssr);
      if (source === "linkedom" && isClient) {
        return linkedomBrowserStub;
      }
      if (isLinkedomCanvasSpecifier(source, importer)) {
        return linkedomCanvasShimId;
      }
      return undefined;
    },
    load(id) {
      return id === linkedomCanvasShimId ? linkedomCanvasShimSource : undefined;
    },
    transform(code, id) {
      const bare = id.replace(/\\/g, "/").split("?")[0];
      if (
        bare.endsWith("/linkedom/commonjs/canvas.cjs") &&
        code.includes("module.exports")
      ) {
        return { code: linkedomCanvasShimSource, map: null };
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
    noExternal: ["dompurify"],
    external: ["linkedom"],
  },
  plugins: [
    linkedomViteCompat(),
    tailwindcss(),
    tanstackStart(),
    nitro({
      routeRules: {
        "/assets/**": {
          headers: {
            "cache-control": "public, max-age=31536000, immutable",
          },
        },
      },
    }),
    viteReact(),
  ],
});
