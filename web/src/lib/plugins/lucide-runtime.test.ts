import { readFileSync, readdirSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import * as Lucide from "./lucide-runtime";

const here = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(here, "../../../..");
const lucideShim = resolve(
  here,
  "../../../public/plugin-runtime/lucide-react.js"
);

function lucideExportNames(): string[] {
  return Object.keys(Lucide).filter((name) => name !== "default");
}

function namedImportsFromLucide(source: string): string[] {
  const names: string[] = [];
  const pattern =
    /import\s+(?:type\s+)?\{([^}]+)\}\s+from\s+["']lucide-react["']/g;
  for (const match of source.matchAll(pattern)) {
    for (const part of match[1].split(",")) {
      const name = part
        .trim()
        .split(/\s+as\s+/)[0]
        ?.trim();
      if (name) {
        names.push(name);
      }
    }
  }
  return names;
}

function pluginFrontendSources(): string[] {
  const pluginsDir = join(repoRoot, "plugins");
  const files: string[] = [];
  for (const plugin of readdirSync(pluginsDir)) {
    const srcDir = join(pluginsDir, plugin, "frontend", "src");
    try {
      for (const name of readdirSync(srcDir)) {
        if (name.endsWith(".ts") || name.endsWith(".tsx")) {
          files.push(join(srcDir, name));
        }
      }
    } catch {
      continue;
    }
  }
  return files;
}

describe("plugin lucide runtime", () => {
  it("exports warning icons used by accounting progress", () => {
    expect(Lucide.TriangleAlert).toBeDefined();
    expect(Lucide.CircleAlert).toBeDefined();
    expect(Lucide.Wallet).toBeDefined();
  });

  it("keeps the lucide shim in sync with the host runtime", () => {
    const shim = readFileSync(lucideShim, "utf8");
    for (const name of lucideExportNames()) {
      expect(shim).toContain(`export const ${name} = m.${name};`);
    }
  });

  it("provides every lucide icon imported by plugin frontends", () => {
    const available = new Set(lucideExportNames());
    for (const file of pluginFrontendSources()) {
      const source = readFileSync(file, "utf8");
      for (const name of namedImportsFromLucide(source)) {
        expect(available.has(name), `${file} imports ${name}`).toBe(true);
        expect((Lucide as Record<string, unknown>)[name]).toBeDefined();
      }
    }
  });
});
