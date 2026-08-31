import path from "node:path";
import { fileURLToPath } from "node:url";

import { defineConfig } from "vite";

const dir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  mode: "production",
  publicDir: false,
  build: {
    lib: {
      entry: path.join(dir, "src/index.tsx"),
      formats: ["es"],
      fileName: () => "index.js",
    },
    outDir: dir,
    emptyOutDir: false,
    minify: false,
    rollupOptions: {
      external: [
        "react",
        "react-dom",
        "react/jsx-runtime",
        "react/jsx-dev-runtime",
        "@elemo/plugin-sdk",
        "@elemo/plugin-ui",
        "lucide-react",
      ],
    },
  },
  esbuild: {
    jsx: "automatic",
    jsxDev: false,
  },
});
